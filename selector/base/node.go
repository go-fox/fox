// Package base
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
package base

import (
	"strconv"

	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/selector"
)

var (
	_ selector.Node = (*baseNode)(nil)
)

// NewNode new node
func NewNode(scheme, addr string, ins *registry.ServiceInstance) selector.Node {
	n := &baseNode{
		scheme: scheme,
		addr:   addr,
	}
	if ins != nil {
		n.name = ins.Name
		n.version = ins.Version
		n.metadata = ins.Metadata
		if str, ok := ins.Metadata["weight"]; ok {
			if weight, err := strconv.ParseInt(str, 10, 64); err == nil {
				n.weight = &weight
			}
		}
	}
	return n
}

type baseNode struct {
	scheme   string
	addr     string
	weight   *int64
	version  string
	name     string
	metadata map[string]string
}

func (b *baseNode) Scheme() string {
	return b.scheme
}

func (b *baseNode) Address() string {
	return b.addr
}

func (b *baseNode) ServiceName() string {
	return b.name
}

func (b *baseNode) InitialWeight() *int64 {
	return b.weight
}

func (b *baseNode) Version() string {
	return b.version
}

func (b *baseNode) Metadata() map[string]string {
	return b.metadata
}
