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
	"testing"

	"github.com/go-fox/sugar/util/surl"

	v1 "github.com/go-fox/fox/internal/testdata/helloword/v1"
	_ "github.com/go-fox/fox/selector/balancer/random"
)

func TestClient(t *testing.T) {
	url := surl.NewURL(surl.Scheme("grpc", false), "127.0.0.1:8888")
	ctx := context.Background()
	client := NewClient(
		WithEndpoint(url.Host),
	)
	serviceClient := v1.NewGreeterServiceClient(client)
	response, err := serviceClient.SayHi(ctx, &v1.SayHiRequest{
		Name: "测试数据",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("response: %v", response)
}
