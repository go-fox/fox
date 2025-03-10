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
	"net"
	"strings"

	"github.com/go-fox/fox/api/gen/go/protocol"
	"github.com/go-fox/fox/transport"
)

const KindWebsocket transport.Kind = "websocket" // KindWebsocket websocket kind

var _ Transporter = (*Transport)(nil)

// Transporter is a websocket transport
type Transporter interface {
	transport.Transporter
}

// Transport is a websocket transport.
type Transport struct {
	endpoint  string
	operation string
	ss        *Session
	req       *protocol.Request
	reply     *protocol.Reply
}

func (t *Transport) RemoteAddr() net.Addr {
	return t.ss.conn.RemoteAddr()
}

// Kind returns the transport kind.
func (t *Transport) Kind() transport.Kind {
	return KindWebsocket
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
	return headerCarrier(t.req.Metadata)
}

// ReplyHeader returns the reply header.
func (t *Transport) ReplyHeader() transport.Header {
	return headerCarrier(t.reply.Metadata)
}

var _ transport.Header = (*headerCarrier)(nil)

type headerCarrier map[string]string

func (h headerCarrier) Get(key string) string {
	s, ok := h[key]
	if !ok {
		return ""
	}
	return s
}

func (h headerCarrier) Set(key string, value string) {
	h[key] = value
}

func (h headerCarrier) Add(key string, value string) {
	oldValue := h[key]
	newValue := ""
	if oldValue != "" {
		newValue = oldValue + ";" + value
	} else {
		newValue = value
	}
	h[key] = newValue
}

func (h headerCarrier) Keys() []string {
	keys := make([]string, len(h))
	for key := range h {
		keys = append(keys, key)
	}
	return keys
}

func (h headerCarrier) Values(key string) []string {
	valStr := h.Get(key)
	return strings.Split(valStr, ";")
}
