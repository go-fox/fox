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

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/json"
	"github.com/go-fox/fox/errors"
	"github.com/go-fox/fox/internal/bytesconv"
	"github.com/go-fox/fox/internal/httputil"
)

// Redirector URL redirector
type Redirector interface {
	Redirect() (string, int)
}

// EncodeResponseFunc response handler
type EncodeResponseFunc func(ctx *Context, v any) error

// EncodeErrorFunc is error handler
type EncodeErrorFunc func(ctx *Context, err error) error

// EncodeRequestFunc client request encoder
type EncodeRequestFunc = func(ctx context.Context, req *Request, v any) ([]byte, error)

// DecodeResponseFunc client response decoder
type DecodeResponseFunc = func(ctx context.Context, resp *Response, v any) error

// DecodeErrorFunc client error decoder
type DecodeErrorFunc func(ctx context.Context, res *Response) error

// CodecForRequest get codec from request
func CodecForRequest(r *Request, name string) codec.Codec {
	impl := codec.GetCodec(httputil.ContentSubtype(bytesconv.BytesToString(r.Header.Peek(name))))
	if impl != nil {
		return impl
	}
	return codec.GetCodec(json.Name)
}

// CodecForResponse 根据响应信息获取解码器
func CodecForResponse(response *Response) codec.Codec {
	c := codec.GetCodec(httputil.ContentSubtype(bytesconv.BytesToString(response.Header.Peek(HeaderContentType))))
	if c != nil {
		return c
	}
	return codec.GetCodec(json.Name)
}

// DefaultResponseHandler default response handler
func DefaultResponseHandler(ctx *Context, v any) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(Redirector); ok {
		url, code := rd.Redirect()
		return ctx.Redirect(code, bytesconv.StringToBytes(url))
	}
	c := CodecForRequest(ctx.Request(), "Accept")
	data, err := c.Marshal(v)
	if err != nil {
		return err
	}
	ctx.Response().Header.Set("Content-Type", httputil.ContentType(c.Name()))
	ctx.Send(data)
	return nil
}

// DefaultErrorHandler default error handler
func DefaultErrorHandler(c *Context, err error) error {
	code := StatusInternalServerError
	var e *errors.Error
	body := err.Error()
	if errors.As(err, &e) {
		code = int(e.Code)
		subtype := httputil.ContentSubtype(c.GetRequestHeader(HeaderContentType))
		switch subtype {
		case "json":
			body = e.JSON()
		default:
			body = e.Error()
		}
	}
	c.SetResponseHeader(HeaderContentType, MIMETextPlainCharsetUTF8)
	return c.SetStatusCode(code).SendString(body)
}

// DefaultRequestEncoder default client request encoder
func DefaultRequestEncoder(ctx context.Context, req *Request, in interface{}) ([]byte, error) {
	encoding := CodecForRequest(req, HeaderContentType)
	body, err := encoding.Marshal(in)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// DefaultResponseDecoder default client response decoder
func DefaultResponseDecoder(ctx context.Context, resp *Response, out interface{}) error {
	data := resp.Body()
	return CodecForResponse(resp).Unmarshal(data, out)
}

// DefaultErrorDecoder default client error handler
func DefaultErrorDecoder(ctx context.Context, resp *Response) error {
	if resp.StatusCode() >= 200 && resp.StatusCode() <= 299 {
		return nil
	}
	data := resp.Body()
	e := new(errors.Error)
	err := CodecForResponse(resp).Unmarshal(data, e)
	if err == nil {
		e.Code = int32(resp.StatusCode())
		return e
	}
	return errors.New(resp.StatusCode(), errors.UnknownReason, "").WithCause(err)
}
