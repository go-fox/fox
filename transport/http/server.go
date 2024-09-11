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
	"net/http"
	"net/url"
	"sync"

	"github.com/go-fox/sugar/util/shost"
	"github.com/go-fox/sugar/util/surl"
	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/errors"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/transport"
)

var (
	_ transport.Server = (*Server)(nil)
)

const (
	// SupportPackageIsVersion1 http version
	SupportPackageIsVersion1 = true
)

// Server is HTTP server
type Server struct {
	baseCtx       context.Context
	endpoint      *url.URL
	fastSrv       *fasthttp.Server
	fsInstances   []*fsInstance
	fsInstanceMux sync.Mutex
	initOnce      sync.Once
	config        *ServerConfig
	*router
}

// NewServer create a server
func NewServer(opts ...ServerOption) *Server {
	conf := DefaultServerConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewWithConfig(conf)
}

// NewWithConfig create server with config
func NewWithConfig(cs ...*ServerConfig) *Server {
	c := DefaultServerConfig()
	if len(cs) > 0 {
		c = cs[0]
	}
	srv := &Server{
		baseCtx:       context.Background(),
		config:        c,
		fsInstanceMux: sync.Mutex{},
	}
	if c.KeyFile != "" && c.CertFile != "" && c.tlsConf == nil {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			panic(err)
		}
		c.tlsConf = &tls.Config{Certificates: []tls.Certificate{cert}}
	}
	srv.init()
	return srv
}

// init is init fast http
func (s *Server) init() {
	s.initOnce.Do(func() {
		conf := s.config
		s.router = &router{
			tree: &node{},
			srv:  s,
		}
		s.fastSrv = &fasthttp.Server{
			Logger: &logger{
				slog: s.config.logger,
			},
			Handler:                       s.ServeFastHTTP,
			ErrorHandler:                  s.fastHTTPErrorHandler,
			LogAllErrors:                  false,
			Name:                          "",
			Concurrency:                   conf.Concurrency,
			DisableHeaderNamesNormalizing: false,
			MaxRequestBodySize:            conf.MaxRequestBodySize,
			NoDefaultServerHeader:         true,
			ReadBufferSize:                conf.ReadBufferSize,
			WriteBufferSize:               conf.WriteBufferSize,
			ReduceMemoryUsage:             conf.ReduceMemoryUsage,
			StreamRequestBody:             conf.StreamRequestBody,
		}
	})
}

// listenAndEndpoint listen and set endpoint
func (s *Server) listenAndEndpoint() error {
	if s.config.listener == nil {
		lis, err := net.Listen(s.config.Network, s.config.Address)
		if err != nil {
			return err
		}
		s.config.listener = lis
	}
	if s.endpoint == nil {
		addr, err := shost.Extract(s.config.Address, s.config.listener)
		if err != nil {
			return err
		}
		s.endpoint = surl.NewURL(surl.Scheme("http", s.config.tlsConf != nil), addr)
	}
	return nil
}

// Config return server config
func (s *Server) Config() ServerConfig {
	return *s.config
}

// Use registers a middleware route that will match requests
// with the provided prefix (which is optional and defaults to "/").
// Also, you can pass another app instance as a sub-router along a routing path.
// It's very useful to split up a large API as many independent routers and
// compose them as a single service using Use. The fiber's error handler and
// any of the fiber's sub apps are added to the application's error handlers
// to be invoked on errors that happen within the prefix route.
//
//		srv := http.New()
//		srv.Use(func(c http.Ctx) error {
//		     return c.Next()
//		})
//		srv.Use("/api", func(c http.Ctx) error {
//		     return c.Next()
//		})
//		srv.Use("/helloworld.v1.Greeter/*", func(handler middleware.Handler) middleware.Handler {
//			return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//				return handler(ctx, request)
//			}
//		})
//		srv.Use("/api", handler, func(c http.Ctx) error {
//		     return c.Next()
//		})
//	 	subRoute := http.NewServeMux()
//		srv.Use("/mounted-path", subRoute)
//
// This method will match all HTTP verbs: GET, POST, PUT, HEAD etc...
func (s *Server) Use(args ...any) Router {
	var (
		selector = ""
		mws      []middleware.Middleware
	)
	for _, arg := range args {
		switch arg := arg.(type) {
		case middleware.Middleware:
			mws = append(mws, arg)
		case string:
			selector = arg
		}
	}
	if len(mws) == 0 || selector == "" {
		return s.router.Use(args...)
	}
	s.config.middlewares.Add(selector, mws...)
	return s
}

// Walk walks the router and all its sub-routers, calling walkFn for each route in the tree.
func (s *Server) Walk(walkFunc WalkFunc) error {
	return Walk(s, walkFunc)
}

// Start is start this server
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	s.config.logger.Info(fmt.Sprintf("[HTTP] server listening on: %s", s.config.listener.Addr().String()))
	var err error
	listener := s.config.listener
	if s.config.tlsConf != nil {
		err = s.fastSrv.Serve(tls.NewListener(listener, s.config.tlsConf))
	} else {
		err = s.fastSrv.Serve(listener)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop is stop this server
func (s *Server) Stop(ctx context.Context) error {
	s.config.logger.Info("[HTTP] server stopping")
	return s.fastSrv.ShutdownWithContext(ctx)
}
