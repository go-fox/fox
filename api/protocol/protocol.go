// Package protocol
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
package protocol

import (
	"github.com/go-fox/sugar/container/spool"
)

var requestPool = spool.New[*Request](func() *Request {
	return &Request{
		Metadata: map[string]string{},
	}
}, func(request *Request) {
	request.Reset()
	request.Metadata = map[string]string{}
})
var replyPool = spool.New[*Reply](func() *Reply {
	return &Reply{
		Metadata: map[string]string{},
	}
}, func(reply *Reply) {
	reply.Reset()
	reply.Metadata = map[string]string{}
})

// AcquireRequest get Request
func AcquireRequest() *Request {
	return requestPool.Get()
}

// ReleaseRequest release Request
func ReleaseRequest(req *Request) {
	requestPool.Put(req)
}

// AcquireReply get Reply
func AcquireReply() *Reply {
	return replyPool.Get()
}

// ReleaseReply release Reply
func ReleaseReply(reply *Reply) {
	replyPool.Put(reply)
}

// WithMetadata setting metadata
func (r *Reply) WithMetadata(metadata map[string]string) *Reply {
	r.Metadata = metadata
	return r
}
