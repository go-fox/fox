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
	"github.com/redis/go-redis/v9"

	"github.com/go-fox/fox/config"
	"github.com/go-fox/fox/contrib/cache"
)

// Config cache config
type Config struct {
	Prefix string `json:"prefix"`
	Client redis.UniversalClient
}

// Build 使用配置构建缓存
func (c *Config) Build() cache.Cache {
	return NewWithConfig(c)
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{}
}

// RawConfig 使用指定键的配置值
func RawConfig(key string) *Config {
	conf := DefaultConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

// ScanConfig 使用键application.cache.redis.{name}的配置值
func ScanConfig(name ...string) *Config {
	key := "application.cache.redis"
	if len(name) > 0 {
		key = key + "." + name[0]
	}
	return RawConfig(key)
}
