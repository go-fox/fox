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
	"context"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/grpc"

	v1 "github.com/go-fox/fox/internal/testdata/helloword/v1"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/transport"
)

type testHelloWordService struct {
	v1.UnimplementedGreeterServiceServer
}

func (t *testHelloWordService) SayHi(ctx context.Context, req *v1.SayHiRequest) (*v1.SayHiResponse, error) {
	return &v1.SayHiResponse{
		Content: "Hello=" + req.GetName(),
	}, nil
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(
		Address("127.0.0.1:8888"),
		Middleware(
			func(handler middleware.Handler) middleware.Handler {
				return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
					if tr, ok := transport.FromServerContext(ctx); ok {
						if tr.ReplyHeader() != nil {
							tr.ReplyHeader().Set("req_id", strconv.Itoa(rand.Int()))
						}
					}
					return handler(ctx, req)
				}
			}),
		UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return handler(ctx, req)
		}),
		Options(grpc.InitialConnWindowSize(0)),
	)
	v1.RegisterGreeterServiceServer(srv, &testHelloWordService{})

	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	} else {
		println("endpoint：" + e.Host)
	}

	// start server
	if err := srv.Start(ctx); err != nil {
		panic(err)
	}
	defer srv.Stop(ctx)
}
