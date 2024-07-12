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
	"regexp"
	"strings"

	"github.com/google/safetext/yamltemplate"

	"github.com/go-fox/fox/codec"
)

// Decoder is config decoder.
type Decoder func(*DataSet, map[string]interface{}) error

// Resolver resolve placeholder in i18n.
type Resolver func(map[string]interface{}) error

// Merge is i18n merge func.
type Merge func(dst, src interface{}) error

// Parser is template parse func
type Parser func(key, template string, data ...any) (string, error)

// Option create i18n option
type Option func(i *i18n)

// WithSources with sources
func WithSources(sources ...Source) Option {
	return func(i *i18n) {
		i.sources = append(i.sources, sources...)
	}
}

// WithDecoder with decoder
func WithDecoder(decoder Decoder) Option {
	return func(i *i18n) {
		i.decoder = decoder
	}
}

// WithParse with template parser
func WithParse(parser Parser) Option {
	return func(i *i18n) {
		i.parser = parser
	}
}

// WithDebug with debug
func WithDebug(debug bool) Option {
	return func(i *i18n) {
		i.debug = debug
	}
}

// WithResolver with resolver
func WithResolver(resolver Resolver) Option {
	return func(i *i18n) {
		i.resolver = resolver
	}
}

// WithMerge with merge
func WithMerge(merge Merge) Option {
	return func(i *i18n) {
		i.merge = merge
	}
}

// defaultDecoder decode i18n from source KeyValue
// to target map[string]interface{} using src.Format codec.
func defaultDecoder(src *DataSet, target map[string]interface{}) error {
	if src.Format == "" {
		// expand key "aaa.bbb" into map[aaa]map[bbb]interface{}
		keys := strings.Split(src.Locale, ".")
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
	sub := map[string]interface{}{}
	err = encoding.Unmarshal(src.Value, &sub)
	if err != nil {
		return err
	}
	target[src.Locale] = sub
	return nil
}

// defaultResolver resolve placeholder in map value,
// placeholder format in ${key:default}.
func defaultResolver(input map[string]interface{}) error {
	mapper := func(name string) string {
		args := strings.SplitN(strings.TrimSpace(name), ":", 2) //nolint:gomnd
		if v, has := readValue(input, args[0]); has {
			return v.String()
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

func defaultParse(key, template string, args ...any) (string, error) {
	if len(args) == 0 {
		return template, nil
	}
	data := args[0]
	parse, err := yamltemplate.New(key).Parse(template)
	if err != nil {
		return "", err
	}
	sb := new(strings.Builder)
	if err := parse.Execute(sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
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
