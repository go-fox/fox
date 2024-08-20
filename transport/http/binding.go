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
	"net/http"
	"net/url"
	"sync"

	"github.com/go-fox/fox/internal/bytesconv"
	"github.com/go-fox/fox/transport/http/binding"
)

var (
	defaultBinding  = binding.New()
	bindRequestPool = sync.Pool{
		New: func() interface{} {
			return &bindRequest{}
		},
	}
)

// SetBinding set binding instance
func SetBinding(b *binding.Binding) {
	defaultBinding = b
}

func acquireBindRequest(req *Request) *bindRequest {
	request := bindRequestPool.Get().(*bindRequest)
	request.req = req
	return request
}

func releaseBindRequest(req *bindRequest) {
	req.req = nil
	bindRequestPool.Put(req)
}

func warpRequest(req *Request) *bindRequest {
	return acquireBindRequest(req)
}

type bindRequest struct {
	req *Request
}

func (br *bindRequest) GetParams() url.Values {
	return nil
}

func (br *bindRequest) GetMethod() string {
	return bytesconv.BytesToString(br.req.Header.Method())
}

func (br *bindRequest) GetQuery() url.Values {
	queryMap := make(url.Values)
	br.req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := bytesconv.BytesToString(key)
		values := queryMap[keyStr]
		values = append(values, bytesconv.BytesToString(value))
		queryMap[keyStr] = values
	})

	return queryMap
}

func (br *bindRequest) GetContentType() string {
	return bytesconv.BytesToString(br.req.Header.ContentType())
}

func (br *bindRequest) GetHeader() http.Header {
	header := make(http.Header)
	br.req.Header.VisitAll(func(key, value []byte) {
		keyStr := bytesconv.BytesToString(key)
		values := header[keyStr]
		values = append(values, bytesconv.BytesToString(value))
		header[keyStr] = values
	})
	return header
}

func (br *bindRequest) GetCookies() []*http.Cookie {
	var cookies []*http.Cookie
	br.req.Header.VisitAllCookie(func(key, value []byte) {
		cookies = append(cookies, &http.Cookie{
			Name:  string(key),
			Value: string(value),
		})
	})
	return cookies
}

func (br *bindRequest) GetBody() ([]byte, error) {
	return br.req.Body(), nil
}

func (br *bindRequest) GetPostForm() (url.Values, error) {
	postMap := make(url.Values)
	br.req.PostArgs().VisitAll(func(key, value []byte) {
		keyStr := bytesconv.BytesToString(key)
		values := postMap[keyStr]
		values = append(values, bytesconv.BytesToString(value))
		postMap[keyStr] = values
	})
	mf, err := br.req.MultipartForm()
	if err == nil {
		for k, v := range mf.Value {
			if len(v) > 0 {
				postMap[k] = v
			}
		}
	}

	return postMap, nil
}

func (br *bindRequest) GetForm() (url.Values, error) {
	formMap := make(url.Values)
	br.req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := bytesconv.BytesToString(key)
		values := formMap[keyStr]
		values = append(values, bytesconv.BytesToString(value))
		formMap[keyStr] = values
	})
	br.req.PostArgs().VisitAll(func(key, value []byte) {
		keyStr := bytesconv.BytesToString(key)
		values := formMap[keyStr]
		values = append(values, bytesconv.BytesToString(value))
		formMap[keyStr] = values
	})

	return formMap, nil
}
