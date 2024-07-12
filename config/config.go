// Package config
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
package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"dario.cat/mergo"
	// init codec
	_ "github.com/go-fox/fox/codec/json"
	_ "github.com/go-fox/fox/codec/proto"
	_ "github.com/go-fox/fox/codec/toml"
	_ "github.com/go-fox/fox/codec/xml"
	_ "github.com/go-fox/fox/codec/yaml"
)

var _ Config = (*config)(nil)

// ErrorNotFound not found error
var ErrorNotFound = errors.New("not found key")

// Observer config observer
type Observer func(string, Value)

// Config interface
type Config interface {
	Load(sources ...Source) error
	Scan(v interface{}) error
	Get(key string) Value
	Watch(key string, o Observer) error
	Close() error
}

// config implement for Config
type config struct {
	opts      options
	reader    Reader
	cached    sync.Map
	observers sync.Map
	watchers  []Watcher
}

// New 构造函数
//
//	@param opts ...Option 创建参数
//	@return Config
func New(opts ...Option) Config {
	o := options{
		decoder:  defaultDecoder,
		resolver: defaultResolver,
		merge: func(dst, src interface{}) error {
			return mergo.Map(dst, src, mergo.WithOverride)
		},
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &config{
		opts:   o,
		reader: newReader(o),
	}
}

func (c *config) Load(sources ...Source) error {
	c.opts.sources = append(c.opts.sources, sources...)
	for _, src := range c.opts.sources {
		dataset, err := src.Load()
		if err != nil {
			return err
		}
		for _, v := range dataset {
			slog.Debug(fmt.Sprintf("config loaded: %s format: %s", v.Key, v.Format))
		}
		if err = c.reader.Merge(dataset...); err != nil {
			slog.Error("failed to merge config source", "err", err)
			return err
		}
		w, err := src.Watch()
		if err != nil {
			slog.Error("failed to watch config source", "err", err)
			return err
		}
		c.watchers = append(c.watchers, w)
		go c.watch(w)
	}
	if err := c.reader.Resolve(); err != nil {
		slog.Error("failed to resolve config source", "err", err)
		return err
	}
	return nil
}

func (c *config) Scan(v interface{}) error {
	data, err := c.reader.Source()
	if err != nil {
		return err
	}
	return unmarshalJSON(data, v)
}

func (c *config) Get(key string) Value {
	if v, ok := c.cached.Load(key); ok {
		return v.(Value)
	}
	if v, ok := c.reader.Value(key); ok {
		c.cached.Store(key, v)
		return v
	}
	return &errValue{err: ErrorNotFound}
}

func (c *config) Watch(key string, o Observer) error {
	if v := c.Get(key); v.Load() == nil {
		return ErrorNotFound
	}
	c.observers.Store(key, o)
	return nil
}

func (c *config) Close() error {
	for _, w := range c.watchers {
		if err := w.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (c *config) watch(w Watcher) {
	for {
		kvs, err := w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("watcher's ctx cancel", "err", err)
				return
			}
			time.Sleep(time.Second)
			slog.Error("failed to watch next config", "err", err)
			continue
		}
		if err := c.reader.Merge(kvs...); err != nil {
			slog.Error("failed to merge next config", "err", err)
			continue
		}
		if err := c.reader.Resolve(); err != nil {
			slog.Error("failed to resolve next config", "err", err)
			continue
		}
		c.cached.Range(func(key, value interface{}) bool {
			k := key.(string)
			v := value.(Value)
			if n, ok := c.reader.Value(k); ok && reflect.TypeOf(n.Load()) == reflect.TypeOf(v.Load()) && !reflect.DeepEqual(n.Load(), v.Load()) {
				v.Store(n.Load())
				if o, ok := c.observers.Load(k); ok {
					o.(Observer)(k, v)
				}
			}
			return true
		})
	}
}
