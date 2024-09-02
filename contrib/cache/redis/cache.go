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
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/go-fox/fox/contrib/cache"
)

var _ cache.Cache = (*impl)(nil)

// impl redis版本实现
type impl struct {
	config *Config
	client redis.UniversalClient
}

// New 创建redis缓存根据参数
func New(opts ...Option) cache.Cache {
	conf := DefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewWithConfig(conf)
}

// NewWithConfig 根据配置信息创建redis缓存
func NewWithConfig(configs ...*Config) cache.Cache {
	conf := DefaultConfig()
	if len(configs) > 0 {
		conf = configs[0]
	}
	if conf.Client == nil {
		panic("redis client is required")
	}
	return &impl{
		config: conf,
		client: conf.Client,
	}
}

func (i *impl) Get(ctx context.Context, key string, value any) error {
	storeKey := i.getStoreKey(key)
	result, err := i.client.Get(ctx, storeKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}
	return cache.DefaultSerializer().Unmarshal([]byte(result), value)
}

func (i *impl) Set(ctx context.Context, key string, value any, timeout time.Duration) error {
	key = i.getStoreKey(key)
	marshal, err := cache.DefaultSerializer().Marshal(value)
	if err != nil {
		return err
	}
	return i.client.Set(ctx, key, string(marshal), timeout).Err()
}

func (i *impl) Update(ctx context.Context, key string, value any) error {
	ttl, err := i.TTL(ctx, key)
	if err != nil {
		return err
	}
	return i.Set(ctx, key, value, ttl)
}

func (i *impl) TTL(ctx context.Context, key string) (time.Duration, error) {
	key = i.getStoreKey(key)
	return i.client.TTL(ctx, key).Result()
}

func (i *impl) UpdateTTL(ctx context.Context, key string, ttl time.Duration) error {
	key = i.getStoreKey(key)
	return i.client.Expire(ctx, key, ttl).Err()
}

func (i *impl) Delete(ctx context.Context, key string) error {
	key = i.getStoreKey(key)
	return i.client.Del(ctx, key).Err()
}

//	func (c *impl) Get(ctx context.Context, key string) (string, error) {
//		key = c.getStoreKey(key)
//		return c.client.Get(ctx, key).Scan(v)
//	}
//
//	func (c *impl) Set(ctx context.Context, key string, value string, timeout time.Duration) error {
//		key = c.getStoreKey(key)
//		return c.client.Set(ctx, key, value, timeout).Err()
//	}
//
//	func (c *impl) TTL(ctx context.Context, key string) (time.Duration, error) {
//		key = c.getStoreKey(key)
//		return c.client.TTL(ctx, key).Result()
//	}
//
//	func (c *impl) UpdateTTL(ctx context.Context, key string, ttl time.Duration) error {
//		key = c.getStoreKey(key)
//		return c.client.Expire(ctx, key, ttl).Err()
//	}
//
//	func (c *impl) UpdateObject(ctx context.Context, key string, obj any) error {
//		ttl, err := c.TTL(ctx, key)
//		if err != nil {
//			return err
//		}
//		return c.SetObject(ctx, key, obj, ttl)
//	}
//
//	func (c *impl) GetObject(ctx context.Context, key string, obj any) error {
//		key = c.getStoreKey(key)
//		result, err := c.client.Get(ctx, key).Result()
//		if err != nil {
//			return err
//		}
//		return cache.DefaultSerialize.Unmarshal([]byte(result), obj)
//	}
//
//	func (c *impl) SetObject(ctx context.Context, key string, obj any, timeout time.Duration) error {
//		key = c.getStoreKey(key)
//		storeBytes, err := cache.DefaultSerialize.Marshal(obj)
//		if err != nil {
//			return err
//		}
//		return c.client.Set(ctx, key, storeBytes, timeout).Err()
//	}
func (i *impl) getStoreKey(key string) string {
	if i.config.Prefix != "" {
		return i.config.Prefix + ":" + key
	}
	return key
}
