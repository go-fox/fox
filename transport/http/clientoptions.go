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
	"log/slog"
	"time"

	"github.com/go-fox/fox/config"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/selector"
	"github.com/go-fox/fox/selector/balancer/wrr"
)

// ClientConfig client config
type ClientConfig struct {
	Debug          bool                    `json:"debug"`      // 是否开启调试模式，默认值为：false
	Endpoint       string                  `json:"endpoint"`   // 请求地址：默认值为：""
	Block          bool                    `json:"block"`      // 是否阻塞调用
	UserAgent      string                  `json:"user_agent"` // user-agent 请求头，默认：""
	Timeout        time.Duration           `json:"timeout"`    // 请求超时时间，默认值：2s
	KeyFile        string                  `json:"key_file"`
	CertFile       string                  `json:"cert_file"`
	BalancerName   string                  `json:"balancer_name"`
	decodeResponse DecodeResponseFunc      // 响应信息解码器
	encodeRequest  EncodeRequestFunc       // 请求体编码器
	errorDecoder   DecodeErrorFunc         // 错误解码器
	middleware     []middleware.Middleware // 中间件
	nodeFilters    []selector.NodeFilter   // 节点过滤器
	discovery      registry.Discovery      // 服务发现
	tlsConf        *tls.Config
	ctx            context.Context
	logger         *slog.Logger
}

// ClientOption create client option
type ClientOption func(c *ClientConfig)

// DefaultClientConfig default client config
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Debug:          false,
		Timeout:        time.Second * 2,
		Block:          true,
		BalancerName:   wrr.Name,
		encodeRequest:  DefaultRequestEncoder,
		decodeResponse: DefaultResponseDecoder,
		errorDecoder:   DefaultErrorDecoder,
		ctx:            context.Background(),
		logger:         slog.Default().With("mod", ""),
	}
}

// RawClientConfig config.Scan() value to ClientConfig
func RawClientConfig(key string) *ClientConfig {
	conf := DefaultClientConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

// ScanClientConfig config.Scan() value to ClientConfig
func ScanClientConfig(names ...string) *ClientConfig {
	key := "application.transport.http.client"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawClientConfig(key)
}

// WithOptions apply config options
func (c *ClientConfig) WithOptions(opts ...ClientOption) *ClientConfig {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Build create client factory
func (c *ClientConfig) Build() *Client {
	return NewClientWithConfig(c)
}

// WithDebug with a client debug option
func WithDebug(debug bool) ClientOption {
	return func(c *ClientConfig) {
		c.Debug = debug
	}
}

// WithEndpoint with a client endpoint option
func WithEndpoint(endpoint string) ClientOption {
	return func(c *ClientConfig) {
		c.Endpoint = endpoint
	}
}

// WithBlock with a client block option
func WithBlock(block bool) ClientOption {
	return func(c *ClientConfig) {
		c.Block = block
	}
}

// WithUserAgent with a client userAgent option
func WithUserAgent(userAgent string) ClientOption {
	return func(c *ClientConfig) {
		c.UserAgent = userAgent
	}
}

// WithTimeout with a client timeout option
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

// WithKeyFile with a client KeyFile option
func WithKeyFile(keyFile string) ClientOption {
	return func(c *ClientConfig) {
		c.KeyFile = keyFile
	}
}

// WithCertFile with a client CertFile option
func WithCertFile(certFile string) ClientOption {
	return func(c *ClientConfig) {
		c.CertFile = certFile
	}
}

// WithBalancerName with a client balancerName option
func WithBalancerName(balancerName string) ClientOption {
	return func(c *ClientConfig) {
		c.BalancerName = balancerName
	}
}

// WithMiddleware with a client middlewares option
func WithMiddleware(mws ...middleware.Middleware) ClientOption {
	return func(c *ClientConfig) {
		c.middleware = mws
	}
}

// WithNodeFilters with a node httpHandlers option
func WithNodeFilters(nodeFilters ...selector.NodeFilter) ClientOption {
	return func(c *ClientConfig) {
		c.nodeFilters = nodeFilters
	}
}

// WithDiscovery with a discovery option
func WithDiscovery(discovery registry.Discovery) ClientOption {
	return func(c *ClientConfig) {
		c.discovery = discovery
	}
}

// WithTLSConfig with tls.Config option
func WithTLSConfig(tlsConf *tls.Config) ClientOption {
	return func(c *ClientConfig) {
		c.tlsConf = tlsConf
	}
}

// WithContext with a client context option
func WithContext(ctx context.Context) ClientOption {
	return func(c *ClientConfig) {
		c.ctx = ctx
	}
}

// WithLogger with client logger option
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *ClientConfig) {
		c.logger = logger
	}
}
