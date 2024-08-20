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
	"crypto/tls"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/go-fox/fox/config"
	"github.com/go-fox/fox/internal/matcher"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/selector"
	"github.com/go-fox/fox/selector/balancer/wrr"
)

// ServerConfig create server config
type ServerConfig struct {
	Network            string        `json:"network"`
	Address            string        `json:"address"`
	Timeout            time.Duration `json:"timeout"`
	CustomHealth       bool          `json:"custom_health"`
	CertFile           string        `json:"cert_file"`
	KeyFile            string        `json:"key_file"`
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	grpcOpts           []grpc.ServerOption
	middleware         matcher.Matcher
	tlsConf            *tls.Config
	lis                net.Listener
	log                *slog.Logger
}

// DefaultSeverConfig default config
func DefaultSeverConfig() *ServerConfig {
	return &ServerConfig{
		Network:    "tcp",
		Address:    "0.0.0.0:0",
		Timeout:    time.Second * 3,
		middleware: matcher.New(),
		log:        slog.With(slog.String("mod", "transport.grpc.sever")),
	}
}

// RawServerConfig scan config.Config value to ServerConfig
func RawServerConfig(key string) *ServerConfig {
	conf := DefaultSeverConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		return nil
	}
	return conf
}

// ScanServerConfig scan config.Config value to ServerConfig
func ScanServerConfig(names ...string) *ServerConfig {
	key := "application.transport.grpc.server"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawServerConfig(key)
}

// With apply options
func (s *ServerConfig) With(opts ...ServerOption) *ServerConfig {
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Build create server
func (s *ServerConfig) Build() *Server {
	return NewServerWithConfig(s)
}

// ServerOption is creating a server option
type ServerOption func(s *ServerConfig)

// Network is with server network option
func Network(network string) ServerOption {
	return func(s *ServerConfig) {
		s.Network = network
	}
}

// Address is with server address option
func Address(addr string) ServerOption {
	return func(s *ServerConfig) {
		s.Address = addr
	}
}

// CustomHealth is with server health check server
func CustomHealth(customHealth bool) ServerOption {
	return func(s *ServerConfig) {
		s.CustomHealth = customHealth
	}
}

// Timeout is with server timeout option
func Timeout(timeout time.Duration) ServerOption {
	return func(s *ServerConfig) {
		s.Timeout = timeout
	}
}

// Middleware is with server middleware option
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *ServerConfig) {
		s.middleware.Use(m...)
	}
}

// AddMiddleware add selector middleware
func AddMiddleware(selector string, mws ...middleware.Middleware) ServerOption {
	return func(s *ServerConfig) {
		s.middleware.Add(selector, mws...)
	}
}

// TLSConfig is with TLS config option
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *ServerConfig) {
		s.tlsConf = c
	}
}

// Listener is with server listener option
func Listener(lis net.Listener) ServerOption {
	return func(s *ServerConfig) {
		s.lis = lis
	}
}

// UnaryInterceptor returns a ServerOption that sets the UnaryServerInterceptor for the server.
func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *ServerConfig) {
		s.unaryInterceptors = in
	}
}

// StreamInterceptor returns a ServerOption that sets the StreamServerInterceptor for the server.
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *ServerConfig) {
		s.streamInterceptors = in
	}
}

// Options with grpc options.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *ServerConfig) {
		s.grpcOpts = opts
	}
}

// ClientConfig client config
type ClientConfig struct {
	Endpoint           string        `json:"endpoint"`
	Timeout            time.Duration `json:"timeout"`
	BalancerName       string        `json:"balancer_name"`
	Insecure           bool          `json:"insecure"`
	Debug              bool          `json:"debug"`
	CertFile           string        `json:"cert_file"`
	KeyFile            string        `json:"key_file"`
	healthCheckConfig  string
	filters            []selector.NodeFilter
	tlsConf            *tls.Config
	discovery          registry.Discovery
	middleware         []middleware.Middleware
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
	grpcOpts           []grpc.DialOption
}

// DefaultClientConfig default config
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:           2000 * time.Millisecond,
		BalancerName:      wrr.Name,
		Insecure:          true,
		healthCheckConfig: `,"healthCheckConfig":{"serviceName":""}`,
	}
}

// RawClientConfig scan config.Config value to ClientConfig
func RawClientConfig(key string) *ClientConfig {
	conf := DefaultClientConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		return nil
	}
	return conf
}

// ScanClientConfig scan config.Config value to ServerConfig
func ScanClientConfig(names ...string) *ClientConfig {
	key := "application.transport.grpc.client"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawClientConfig(key)
}

// With apply options
func (c *ClientConfig) With(opts ...ClientOption) *ClientConfig {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Build create grpc client
func (c *ClientConfig) Build() *Client {
	return NewClientWithConfig(c)
}

// ClientOption is creating client options
type ClientOption func(c *ClientConfig)

// WithEndpoint is with client endpoint option
func WithEndpoint(endpoint string) ClientOption {
	return func(c *ClientConfig) {
		c.Endpoint = endpoint
	}
}

// WithDebug is with client debug option
func WithDebug(isDebug bool) ClientOption {
	return func(c *ClientConfig) {
		c.Debug = isDebug
	}
}

// WithInsecure is with client insecure option
func WithInsecure(isInsecure bool) ClientOption {
	return func(c *ClientConfig) {
		c.Insecure = isInsecure
	}
}

// WithTimeout is with client timeout option
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

// WithMiddleware is with client middleware option
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(c *ClientConfig) {
		c.middleware = m
	}
}

// WithDiscovery is with client discovery
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(c *ClientConfig) {
		c.discovery = d
	}
}

// WithTLSConfig is with client TLS
func WithTLSConfig(tls *tls.Config) ClientOption {
	return func(c *ClientConfig) {
		c.tlsConf = tls
	}
}

// WithUnaryInterceptor is with client grpc.UnaryClientInterceptor
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(c *ClientConfig) {
		c.unaryInterceptors = in
	}
}

// WithStreamInterceptor is with client grpc.StreamInterceptor
func WithStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(c *ClientConfig) {
		c.streamInterceptors = in
	}
}

// WithGRPCOptions is with grpc.DialOption
func WithGRPCOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *ClientConfig) {
		c.grpcOpts = opts
	}
}

// WithNodeFilter is with client node filter
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(c *ClientConfig) {
		c.filters = filters
	}
}

// WithHealthCheck is with client health check
func WithHealthCheck(healthCheck bool) ClientOption {
	return func(c *ClientConfig) {
		if !healthCheck {
			c.healthCheckConfig = ""
		}
	}
}

// WithCertFile is with client tls.CertFile
func WithCertFile(certFile string) ClientOption {
	return func(c *ClientConfig) {
		c.CertFile = certFile
	}
}

// WithKeyFile is with client tls.KeyFile
func WithKeyFile(keyFile string) ClientOption {
	return func(c *ClientConfig) {
		c.KeyFile = keyFile
	}
}
