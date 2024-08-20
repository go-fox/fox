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

type callHook int

const (
	before callHook = iota
	after
)

type (
	callInfo struct {
		contentType  string // body type
		operation    string // grpc operation
		pathTemplate string // http path template
	}
	// CallOption http invoke option
	CallOption func(info *callInfo, hook callHook, ctx interface{}) error
)

func defaultCallInfo(path string) *callInfo {
	return &callInfo{
		contentType:  "application/json",
		operation:    path,
		pathTemplate: path,
	}
}

// WithCallOperation with an operation option
func WithCallOperation(op string) CallOption {
	return func(info *callInfo, hook callHook, ctx interface{}) error {
		if hook == before {
			info.operation = op
		}
		return nil
	}
}

// WithCallContentType with a content type option
func WithCallContentType(contentType string) CallOption {
	return func(info *callInfo, hook callHook, ctx interface{}) error {
		if hook == before {
			info.contentType = contentType
		}
		return nil
	}
}

// WithCallPathTemplate with path template option
func WithCallPathTemplate(pathTemplate string) CallOption {
	return func(info *callInfo, hook callHook, ctx interface{}) error {
		if hook == before {
			info.pathTemplate = pathTemplate
		}
		return nil
	}
}

// WithCallBeforeHeader with before header options
func WithCallBeforeHeader(key, value string) CallOption {
	return func(info *callInfo, hook callHook, ctx interface{}) error {
		if hook == before {
			req, ok := ctx.(*Request)
			if ok {
				req.Header.Set(key, value)
			}
		}
		return nil
	}
}

// WithCallAfterHeader with after header options
func WithCallAfterHeader(key, value string) CallOption {
	return func(info *callInfo, hook callHook, ctx interface{}) error {
		if hook == after {
			resp, ok := ctx.(*Response)
			if ok {
				resp.Header.Set(key, value)
			}
		}
		return nil
	}
}
