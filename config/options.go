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
	"regexp"
	"strings"

	"github.com/go-fox/fox/codec"
)

// Decoder is config decoder.
type Decoder func(*DataSet, map[string]interface{}) error

// Resolver resolve placeholder in config.
type Resolver func(map[string]interface{}) error

// Merge is config merge func.
type Merge func(dst, src interface{}) error

type options struct {
	sources  []Source
	decoder  Decoder
	resolver Resolver
	merge    Merge
}

// Option 构造函数参数
type Option func(o *options)

// WithSources configure sources
//
//	@param sources ...Source
//	@return Option
//	@player
func WithSources(sources ...Source) Option {
	return func(o *options) {
		o.sources = sources
	}
}

// WithDecoder configure Decoder
//
//	@param decoder Decoder
//	@return Option
//	@player
func WithDecoder(decoder Decoder) Option {
	return func(o *options) {
		o.decoder = decoder
	}
}

// WithResolver configure Resolver
//
//	@param resolver Resolver
//	@return Option
//	@player
func WithResolver(resolver Resolver) Option {
	return func(o *options) {
		o.resolver = resolver
	}
}

// WithMerge configure Merge
//
//	@param merge Merge
//	@return Option
//	@player
func WithMerge(merge Merge) Option {
	return func(o *options) {
		o.merge = merge
	}
}

// defaultDecoder decode config from source KeyValue
// to target map[string]interface{} using src.Format codec.
func defaultDecoder(src *DataSet, target map[string]interface{}) error {
	if src.Format == "" {
		// expand key "aaa.bbb" into map[aaa]map[bbb]interface{}
		keys := strings.Split(src.Key, ".")
		for i, k := range keys {
			if i == len(keys)-1 {
				target[k] = src.Value
			} else {
				sub := make(map[string]interface{})
				target[k] = sub
				target = sub
			}
		}
		return nil
	}
	encoding, err := codec.GetCodec(src.Format)
	if err != nil {
		return err
	}
	return encoding.Unmarshal(src.Value, &target)
}

// defaultResolver resolve placeholder in map value,
// placeholder format in ${key:default}.
func defaultResolver(input map[string]interface{}) error {
	mapper := func(name string) string {
		args := strings.SplitN(strings.TrimSpace(name), ":", 2) //nolint:gomnd
		if v, has := readValue(input, args[0]); has {
			s, _ := v.String()
			return s
		} else if len(args) > 1 { // default value
			return args[1]
		}
		return ""
	}

	var resolve func(map[string]interface{}) error
	resolve = func(sub map[string]interface{}) error {
		for k, v := range sub {
			switch vt := v.(type) {
			case string:
				sub[k] = expand(vt, mapper)
			case map[string]interface{}:
				if err := resolve(vt); err != nil {
					return err
				}
			case []interface{}:
				for i, iface := range vt {
					switch it := iface.(type) {
					case string:
						vt[i] = expand(it, mapper)
					case map[string]interface{}:
						if err := resolve(it); err != nil {
							return err
						}
					}
				}
				sub[k] = vt
			}
		}
		return nil
	}
	return resolve(input)
}

func expand(s string, mapping func(string) string) string {
	r := regexp.MustCompile(`\${(.*?)}`)
	re := r.FindAllStringSubmatch(s, -1)
	for _, i := range re {
		if len(i) == 2 { //nolint:gomnd
			s = strings.ReplaceAll(s, i[0], mapping(i[1]))
		}
	}
	return s
}
