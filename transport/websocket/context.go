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

	"github.com/go-fox/sugar/container/spool"

	"github.com/go-fox/fox/api/protocol"
	"github.com/go-fox/fox/middleware"
)

var ctxPool = spool.New[*wrappedContext](func() *wrappedContext {
	return &wrappedContext{}
}, func(c *wrappedContext) {
	c.ss = nil
	c.Context = nil
	c.srv = nil
	c.request = nil
	c.reply = nil
})

// Context ...context
type Context interface {
	context.Context
	Result(data any) error
	Bind(any) error
}

type wrappedContext struct {
	context.Context
	ss      *Session
	request *protocol.Request // 请求信息
	reply   *protocol.Reply   // 响应信息
	srv     *Server           // 服务
}

func acquireContext(ss *Session, srv *Server, r *protocol.Request, reply *protocol.Reply) *wrappedContext {
	ctx := ctxPool.Get()
	ctx.ss = ss
	ctx.Context = context.Background()
	ctx.srv = srv
	ctx.request = r
	ctx.reply = reply
	return ctx
}

func releaseContext(ctx *wrappedContext) {
	ctxPool.Put(ctx)
}

// Middleware run middleware
func (w *wrappedContext) Middleware(h middleware.Handler) middleware.Handler {
	return middleware.Chain(w.srv.ms...)(h)
}

// Result writes data to a client
func (w *wrappedContext) Result(data any) error {
	return w.srv.enc(w.ss, w.request, w.reply, data)
}

// Bind bind data
func (w *wrappedContext) Bind(data any) error {
	return w.srv.dec(w.request, data)
}
