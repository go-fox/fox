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
	"net"

	"github.com/go-fox/fox/internal/bytesconv"
	"github.com/go-fox/fox/transport"
)

// KindHttp http kind
const KindHttp transport.Kind = "http"

var _ transport.Transporter = (*Transport)(nil)

// Transport is an HTTP transport.
type Transport struct {
	endpoint     string    // 入口地址
	operation    string    // 操作地址
	pathTemplate string    // 地址模板
	original     string    // 原始地址
	request      *Request  // 请求信息
	response     *Response // 响应信息
	remoteAddr   net.Addr  // 远程连接地址
}

func (t *Transport) RemoteAddr() net.Addr {
	return t.remoteAddr
}

// Kind returns the transport kind.
func (t *Transport) Kind() transport.Kind {
	return KindHttp
}

// Endpoint returns the transport endpoint.
func (t *Transport) Endpoint() string {
	return t.endpoint
}

// Operation returns the transport operation.
func (t *Transport) Operation() string {
	return t.operation
}

// RequestHeader returns the request header.
func (t *Transport) RequestHeader() transport.Header {
	return newHeaderCarrier(&t.request.Header)
}

// ReplyHeader returns the reply header.
func (t *Transport) ReplyHeader() transport.Header {
	return newHeaderCarrier(&t.response.Header)
}

type header interface {
	Add(key string, value string)
	Peek(key string) []byte
	PeekAll(key string) [][]byte
	PeekKeys() [][]byte
	Set(key, value string)
}

type headerCarrier struct {
	hd header
}

func newHeaderCarrier(hd header) *headerCarrier {
	return &headerCarrier{
		hd: hd,
	}
}

func (h *headerCarrier) Add(key string, value string) {
	h.hd.Add(key, value)
}

func (h *headerCarrier) Get(key string) string {
	return bytesconv.BytesToString(h.hd.Peek(key))
}

func (h *headerCarrier) Values(key string) []string {
	all := h.hd.PeekAll(key)
	result := make([]string, len(all))
	for i, v := range all {
		result[i] = bytesconv.BytesToString(v)
	}
	return result
}

func (h *headerCarrier) Set(key, value string) {
	h.hd.Set(key, value)
}

func (h *headerCarrier) Keys() []string {
	keys := h.hd.PeekKeys()
	res := make([]string, 0, len(keys))
	for _, key := range keys {
		res = append(res, bytesconv.BytesToString(key))
	}
	return res
}

// SetOperation set transport operation
func SetOperation(ctx context.Context, op string) {
	tr, ok := transport.FromServerContext(ctx)
	if ok {
		tr, ok := tr.(*Transport)
		if ok {
			tr.operation = op
		}
	}
}
