// Package redis
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
package redis

import (
	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

// Client redis client
type Client = redis.UniversalClient

// New create Client with option
func New(opts ...Option) Client {
	conf := DefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewWithConfig(conf)
}

// NewWithConfig create Client with Config
func NewWithConfig(configs ...*Config) Client {
	conf := DefaultConfig()
	if len(configs) > 0 {
		conf = configs[0]
	}
	if conf.TLSConfig == nil && conf.CertFile != "" && conf.KeyFile != "" {
		pair, err := tls.LoadX509KeyPair(conf.CertFile, conf.KeyFile)
		if err != nil {
			panic(err)
		}
		conf.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{pair},
		}
	}
	return redis.NewUniversalClient(conf.toRedisConf())
}
