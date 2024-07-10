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
	"net"
	"net/url"
	"time"

	"google.golang.org/grpc"

	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/selector"
)

// ServerOption is creating a server option
type ServerOption func(server *Server)

// Network is with server network option
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address is with server address option
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Endpoint is with server endpoint option
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// CustomHealth is with server health check server
func CustomHealth(customHealth bool) ServerOption {
	return func(s *Server) {
		s.customHealth = customHealth
	}
}

// Timeout is with server timeout option
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Middleware is with server middleware option
func Middleware(m ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middleware.Use(m...)
	}
}

// TLSConfig is with TLS config option
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConf = c
	}
}

// Listener is with server listener option
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.lis = lis
	}
}

// UnaryInterceptor returns a ServerOption that sets the UnaryServerInterceptor for the server.
func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInterceptors = in
	}
}

// StreamInterceptor returns a ServerOption that sets the StreamServerInterceptor for the server.
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInterceptors = in
	}
}

// Options with grpc options.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// ClientOption is creating client options
type ClientOption func(c *Client)

// WithEndpoint is with client endpoint option
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithDebug is with client debug option
func WithDebug(isDebug bool) ClientOption {
	return func(c *Client) {
		c.debug = isDebug
	}
}

// WithInsecure is with client insecure option
func WithInsecure(isInsecure bool) ClientOption {
	return func(c *Client) {
		c.insecure = isInsecure
	}
}

// WithTimeout is with client timeout option
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithMiddleware is with client middleware option
func WithMiddleware(m ...middleware.Middleware) ClientOption {
	return func(c *Client) {
		c.middleware = m
	}
}

// WithDiscovery is with client discovery
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(c *Client) {
		c.discovery = d
	}
}

// WithTLSConfig is with client TLS
func WithTLSConfig(tls *tls.Config) ClientOption {
	return func(c *Client) {
		c.tlsConf = tls
	}
}

// WithUnaryInterceptor is with client grpc.UnaryClientInterceptor
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(c *Client) {
		c.unaryInterceptors = in
	}
}

// WithStreamInterceptor is with client grpc.StreamInterceptor
func WithStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(c *Client) {
		c.streamInterceptors = in
	}
}

// WithGRPCOptions is with grpc.DialOption
func WithGRPCOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *Client) {
		c.grpcOpts = opts
	}
}

// WithNodeFilter is with client node filter
func WithNodeFilter(filters ...selector.NodeFilter) ClientOption {
	return func(c *Client) {
		c.filters = filters
	}
}

// WithHealthCheck is with client health check
func WithHealthCheck(healthCheck bool) ClientOption {
	return func(c *Client) {
		if !healthCheck {
			c.healthCheckConfig = ""
		}
	}
}
