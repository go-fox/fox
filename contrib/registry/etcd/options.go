// Package etcd
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
package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/go-fox/fox/config"
)

// Option is creating a registry option
type Option func(r *Config)

// WithClient with an etcd client option
func WithClient(client *clientv3.Client) Option {
	return func(r *Config) {
		r.Client = client
	}
}

// WithPrefix with prefix option
func WithPrefix(prefix string) Option {
	return func(r *Config) {
		r.Prefix = prefix
	}
}

// WithTTL with ttl option
func WithTTL(ttl time.Duration) Option {
	return func(r *Config) {
		r.TTL = config.Duration{Duration: ttl}
	}
}

// WithMaxRetry with a max retry option
func WithMaxRetry(maxRetry int) Option {
	return func(r *Config) {
		r.MaxRetry = maxRetry
	}
}
