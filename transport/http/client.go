// Package http
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/errors"
	"github.com/go-fox/fox/internal/bytesconv"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/selector"
	"github.com/go-fox/fox/transport"
)

// Client http client
type Client struct {
	insecure bool
	config   *ClientConfig
	resolver *resolver
	target   *Target
	selector selector.Selector
	cc       *fasthttp.HostClient
}

// NewClient create client with option
func NewClient(opts ...ClientOption) *Client {
	conf := DefaultClientConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewClientWithConfig(conf)
}

// NewClientWithConfig create a client with config
func NewClientWithConfig(configs ...*ClientConfig) *Client {
	c := DefaultClientConfig()
	if len(configs) > 0 {
		c = configs[0]
	}
	if c.KeyFile != "" && c.CertFile != "" && c.tlsConf == nil {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			panic(err)
		}
		c.tlsConf = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	insecure := c.tlsConf == nil
	target, err := parseTarget(c.Endpoint, insecure)
	if err != nil {
		panic(err)
	}
	var resolver *resolver
	var se selector.Selector
	// 如果有复制均衡
	if c.discovery != nil {
		se = selector.Get(c.BalancerName).Build()
		resolver, err = newResolver(c.ctx, c.logger, c.discovery, target, se, c.Block, insecure)
	}
	return &Client{
		target:   target,
		insecure: insecure,
		resolver: resolver,
		selector: se,
		config:   c,
		cc: &fasthttp.HostClient{
			TLSConfig: c.tlsConf,
			IsTLS:     !insecure,
			Name:      c.UserAgent,
		},
	}
}

// Invoke ...
func (c *Client) Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) (err error) {
	var (
		contentType string
		body        []byte
	)
	req := reqPool.Get()
	resp := respPool.Get()
	info := defaultCallInfo(path)
	for _, o := range opts {
		if err = o(info, before, ctx); err != nil {
			return err
		}
	}
	// 设置header信息
	req.Header.Set("Content-Type", contentType)
	// 设置user-agent
	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}
	if args != nil {
		body, err = c.config.encodeRequest(ctx, req, args)
		if err != nil {
			return err
		}
		contentType = info.contentType
	}

	// 设置url
	url := fmt.Sprintf("%s://%s%s", c.target.Scheme, c.target.Authority, path)
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.SetBody(body)
	transport.NewClientContext(ctx, &Transport{
		endpoint:     c.config.Endpoint,
		request:      req,
		response:     resp,
		pathTemplate: info.pathTemplate,
		operation:    info.operation,
	})
	return c.invoke(ctx, req, resp, args, reply, info, opts...)
}

func (c *Client) invoke(ctx context.Context, req *Request, resp *Response, args interface{}, reply interface{}, info *callInfo, opts ...CallOption) error {
	h := func(ctx context.Context, in interface{}) (interface{}, error) {
		err := c.do(ctx, req, resp)
		if err != nil {
			return nil, err
		}
		for _, opt := range opts {
			err := opt(info, after, resp)
			if err != nil {
				return nil, err
			}
		}
		if err := c.config.decodeResponse(ctx, resp, reply); err != nil {
			return nil, err
		}
		return reply, nil
	}
	var p selector.Peer
	ctx = transport.NewPeerClient(ctx, &p)
	if len(c.config.middleware) > 0 {
		h = middleware.Chain(c.config.middleware...)(h)
	}
	_, err := h(ctx, args)
	return err
}

// Do 发送请求
func (c *Client) Do(ctx context.Context, req *Request, resp *Response, opts ...CallOption) error {
	info := defaultCallInfo(bytesconv.BytesToString(req.URI().Path()))
	for _, opt := range opts {
		if err := opt(info, before, req); err != nil {
			return err
		}
	}
	return c.do(ctx, req, resp)
}

func (c *Client) do(ctx context.Context, req *Request, resp *Response) error {
	var done func(context.Context, selector.DoneInfo)
	if c.resolver != nil {
		var (
			err  error
			node selector.Node
		)
		if node, done, err = c.selector.Select(ctx, selector.WithNodeFilter(c.config.nodeFilters...)); err != nil {
			return errors.ServiceUnavailable("NODE_NOT_FOUND", err.Error())
		}
		if c.insecure {
			req.URI().SetScheme("http")
		} else {
			req.URI().SetScheme("https")
		}
		req.URI().SetHost(node.Address())
		req.SetHost(node.Address())
	}
	c.cc.Addr = addMissingPort(bytesconv.BytesToString(req.URI().Host()), !c.insecure)
	err := c.cc.Do(req, resp)
	if err == nil {
		err = c.config.errorDecoder(ctx, resp)
	}
	if done != nil {
		done(ctx, selector.DoneInfo{Err: err})
	}
	if err != nil {
		return err
	}
	return nil
}

func addMissingPort(addr string, isTLS bool) string {
	n := strings.Index(addr, ":")
	if n >= 0 {
		return addr
	}
	port := 80
	if isTLS {
		port = 443
	}
	return net.JoinHostPort(addr, strconv.Itoa(port))
}
