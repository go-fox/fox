// Package grpc
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
package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	grpcmd "google.golang.org/grpc/metadata"

	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/selector"
	"github.com/go-fox/fox/transport"
	"github.com/go-fox/fox/transport/grpc/resolver/discovery"
)

func init() {
	all := selector.GetAll()
	all.Iterator(func(key string, builder selector.Builder) bool {
		old := balancer.Get(builder.Name())
		if old == nil {
			balancer.Register(newBalancerBuilder(builder))
		}
		return true
	})
}

// Client is grpc client
type Client struct {
	config *ClientConfig
	*grpc.ClientConn
}

// NewClient create a grpc client
func NewClient(opts ...ClientOption) *Client {
	c := DefaultClientConfig()
	for _, opt := range opts {
		opt(c)
	}
	return NewClientWithConfig(c)
}

// NewClientWithConfig create a client with client config
func NewClientWithConfig(cs ...*ClientConfig) *Client {
	c := DefaultClientConfig()
	if len(cs) > 0 {
		c = cs[0]
	}
	client := &Client{
		config: c,
	}
	if c.KeyFile != "" && c.CertFile != "" && c.tlsConf == nil {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			panic(err)
		}
		c.tlsConf = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	unaryInterceptors := []grpc.UnaryClientInterceptor{
		client.unaryClientInterceptor(c.middleware, c.Timeout, c.filters),
	}
	streamInterceptors := []grpc.StreamClientInterceptor{
		client.streamClientInterceptor(c.filters),
	}

	if len(c.unaryInterceptors) > 0 {
		unaryInterceptors = append(unaryInterceptors, c.unaryInterceptors...)
	}
	if len(c.streamInterceptors) > 0 {
		streamInterceptors = append(streamInterceptors, c.streamInterceptors...)
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]%s}`,
			c.BalancerName, c.healthCheckConfig)),
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	}

	if c.discovery != nil {
		grpcOpts = append(grpcOpts,
			grpc.WithResolvers(
				discovery.NewBuilder(
					c.discovery,
					discovery.WithDebug(c.Debug),
					discovery.WithInsecure(c.Insecure),
				)))
	}
	if c.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if c.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConf)))
	}

	if len(c.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, c.grpcOpts...)
	}
	conn, err := grpc.NewClient(c.Endpoint, grpcOpts...)
	if err != nil {
		panic(err)
	}

	client.ClientConn = conn
	return client
}

func (c *Client) unaryClientInterceptor(ms []middleware.Middleware, timeout time.Duration, filters []selector.NodeFilter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:    cc.Target(),
			operation:   method,
			reqHeader:   headerCarrier{},
			nodeFilters: filters,
		})
		if timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				header := tr.RequestHeader()
				keys := header.Keys()
				keyvals := make([]string, 0, len(keys))
				for _, k := range keys {
					keyvals = append(keyvals, k, header.Get(k))
				}
				ctx = grpcmd.AppendToOutgoingContext(ctx, keyvals...)
			}
			return reply, invoker(ctx, method, req, reply, cc, opts...)
		}
		if len(ms) > 0 {
			h = middleware.Chain(ms...)(h)
		}
		var p selector.Peer
		ctx = selector.NewPeerContext(ctx, &p)
		_, err := h(ctx, req)
		return err
	}
}

func (c *Client) streamClientInterceptor(filters []selector.NodeFilter) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) { // nolint
		ctx = transport.NewClientContext(ctx, &Transport{
			endpoint:    cc.Target(),
			operation:   method,
			reqHeader:   headerCarrier{},
			nodeFilters: filters,
		})
		var p selector.Peer
		ctx = selector.NewPeerContext(ctx, &p)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
