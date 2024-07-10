// Package nacos
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
package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
)

// Option is registry option
type Option func(r *Registry)

// WithClient with a nacos client option
func WithClient(client naming_client.INamingClient) Option {
	return func(r *Registry) {
		r.cli = client
	}
}

// WithPrefix with a prefix option
func WithPrefix(prefix string) Option {
	return func(r *Registry) { r.prefix = prefix }
}

// WithWeight with a weight option.
func WithWeight(weight float64) Option {
	return func(r *Registry) { r.weight = weight }
}

// WithCluster with a cluster option.
func WithCluster(cluster string) Option {
	return func(r *Registry) { r.cluster = cluster }
}

// WithGroup with a group option.
func WithGroup(group string) Option {
	return func(r *Registry) { r.group = group }
}
