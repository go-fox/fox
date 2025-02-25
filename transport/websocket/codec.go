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
	"github.com/go-fox/fox/api/gen/go/protocol"
	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/proto"
	"github.com/go-fox/fox/errors"
)

// EncoderResponseFunc is encode response func.
type EncoderResponseFunc func(ss *Session, r *protocol.Request, reply *protocol.Reply, v any) error

// DecoderRequestFunc is decoder request func.
type DecoderRequestFunc func(r *protocol.Request, v any) error

// EncoderErrorFunc is encode error func.
type EncoderErrorFunc func(ss *Session, r *protocol.Request, err error)

// DefaultResponseEncoder default response encoder
func DefaultResponseEncoder(ss *Session, r *protocol.Request, reply *protocol.Reply, v any) error {
	// 如果没有值
	if v == nil {
		return nil
	}
	encoding := CodecForRequest(r, "Content-Type")
	data, err := encoding.Marshal(v)
	if err != nil {
		return err
	}
	reply.Operation = r.Operation
	reply.Data = data
	reply.Metadata["Content-Type"] = encoding.Name()
	return ss.Send(reply)
}

// DefaultErrorEncoder default error encoder
func DefaultErrorEncoder(ss *Session, r *protocol.Request, err error) {
	fromError := errors.FromError(err)
	reply := protocol.AcquireReply()
	reply.Operation = r.Operation
	reply.Status = fromError.GRPCStatus().Proto()
	_ = ss.Send(reply)
}

// DefaultRequestDecoder default request decoder
func DefaultRequestDecoder(r *protocol.Request, v any) error {
	encoding := codec.GetCodec(proto.Name)
	contentType := headerCarrier(r.Metadata).Get("Content-Type")
	if len(contentType) > 0 {
		e := codec.GetCodec(contentType)
		if e != nil {
			encoding = e
		}
	}
	return encoding.Unmarshal(r.Data, v)
}

// CodecForRequest get encoding.Codec via http.Request
func CodecForRequest(r *protocol.Request, name string) codec.Codec {
	encoding := codec.GetCodec(r.Metadata[name])
	if encoding != nil {
		return encoding
	}
	return codec.GetCodec(proto.Name)
}
