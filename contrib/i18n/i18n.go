// Package i18n
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
package i18n

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"dario.cat/mergo"

	// init codec
	_ "github.com/go-fox/fox/codec/json"
	_ "github.com/go-fox/fox/codec/proto"
	_ "github.com/go-fox/fox/codec/toml"
	_ "github.com/go-fox/fox/codec/xml"
	_ "github.com/go-fox/fox/codec/yaml"
)

var _ I18n = (*i18n)(nil)

// I18n i18n interface
type I18n interface {
	Locale(locale string) Locale
	Scope(scope string) Locale
	T(locale, key string, args ...interface{}) string
}

type i18n struct {
	debug    bool
	sources  []Source
	decoder  Decoder
	resolver Resolver
	cached   sync.Map
	merge    Merge
	reader   Reader
	parser   Parser
}

// New create an i18n
func New(opts ...Option) (I18n, error) {
	i := &i18n{
		sources:  []Source{},
		decoder:  defaultDecoder,
		resolver: defaultResolver,
		cached:   sync.Map{},
		merge: func(dst, src interface{}) error {
			return mergo.Map(dst, src, mergo.WithOverride)
		},
		parser: defaultParse,
	}
	for _, opt := range opts {
		opt(i)
	}
	i.reader = newReader(i)
	if err := i.load(); err != nil {
		return nil, err
	}
	return i, nil
}

func (i *i18n) load() error {
	for _, src := range i.sources {
		dataset, err := src.Load()
		if err != nil {
			return err
		}
		for _, v := range dataset {
			if i.debug {
				slog.Debug(fmt.Sprintf("i18n loaded locale: %s format: %s", v.Locale, v.Format))
			}
		}
		if err = i.reader.Merge(dataset...); err != nil {
			return err
		}
	}
	return i.reader.Resolve()
}

// Locale with locale
func (i *i18n) Locale(locale string) Locale {
	return &localeImpl{
		locale: locale,
		i18n:   i,
	}
}

// Scope with scope
func (i *i18n) Scope(scope string) Locale {
	return &localeImpl{
		scope: scope,
		i18n:  i,
	}
}

// T translate with locale and key
func (i *i18n) T(locale, key string, args ...any) string {
	storeKey := strings.Join([]string{locale, key}, ".")
	value := ""
	if v, ok := i.cached.Load(storeKey); ok {
		value = v.(Value).String()
	}
	if v, ok := i.reader.Value(storeKey); ok {
		i.cached.Store(key, v)
		value = v.(Value).String()
	}
	parser, err := i.parser(storeKey, value, args...)
	if err != nil {
		if i.debug {
			slog.Error("i18n translate error", "error", err, "locale", locale, "key", key, "args", args)
		}
		return ""
	}
	return parser
}
