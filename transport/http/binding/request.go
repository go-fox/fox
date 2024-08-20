// Package binding
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

package binding

import (
	"net/http"
	"net/url"
)

// Request request interface
type Request interface {
	// GetMethod 获取请求方法
	GetMethod() string
	// GetParams 路径参数
	GetParams() url.Values
	// GetQuery 获取get请求参数
	GetQuery() url.Values
	// GetContentType 获取请求方式
	GetContentType() string
	// GetHeader 获取请求头
	GetHeader() http.Header
	// GetCookies 获取cookie
	GetCookies() []*http.Cookie
	// GetBody 获取请求body
	GetBody() ([]byte, error)
	// GetPostForm 获取post请求参数
	GetPostForm() (url.Values, error)
	// GetForm 获取get请求参数
	GetForm() (url.Values, error)
}

// MarshalRequest marshal request interface
type MarshalRequest interface {
	Request
	// GetPath 获取完整路径
	GetPath() string
	// HasBody 是否有body
	HasBody() bool
	//Body 需要请求的body体
	Body() map[string]interface{}
}

type bindRequest struct {
	path    string
	hasBody bool
	params  url.Values
	query   url.Values
	header  http.Header
	cookie  []*http.Cookie
	body    map[string]interface{}
}

func (b bindRequest) GetPath() string {
	return b.path
}

func (b bindRequest) GetMethod() string {
	return ""
}

func (b bindRequest) GetParams() url.Values {
	return b.params
}

func (b bindRequest) GetQuery() url.Values {
	return b.query
}

func (b bindRequest) GetContentType() string {
	return ""
}

func (b bindRequest) GetHeader() http.Header {
	return b.header
}

func (b bindRequest) GetCookies() []*http.Cookie {
	return b.cookie
}

func (b *bindRequest) HasBody() bool {
	return b.hasBody
}

// Body 返回要编码的body
func (b *bindRequest) Body() map[string]interface{} {
	return b.body
}

func (b bindRequest) GetBody() ([]byte, error) {
	return nil, nil
}

func (b bindRequest) GetPostForm() (url.Values, error) {
	return nil, nil
}

func (b bindRequest) GetForm() (url.Values, error) {
	return nil, nil
}
