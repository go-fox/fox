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
	"testing"

	"github.com/go-fox/fox/transport"
)

func TestRouter(t *testing.T) {
	v2 := NewServer()
	v2.Use(func(ctx *Context) error {
		return ctx.Next()
	})
	{
		v2.Static("/files/*", FSRootPath("./static"))
		v2.Get("/download/{path}", func(ctx *Context) error {
			return ctx.SendFile("./files/excel.xlsx")
		})
		api := v2.Group("/api", func(ctx *Context) error {
			t.Logf("执行v2/api中间件")
			return ctx.Next()
		})
		category := api.Group("/category", func(ctx *Context) error {
			t.Logf("执行/v2/api/category的中间件")
			return ctx.Next()
		})
		{
			category.Get("/{id}", func(ctx *Context) error {
				return ctx.SendString("执行/v2/api/category/{id}的函数")
			})
		}
	}
	srv := NewServer(
		Address("0.0.0.0:8080"),
	)
	v1 := srv.Group("/v1", func(ctx *Context) error {
		tr, ok := transport.FromServerContext(ctx.Context())
		if !ok {
			t.Error("not found transport")
			return nil
		}
		t.Logf("执行/v1的中间件,%s", tr.Operation())
		return ctx.Next()
	})
	{
		api := v1.Group("/api", func(ctx *Context) error {
			t.Logf("执行/v1/api的中间件")
			return ctx.Next()
		})
		{
			category := api.Group("/category", func(ctx *Context) error {
				t.Logf("执行/v1/api/categrory的中间件")
				return ctx.Next()
			})
			{
				category.Get("/{id}", func(ctx *Context) error {
					return ctx.SendString("执行/v1/api/category/{id}的函数")
				})
			}
		}
	}
	srv.Mount("/v2", v2)
	if err := srv.Walk(func(method string, route string, handler Handler, middlewares ...Handler) error {
		println(method, route, handler, middlewares)
		return nil
	}); err != nil {
		t.Error(err)
	}
	if err := srv.Start(context.Background()); err != nil {
		t.Error(err)
	}
}
