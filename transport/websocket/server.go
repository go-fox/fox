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
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/go-fox/sugar/container/smap"
	"github.com/go-fox/sugar/util/shost"
	"github.com/go-fox/sugar/util/surl"
	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/api/protocol"
	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/proto"
	"github.com/go-fox/fox/errors"

	"github.com/panjf2000/ants/v2"

	"github.com/go-fox/fox/internal/matcher"
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
	srv                     *fasthttp.Server
	baseCtx                 context.Context
	tlsConf                 *tls.Config
	lis                     net.Listener
	err                     error
	network                 string
	address                 string
	endpoint                *url.URL
	middleware              matcher.Matcher
	handlerMap              *smap.Map[string, HandlerFunc]
	logger                  *slog.Logger
	upgrader                websocket.FastHTTPUpgrader
	sessionPoolSize         int
	handlerPoolSize         int
	timeout                 time.Duration
	sessionPool             *ants.Pool
	handlerPool             *ants.Pool
	codec                   codec.Codec
	ms                      []middleware.Middleware
	authorization           AuthorizationHandler
	connectedInterceptor    ConnectedInterceptor
	disconnectedInterceptor DisconnectedInterceptor
	dec                     DecoderRequestFunc
	enc                     EncoderResponseFunc
	ene                     EncoderErrorFunc
}

// NewServer new a websocket server by options
func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":0",
		srv: &fasthttp.Server{
			Name: "websocket",
		},
		sessionPoolSize: math.MaxInt32,
		handlerPoolSize: math.MaxInt32,
		logger:          slog.Default(),
		codec:           proto.Codec{},
		handlerMap:      smap.New[string, HandlerFunc](),
		dec:             DefaultRequestDecoder,
		enc:             DefaultResponseEncoder,
		ene:             DefaultErrorEncoder,
		upgrader: websocket.FastHTTPUpgrader{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
		},
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.logger = srv.logger.With("transport.websocket")

	srv.srv.Handler = srv.serveWs

	// 初始化session处理协程池
	clientPool, err := ants.NewPool(srv.sessionPoolSize)
	if err != nil {
		panic(err)
	}
	srv.sessionPool = clientPool
	// 初始化request处理协程池
	handlerPool, err := ants.NewPool(srv.handlerPoolSize)
	if err != nil {
		panic(err)
	}
	srv.handlerPool = handlerPool
	return srv
}

// Endpoint 获取地址
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, s.err
	}
	return s.endpoint, nil
}

// Start is start this server
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	s.baseCtx = ctx
	s.logger.Info(fmt.Sprintf("[HTTP] server listening on: %s", s.lis.Addr().String()))
	var err error
	listener := s.lis
	if s.tlsConf != nil {
		err = s.srv.Serve(tls.NewListener(listener, s.tlsConf))
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
	s.logger.Info("[HTTP] server stopping")
	return s.srv.ShutdownWithContext(ctx)
}

// Use uses service middleware with selector.
func (s *Server) Use(selector string, m ...middleware.Middleware) {
	s.middleware.Add(selector, m...)
}

// Handler 设置handler
func (s *Server) Handler(operator string, handler HandlerFunc) {
	s.handlerMap.Set(operator, handler)
}

// serveWs is server websocket
func (s *Server) serveWs(ctx *fasthttp.RequestCtx) {
	if err := s.upgrader.Upgrade(ctx, s.handlerConn); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusServiceUnavailable)
	}
}

// handlerConn conn process handler
func (s *Server) handlerConn(conn *websocket.Conn) {
	// 1.借用session
	ss := acquireSession(s.baseCtx, s.codec, conn)
	defer releaseSession(ss)

	// 2.处理客户端链接成功
	if err := s.onConnected(ss); err != nil {
		s.logger.Error("[WS] onConnected error", "error", errors.FromError(err).WithStack())
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
					s.logger.Error("[WS] receiveRequest error", "error", errors.FromError(err).WithStack())
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
	if s.connectedInterceptor != nil {
		if err := s.connectedInterceptor(ss); err != nil {
			return err
		}
	}

	// 2.认证处理
	if s.authorization != nil {
		if err := s.authorization(ss); err != nil {
			return err
		}
	}

	return nil
}

// onDisconnected client disconnected handler
func (s *Server) onDisconnected(ss *Session) {
	if s.disconnectedInterceptor != nil {
		s.disconnectedInterceptor(ss)
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
			s.ene(ss, req, err)
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
		if s.timeout > 0 {
			ctx, cancel = context.WithTimeout(foxCtx.Context, s.timeout)
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
		s.endpoint = surl.NewURL(surl.Scheme("http", s.tlsConf != nil), addr)
	}
	return nil
}
