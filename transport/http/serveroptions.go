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
	"crypto/tls"
	"log/slog"
	"net"
	"time"

	"github.com/go-fox/fox/config"
	"github.com/go-fox/fox/internal/matcher"
	"github.com/go-fox/fox/middleware"
)

// ServerConfig constructors config
type ServerConfig struct {
	Network            string        `json:"network"`
	Address            string        `json:"address"`
	KeyFile            string        `json:"key_file"`
	CertFile           string        `json:"cert_file"`
	Timeout            time.Duration `json:"timeout"`
	Concurrency        int           `json:"concurrency"`
	MaxRequestBodySize int           `json:"max_request_body_size"`
	ReadBufferSize     int           `json:"read_buffer_size"`
	WriteBufferSize    int           `json:"write_buffer_size"`
	ReduceMemoryUsage  bool          `json:"reduce_memory_usage"`
	StreamRequestBody  bool          `json:"stream_request_body"`
	httpMiddlewares    []Handler     // http中间件
	listener           net.Listener
	tlsConf            *tls.Config
	ene                EncodeErrorFunc
	enc                EncodeResponseFunc
	decQuery           DecodeRequestFunc
	decVars            DecodeRequestVarsFunc
	decBody            DecodeRequestFunc
	middlewares        matcher.Matcher
	logger             *slog.Logger
}

// DefaultServerConfig default server options
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Network:         "tcp",
		Address:         "0.0.0.0:0",
		Timeout:         3 * time.Second,
		middlewares:     matcher.New(),
		ene:             DefaultErrorHandler,
		enc:             DefaultResponseHandler,
		decQuery:        DefaultDecodeRequestQuery,
		decVars:         DefaultDecodeRequestVars,
		decBody:         DefaultDecodeRequestBody,
		logger:          slog.Default().With(slog.String("mod", "transport.http.server")),
		httpMiddlewares: make([]Handler, 0),
	}
}

// ScanServerConfig scan config.Config value to ServerConfig
func ScanServerConfig(names ...string) *ServerConfig {
	key := "application.transport.http.server"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawServerConfig(key)
}

// RawServerConfig scan config.Config value to ServerConfig
func RawServerConfig(key string) *ServerConfig {
	conf := DefaultServerConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

// ServerOption server option
type ServerOption func(o *ServerConfig)

// Network http server network
func Network(network string) ServerOption {
	return func(o *ServerConfig) {
		o.Network = network
	}
}

// Address http address
func Address(address string) ServerOption {
	return func(o *ServerConfig) {
		o.Address = address
	}
}

// WithFilter with a http handler
func WithFilter(h ...Handler) ServerOption {
	return func(o *ServerConfig) {
		o.httpMiddlewares = h
	}
}

// Timeout http  request precess timeout
func Timeout(timeout time.Duration) ServerOption {
	return func(o *ServerConfig) {
		o.Timeout = timeout
	}
}

// Concurrency with a Concurrency option.
func Concurrency(concurrency int) ServerOption {
	return func(o *ServerConfig) {
		o.Concurrency = concurrency
	}
}

// MaxRequestBodySize with a MaxRequestBodySize option.
func MaxRequestBodySize(maxRequestBodySize int) ServerOption {
	return func(o *ServerConfig) {
		o.MaxRequestBodySize = maxRequestBodySize
	}
}

// ReadBufferSize with a ReadBufferSize option.
func ReadBufferSize(readBufferSize int) ServerOption {
	return func(o *ServerConfig) {
		o.ReadBufferSize = readBufferSize
	}
}

// WriteBufferSize with a WriteBufferSize option.
func WriteBufferSize(writeBufferSize int) ServerOption {
	return func(o *ServerConfig) {
		o.WriteBufferSize = writeBufferSize
	}
}

// ReduceMemoryUsage with a ReduceMemoryUsage option.
func ReduceMemoryUsage(reduceMemoryUsage bool) ServerOption {
	return func(o *ServerConfig) {
		o.ReduceMemoryUsage = reduceMemoryUsage
	}
}

// StreamRequestBody with a StreamRequestBody option.
func StreamRequestBody(streamRequestBody bool) ServerOption {
	return func(o *ServerConfig) {
		o.StreamRequestBody = streamRequestBody
	}
}

// Listener with a server lis option
func Listener(lis net.Listener) ServerOption {
	return func(o *ServerConfig) {
		o.listener = lis
	}
}

// TLSConfig with TLS config.
func TLSConfig(c *tls.Config) ServerOption {
	return func(o *ServerConfig) {
		o.tlsConf = c
	}
}

// ErrorEncoder with error encoder.
func ErrorEncoder(ene EncodeErrorFunc) ServerOption {
	return func(o *ServerConfig) {
		o.ene = ene
	}
}

// ResponseEncode response encoder
func ResponseEncode(enc EncodeResponseFunc) ServerOption {
	return func(o *ServerConfig) {
		o.enc = enc
	}
}

// Middleware with a service middleware option.
func Middleware(ms ...middleware.Middleware) ServerOption {
	return func(o *ServerConfig) {
		o.middlewares.Use(ms...)
	}
}

// UseMiddleware with a selector middlewares
func UseMiddleware(selector string, ms ...middleware.Middleware) ServerOption {
	return func(o *ServerConfig) {
		o.middlewares.Add(selector, ms...)
	}
}

// Logger with a *slog.Logger option.
func Logger(l *slog.Logger) ServerOption {
	return func(o *ServerConfig) {
		o.logger = l
	}
}

// WithOption reset config with options
func (s *ServerConfig) WithOption(opts ...ServerOption) *ServerConfig {
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Build is build server on configuration
func (s *ServerConfig) Build() *Server {
	return NewWithConfig(s)
}
