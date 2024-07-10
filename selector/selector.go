// Package selector
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
package selector

import (
	"context"
	"strings"

	"github.com/go-fox/sugar/container/smap"

	"github.com/go-fox/fox/errors"
)

// ErrNoAvailable is no available node.
var ErrNoAvailable = errors.ServiceUnavailable("no_available_node", "")

var m = smap.New[string, Builder](true)

// DoneInfo is callback info when RPC invoke done.
type DoneInfo struct {
	// Response Error
	Err error
	// Response Metadata
	ReplyMD ReplyMD

	// BytesSent indicates if any bytes have been sent to the server.
	BytesSent bool
	// BytesReceived indicates if any byte has been received from the server.
	BytesReceived bool
}

// ReplyMD is a Reply Metadata.
type ReplyMD interface {
	Get(key string) string
}

// DoneFunc refactor balance.DoneFunc
type DoneFunc func(ctx context.Context, info DoneInfo)

// Register register selector builder
func Register(b Builder) {
	m.Set(strings.ToLower(b.Name()), b)
}

// Get is get a selector builder
func Get(name string) Builder {
	if v, ok := m.Get(strings.ToLower(name)); ok {
		return v
	}
	return nil
}

// GetAll is get all selector builder
func GetAll() *smap.Map[string, Builder] {
	return m
}

// Storage is node storage
type Storage interface {
	Store(nodes []Node)
}

// Selector is node pick balancer
type Selector interface {
	Storage
	// Select is pick node
	Select(ctx context.Context, opts ...SelectOption) (selected Node, done DoneFunc, err error)
}

// Builder is selector builder
type Builder interface {
	Name() string
	Build() Selector
}
