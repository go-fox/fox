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
	"time"

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/proto"
	"github.com/go-fox/fox/transport"
)

// EncoderRequestFunc is request encode func.
type EncoderRequestFunc func(ctx context.Context, args any) ([]byte, error)

// DecodeResponseFunc is response decode func.
type DecodeResponseFunc func(ctx context.Context, data []byte, out any) error

// DefaultRequestEncoder default request encoder
func DefaultRequestEncoder(ctx context.Context, args any) ([]byte, error) {
	encoding := codec.GetCodec(proto.Name)
	tr, ok := transport.FromClientContext(ctx)
	if ok {
		contextType := tr.RequestHeader().Get("Content-Type")
		if contextType != "" {
			en := codec.GetCodec(contextType)
			if en != nil {
				encoding = en
			}
		}
	}
	return encoding.Marshal(args)
}

// DefaultResponseDecoder default response decoder
func DefaultResponseDecoder(ctx context.Context, data []byte, out any) error {
	encoding := codec.GetCodec(proto.Name)
	tr, ok := transport.FromClientContext(ctx)
	if ok {
		contextType := tr.RequestHeader().Get("Content-Type")
		if contextType != "" {
			en := codec.GetCodec(contextType)
			if en != nil {
				encoding = en
			}
		}
	}
	return encoding.Unmarshal(data, out)
}

// ClientOption client option
type ClientOption func(c *Client)

// WithTimeout with client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithEndpoint with client endpoint
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithRequestEncoder with client request encoder.
func WithRequestEncoder(requestFunc EncoderRequestFunc) ClientOption {
	return func(c *Client) {
		c.encoder = requestFunc
	}
}

// WithResponseDecoder with client response decoder.
func WithResponseDecoder(decode DecodeResponseFunc) ClientOption {
	return func(c *Client) {
		c.decoder = decode
	}
}

// WithCodec with client protocol codec
func WithCodec(codec codec.Codec) ClientOption {
	return func(c *Client) {
		c.codec = codec
	}
}
