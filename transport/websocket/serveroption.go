// Package websocket
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
package websocket

import (
	"crypto/tls"
	"log/slog"
	"math"
	"net"
	"time"

	"github.com/fasthttp/websocket"

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/proto"
	"github.com/go-fox/fox/internal/matcher"
	"github.com/go-fox/fox/middleware"
)

// AuthorizationHandler is a session authorization handler
type AuthorizationHandler func(ss *Session) error

// ConnectedInterceptor client connect interceptor
type ConnectedInterceptor func(ss *Session) error

// DisconnectedInterceptor client disconnect interceptor
type DisconnectedInterceptor func(ss *Session)

// ServerConfig websocket server config
type ServerConfig struct {
	Network                 string        `json:"network"`
	Address                 string        `json:"address"`
	SessionPoolSize         int           `json:"session_pool_size"`
	HandlerPoolSize         int           `json:"handler_pool_size"`
	CertFile                string        `json:"cert_file"`
	KeyFile                 string        `json:"key_file"`
	Timeout                 time.Duration `json:"timeout"`
	tlsConf                 *tls.Config
	lis                     net.Listener
	middleware              matcher.Matcher
	logger                  *slog.Logger
	upgrader                websocket.FastHTTPUpgrader
	codec                   codec.Codec
	ms                      []middleware.Middleware
	authorization           AuthorizationHandler
	connectedInterceptor    ConnectedInterceptor
	disconnectedInterceptor DisconnectedInterceptor
	dec                     DecoderRequestFunc
	enc                     EncoderResponseFunc
	ene                     EncoderErrorFunc
}

// ServerOption is websocket server options
type ServerOption func(s *ServerConfig)

// DefaultServerConfig default server option
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Network:         "tcp",
		Address:         ":0",
		SessionPoolSize: math.MaxInt32,
		HandlerPoolSize: math.MaxInt32,
		logger:          slog.With(slog.String("mod", "transport.websocket")),
		codec:           codec.GetCodec(proto.Name),
		dec:             DefaultRequestDecoder,
		enc:             DefaultResponseEncoder,
		ene:             DefaultErrorEncoder,
		upgrader: websocket.FastHTTPUpgrader{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		},
	}
}

// TLSConfig with tls config
func TLSConfig(config *tls.Config) ServerOption {
	return func(s *ServerConfig) {
		s.tlsConf = config
	}
}

// Codec with a codec option
func Codec(codec codec.Codec) ServerOption {
	return func(s *ServerConfig) {
		s.codec = codec
	}
}

// Authorization with an authorization option.
func Authorization(handler AuthorizationHandler) ServerOption {
	return func(s *ServerConfig) {
		s.authorization = handler
	}
}

// OnConnected with a connected callback option
func OnConnected(onConnect ConnectedInterceptor) ServerOption {
	return func(s *ServerConfig) {
		s.connectedInterceptor = onConnect
	}
}

// OnDisconnected with a disconnect callback option
func OnDisconnected(disconnect DisconnectedInterceptor) ServerOption {
	return func(s *ServerConfig) {
		s.disconnectedInterceptor = disconnect
	}
}

// Network with a network option
func Network(network string) ServerOption {
	return func(s *ServerConfig) {
		s.Network = network
	}
}

// Address with an address option
func Address(address string) ServerOption {
	return func(s *ServerConfig) {
		s.Address = address
	}
}

// Middleware with middlewares
func Middleware(ms ...middleware.Middleware) ServerOption {
	return func(s *ServerConfig) {
		s.middleware.Use(ms...)
	}
}

// AddMiddleware add middlewares
func AddMiddleware(selector string, mws ...middleware.Middleware) ServerOption {
	return func(s *ServerConfig) {
		s.middleware.Add(selector, mws...)
	}
}

// SessionPoolSize with session pool size
func SessionPoolSize(size int) ServerOption {
	return func(s *ServerConfig) {
		s.SessionPoolSize = size
	}
}

// Timeout with timeout option
func Timeout(timeout time.Duration) ServerOption {
	return func(s *ServerConfig) {
		s.Timeout = timeout
	}
}

// HandlerPoolSize with handler pool size
func HandlerPoolSize(size int) ServerOption {
	return func(s *ServerConfig) {
		s.HandlerPoolSize = size
	}
}

// Upgrader with websocket.FastHTTPUpgrader option
func Upgrader(upgrader websocket.FastHTTPUpgrader) ServerOption {
	return func(s *ServerConfig) {
		s.upgrader = upgrader
	}
}

// Logger with logger option
func Logger(logger *slog.Logger) ServerOption {
	return func(s *ServerConfig) {
		s.logger = logger
	}
}
