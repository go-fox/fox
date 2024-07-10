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
	"log/slog"
	"net"
	"net/url"
	"time"

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
	baseCtx            context.Context
	tlsConf            *tls.Config
	lis                net.Listener
	err                error
	network            string
	address            string
	endpoint           *url.URL
	timeout            time.Duration
	middleware         matcher.Matcher
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	grpcOpts           []grpc.ServerOption
	health             *health.Server
	customHealth       bool
	log                *slog.Logger
	adminClean         func()
}

// NewServer 构造函数
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		baseCtx:    context.Background(),
		network:    "tcp",
		address:    ":0",
		timeout:    1 * time.Second,
		health:     health.NewServer(),
		middleware: matcher.New(),
		log:        slog.With(slog.String("mod", "grpc.sever")),
	}
	for _, opt := range opts {
		opt(srv)
	}
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		srv.unaryServerInterceptor(),
	}
	streamInterceptors := []grpc.StreamServerInterceptor{
		srv.streamServerInterceptor(),
	}
	if len(srv.unaryInterceptors) > 0 {
		unaryInterceptors = append(unaryInterceptors, srv.unaryInterceptors...)
	}
	if len(srv.streamInterceptors) > 0 {
		streamInterceptors = append(streamInterceptors, srv.streamInterceptors...)
	}
	grpcOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}
	if srv.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(srv.tlsConf)))
	}
	if len(srv.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, srv.grpcOpts...)
	}
	srv.Server = grpc.NewServer(grpcOpts...)
	srv.adminClean, _ = admin.Register(srv.Server)
	// health check
	if !srv.customHealth {
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
	s.log.Info(fmt.Sprintf("[gRPC] server listening on: %s", s.lis.Addr().String()))
	return s.Serve(s.lis)
}

// Stop 停止
func (s *Server) Stop(ctx context.Context) error {
	if s.adminClean != nil {
		s.adminClean()
	}
	s.GracefulStop()
	s.health.Shutdown()
	s.log.Info("[gRPC] server stopping")
	return nil
}

// listenAndEndpoint is get server start address
func (s *Server) listenAndEndpoint() error {
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := shost.Extract(s.address, s.lis)
		if err != nil {
			return err
		}
		s.endpoint = surl.NewURL(surl.Scheme("grpc", s.tlsConf != nil), addr)
	}
	return nil
}
