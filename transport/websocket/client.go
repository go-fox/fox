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
	"fmt"
	"log/slog"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/go-fox/sugar/container/smap"
	"github.com/google/uuid"

	"github.com/go-fox/fox/api/gen/go/protocol"
	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/proto"
	"github.com/go-fox/fox/middleware"

	"github.com/go-fox/fox/transport"
)

type callbackInfo struct {
	data []byte
	err  error
}

// Client ws client
type Client struct {
	conn       *websocket.Conn
	endpoint   string
	encoder    EncoderRequestFunc
	decoder    DecodeResponseFunc
	codec      codec.Codec
	callbacks  *smap.Map[string, chan callbackInfo]
	baseCtx    context.Context
	cancel     context.CancelFunc
	timeout    time.Duration
	middleware []middleware.Middleware
}

// NewClient create a websocket client
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	context, cancel := context.WithCancel(ctx)
	c := &Client{
		baseCtx:   context,
		cancel:    cancel,
		callbacks: smap.New[string, chan callbackInfo](true),
		timeout:   5 * time.Second,
		codec:     proto.Codec{},
		encoder:   DefaultRequestEncoder,
		decoder:   DefaultResponseDecoder,
	}
	for _, opt := range opts {
		opt(c)
	}

	conn, _, err := websocket.DefaultDialer.Dial(c.endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	c.readMessage()
	return c, nil
}

// Invoke makes a rpc call procedure for remote service.
func (c *Client) Invoke(ctx context.Context, operation string, args any, reply any, opts ...CallOption) error {
	var (
		cancel context.CancelFunc
		data   []byte
	)
	ctx, cancel = context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req := protocol.AcquireRequest()
	req.Operation = operation
	defer protocol.ReleaseRequest(req)
	for _, opt := range opts {
		if err := opt.before(req); err != nil {
			return err
		}
	}
	if args != nil {
		encoder, err := c.encoder(ctx, args)
		if err != nil {
			return err
		}
		data = encoder
	}
	req.Data = data
	ctx = transport.NewClientContext(ctx, &Transport{
		endpoint:  c.endpoint,
		operation: operation,
		req:       req,
	})
	return c.invoke(ctx, req, reply, opts...)
}

func (c *Client) invoke(ctx context.Context, req *protocol.Request, reply any, opts ...CallOption) error {
	h := func(ctx context.Context, in any) (any, error) {
		req.Id = uuid.NewString()
		body, err := c.codec.Marshal(req)
		if err != nil {
			return nil, err
		}
		if err := c.conn.WriteMessage(websocket.BinaryMessage, body); err != nil {
			return nil, err
		}
		channel := make(chan callbackInfo)
		defer close(channel)
		c.callbacks.Set(req.Id, channel)
		defer c.callbacks.Del(req.Id)
		select {
		case <-c.baseCtx.Done():
			return nil, c.baseCtx.Err()
		case <-ctx.Done():
			return nil, ctx.Err()
		case info := <-channel:
			if info.err != nil {
				return nil, info.err
			}
			if err := c.decoder(ctx, info.data, reply); err != nil {
				return nil, err
			}
		}
		return reply, nil
	}
	if len(c.middleware) > 0 {
		h = middleware.Chain(c.middleware...)(h)
	}
	_, err := h(ctx, req)
	return err
}

func (c *Client) readMessage() {
	go func() {
		for {
			select {
			case <-c.baseCtx.Done():
				return
			default:
				messageType, data, err := c.conn.ReadMessage()
				if err != nil {
					return
				}
				if messageType == websocket.BinaryMessage {
					c.handlerReply(data)
				}
			}
		}
	}()
}

func (c *Client) handlerReply(data []byte) {
	reply := protocol.AcquireReply()
	defer protocol.ReleaseReply(reply)
	err := c.codec.Unmarshal(data, reply)
	if err != nil {
		slog.Error("proto unmarshal fail", slog.Any("error", err))
		return
	}
	v, ok := c.callbacks.Get(reply.Id)
	if !ok {
		slog.Error(fmt.Sprintf("not found request id %s", reply.Id))
		return
	}
	v <- callbackInfo{
		err:  nil,
		data: reply.Data,
	}
}

// Close is websocket conn close func.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
