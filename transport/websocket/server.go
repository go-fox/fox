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
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/fasthttp/websocket"
	"github.com/go-fox/sugar/container/smap"
	"github.com/go-fox/sugar/util/shost"
	"github.com/go-fox/sugar/util/surl"
	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/api/protocol"
	"github.com/go-fox/fox/errors"

	"github.com/panjf2000/ants/v2"

	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/transport"
)

var (
	_ transport.Server     = (*Server)(nil)
	_ transport.Endpointer = (*Server)(nil)
	// Name is name for transport
	Name transport.Kind = "websocket"
)

// Server is impl transport.Server
type Server struct {
	baseCtx     context.Context
	srv         *fasthttp.Server
	endpoint    *url.URL
	config      *ServerConfig
	sessionPool *ants.Pool
	handlerPool *ants.Pool
	handlerMap  *smap.Map[string, HandlerFunc]
}

// NewServer new a websocket server by options
func NewServer(opts ...ServerOption) *Server {
	c := DefaultServerConfig()
	for _, opt := range opts {
		opt(c)
	}
	return NewServerWithConfig(c)
}

// NewServerWithConfig create websocket with server config
func NewServerWithConfig(cs ...*ServerConfig) *Server {
	c := DefaultServerConfig()
	if len(cs) > 0 {
		c = cs[0]
	}
	srv := &Server{
		config:     c,
		handlerMap: smap.New[string, HandlerFunc](),
		srv: &fasthttp.Server{
			Name: "websocket",
		},
	}

	if c.KeyFile != "" && c.CertFile != "" && c.tlsConf == nil {
		cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			panic(err)
		}
		c.tlsConf = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	srv.srv.Handler = srv.serveWs

	// 初始化session处理协程池
	clientPool, err := ants.NewPool(c.SessionPoolSize)
	if err != nil {
		panic(err)
	}
	srv.sessionPool = clientPool

	// 初始化request处理协程池
	handlerPool, err := ants.NewPool(c.HandlerPoolSize)
	if err != nil {
		panic(err)
	}
	srv.handlerPool = handlerPool
	return srv
}

// Endpoint 获取地址
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

// Start is start this server
func (s *Server) Start(ctx context.Context) error {
	c := s.config
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	s.baseCtx = ctx
	c.logger.Info(fmt.Sprintf("[HTTP] server listening on: %s", c.lis.Addr().String()))
	var err error
	listener := c.lis
	if c.tlsConf != nil {
		err = s.srv.Serve(tls.NewListener(listener, c.tlsConf))
	} else {
		err = s.srv.Serve(listener)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return err
}

// Stop is stop this server
func (s *Server) Stop(ctx context.Context) error {
	s.config.logger.Info("[HTTP] server stopping")
	return s.srv.ShutdownWithContext(ctx)
}

// Use uses service middleware with selector.
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.config.middleware.Add(selector, m...)
}

// Handler 设置handler
func (s *Server) Handler(operator string, handler HandlerFunc) {
	s.handlerMap.Set(operator, handler)
}

// serveWs is server websocket
func (s *Server) serveWs(ctx *fasthttp.RequestCtx) {
	if err := s.config.upgrader.Upgrade(ctx, s.handlerConn); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
	}
}

// handlerConn conn process handler
func (s *Server) handlerConn(conn *websocket.Conn) {
	// 1.借用session
	ss := acquireSession(s.baseCtx, s.config.codec, conn)
	defer releaseSession(ss)

	// 2.处理客户端链接成功
	if err := s.onConnected(ss); err != nil {
		s.config.logger.Error("[WS] onConnected error", "error", errors.FromError(err).WithStack())
		_ = ss.Close()
		return
	}
	defer s.onDisconnected(ss)

	// 3.读取消息，并交给handler
	for {
		select {
		case <-s.baseCtx.Done():
			return
		default:
			if err := s.receiveRequest(ss); err != nil {
				if !errors.IsClientClosed(err) {
					s.config.logger.Error("[WS] receiveRequest error", "error", errors.FromError(err).WithStack())
					continue
				}
				return
			}
		}
	}
}

// onConnected client connected process handler
func (s *Server) onConnected(ss *Session) error {
	// 1.拦截器处理
	if s.config.connectedInterceptor != nil {
		if err := s.config.connectedInterceptor(ss); err != nil {
			return err
		}
	}

	// 2.认证处理
	if s.config.authorization != nil {
		if err := s.config.authorization(ss); err != nil {
			return err
		}
	}

	return nil
}

// onDisconnected client disconnected handler
func (s *Server) onDisconnected(ss *Session) {
	if s.config.disconnectedInterceptor != nil {
		s.config.disconnectedInterceptor(ss)
	}
}

// receiveRequest receive request
func (s *Server) receiveRequest(ss *Session) error {
	// 1.获取request
	req := protocol.AcquireRequest()
	if err := ss.Receive(req); err != nil {
		protocol.ReleaseRequest(req)
		return err
	}

	// 2.交给处理函数
	if err := s.handlerPool.Submit(func() {
		defer protocol.ReleaseRequest(req)
		if err := s.onReceiveRequest(ss, req); err != nil {
			s.config.ene(ss, req, err)
		}
	}); err != nil {
		protocol.ReleaseRequest(req)
		return err
	}
	return nil
}

// onReceiveRequest process
func (s *Server) onReceiveRequest(sess *Session, req *protocol.Request) error {
	v, ok := s.handlerMap.Get(req.Operation)
	if ok {
		reply := protocol.AcquireReply()
		defer protocol.ReleaseReply(reply)
		reply.Id = req.Id

		foxCtx := acquireContext(sess, s, req, reply)
		defer releaseContext(foxCtx)
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if s.config.Timeout > 0 {
			ctx, cancel = context.WithTimeout(foxCtx.Context, s.config.Timeout)
		} else {
			ctx, cancel = context.WithCancel(foxCtx.Context)
		}
		defer cancel()

		tr := &Transport{
			endpoint:  s.endpoint.String(),
			operation: req.Operation,
			ss:        sess,
			req:       req,
			reply:     reply,
		}
		if s.endpoint != nil {
			tr.endpoint = s.endpoint.String()
		}
		foxCtx.Context = transport.NewServerContext(ctx, tr)
		return v(foxCtx)
	}
	return errors.NotFound("OPERATION_NOT_FOUND", fmt.Sprintf("operation: %s not fond", req.Operation))
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
		s.endpoint = surl.NewURL(surl.Scheme("http", s.config.tlsConf != nil), addr)
	}
	return nil
}
