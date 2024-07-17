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
	"time"

	"github.com/fasthttp/websocket"

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/middleware"
)

// AuthorizationHandler is a session authorization handler
type AuthorizationHandler func(ss *Session) error

// ConnectedInterceptor client connect interceptor
type ConnectedInterceptor func(ss *Session) error

// DisconnectedInterceptor client disconnect interceptor
type DisconnectedInterceptor func(ss *Session)

// ServerOption is websocket server options
type ServerOption func(s *Server)

// TLSConfig with tls config
func TLSConfig(config *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConf = config
	}
}

// Codec with a codec option
func Codec(codec codec.Codec) ServerOption {
	return func(s *Server) {
		s.codec = codec
	}
}

// Authorization with an authorization option.
func Authorization(handler AuthorizationHandler) ServerOption {
	return func(s *Server) {
		s.authorization = handler
	}
}

// OnConnected with a connected callback option
func OnConnected(onConnect ConnectedInterceptor) ServerOption {
	return func(s *Server) {
		s.connectedInterceptor = onConnect
	}
}

// OnDisconnected with a disconnect callback option
func OnDisconnected(disconnect DisconnectedInterceptor) ServerOption {
	return func(s *Server) {
		s.disconnectedInterceptor = disconnect
	}
}

// Network with a network option
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with an address option
func Address(address string) ServerOption {
	return func(s *Server) {
		s.address = address
	}
}

// Middleware with middlewares
func Middleware(ms ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware.Use(ms...)
	}
}

// SessionPoolSize with session pool size
func SessionPoolSize(size int) ServerOption {
	return func(s *Server) {
		s.sessionPoolSize = size
	}
}

// Timeout with timeout option
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// HandlerPoolSize with handler pool size
func HandlerPoolSize(size int) ServerOption {
	return func(s *Server) {
		s.handlerPoolSize = size
	}
}

// Upgrader with websocket.FastHTTPUpgrader option
func Upgrader(upgrader websocket.FastHTTPUpgrader) ServerOption {
	return func(s *Server) {
		s.upgrader = upgrader
	}
}

// Logger with logger option
func Logger(logger *slog.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}
