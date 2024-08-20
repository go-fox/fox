// Package grpc
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox copy from https://github.com/go-kratos/kratos.git
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
	"net"
	"net/url"

	"github.com/go-fox/sugar/util/shost"
	"github.com/go-fox/sugar/util/surl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/go-fox/fox/internal/matcher"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/transport"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
)

// Server Grpc服务
type Server struct {
	*grpc.Server
	endpoint   *url.URL
	baseCtx    context.Context
	config     *ServerConfig
	health     *health.Server
	middleware matcher.Matcher
	adminClean func()
}

// NewServer create server
func NewServer(opts ...ServerOption) *Server {
	conf := DefaultSeverConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewServerWithConfig(conf)
}

// NewServerWithConfig create server with config
func NewServerWithConfig(cs ...*ServerConfig) *Server {
	conf := DefaultSeverConfig()
	if len(cs) > 0 {
		conf = cs[0]
	}
	srv := &Server{
		baseCtx:    context.Background(),
		config:     conf,
		health:     health.NewServer(),
		middleware: matcher.New(),
	}
	if conf.KeyFile != "" && conf.CertFile != "" && conf.tlsConf == nil {
		var err error
		conf.tlsConf = new(tls.Config)
		conf.tlsConf.NextProtos = append(conf.tlsConf.NextProtos, "http/1.1")
		conf.tlsConf.Certificates = make([]tls.Certificate, 1)
		conf.tlsConf.Certificates[0], err = tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
		if err != nil {
			panic(err)
		}
	}
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		srv.unaryServerInterceptor(),
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		srv.streamServerInterceptor(),
	}
	if len(conf.unaryInterceptors) > 0 {
		unaryInterceptors = append(unaryInterceptors, conf.unaryInterceptors...)
	}
	if len(conf.streamInterceptors) > 0 {
		streamInterceptors = append(streamInterceptors, conf.streamInterceptors...)
	}
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}
	if conf.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(conf.tlsConf)))
	}
	if len(conf.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, conf.grpcOpts...)
	}
	srv.Server = grpc.NewServer(grpcOpts...)
	srv.adminClean, _ = admin.Register(srv.Server)
	// health check
	if !conf.CustomHealth {
		grpc_health_v1.RegisterHealthServer(srv.Server, srv.health)
	}
	return srv
}

// Endpoint 获取服务器启动地址
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

// Use uses service middleware with selector.
// selector:
//   - '/*'
//   - '/helloworld.v1.Greeter/*'
//   - '/helloworld.v1.Greeter/SayHello'
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

// Start 启动
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	s.baseCtx = ctx
	s.health.Resume()
	s.config.log.Info(fmt.Sprintf("[gRPC] server listening on: %s", s.config.lis.Addr().String()))
	return s.Serve(s.config.lis)
}

// Stop 停止
func (s *Server) Stop(ctx context.Context) error {
	if s.adminClean != nil {
		s.adminClean()
	}
	s.GracefulStop()
	s.health.Shutdown()
	s.config.log.Info("[gRPC] server stopping")
	return nil
}

// listenAndEndpoint is get server start address
func (s *Server) listenAndEndpoint() error {
	if s.config.lis == nil {
		lis, err := net.Listen(s.config.Network, s.config.Address)
		if err != nil {
			return err
		}
		s.config.lis = lis
	}
	if s.endpoint == nil {
		addr, err := shost.Extract(s.config.Address, s.config.lis)
		if err != nil {
			return err
		}
		s.endpoint = surl.NewURL(surl.Scheme("grpc", s.config.tlsConf != nil), addr)
	}
	return nil
}
