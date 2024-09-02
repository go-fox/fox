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

// Config etcd registry config
type Config struct {
	Prefix   string          `json:"prefix"`
	TTL      config.Duration `json:"ttl"`
	MaxRetry int             `json:"max_retry"`
	Client   *clientv3.Client
}

// Build builder Registry
func (c *Config) Build() *Registry {
	return NewWithConfig(c)
}

// DefaultConfig default config
func DefaultConfig() *Config {
	return &Config{
		Prefix:   "/fox",
		TTL:      config.Duration{Duration: time.Second * 15},
		MaxRetry: 5,
	}
}

// RawConfig scan config value to Config
func RawConfig(key string) *Config {
	conf := DefaultConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

// ScanConfig scan config value to Config
func ScanConfig(names ...string) *Config {
	key := "application.registry.etcd"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawConfig(key)
}
