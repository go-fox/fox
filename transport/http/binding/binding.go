// Package binding
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

package binding

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/bytedance/go-tagexpr/v2/binding"

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/json"
	"github.com/go-fox/fox/internal/bytesconv"
)

// PathReplace path replaces func
type PathReplace func(name, value string, path string) string

// Binding binding
type Binding struct {
	options *Options
	bind    *binding.Binding
}

// New create binding
func New(opts ...Option) *Binding {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	bind := binding.New(&binding.Config{
		// PathParam use 'path' by default when empty
		PathParam: options.PathParam,
		// Query use 'query' by default when empty
		Query: options.Query,
		// Header use 'header' by default when empty
		Header: options.Header,
		// Cookie use 'cookie' by default when empty
		Cookie: options.Cookie,
		// RawBody use 'raw' by default when empty
		RawBody: options.RawBody,
		// FormBody use 'form' by default when empty
		FormBody: options.FormBody,
		// Validator use 'vd' by default when empty
		Validator: options.Validator,
	})
	bind.ResetJSONUnmarshaler(codec.GetCodec(json.Name).Unmarshal)
	return &Binding{
		options: options,
		bind:    bind,
	}
}

// Marshal 解析出Request
func (b *Binding) Marshal(path string, pointer interface{}) (req MarshalRequest, err error) {
	//1. 反解
	return b.marshal(path, pointer)
}

func (b *Binding) marshal(path string, pointer interface{}) (req MarshalRequest, err error) {
	elemValue, err := b.receiverValueOf(pointer)
	if err != nil {
		return
	}
	if elemValue.Kind() == reflect.Struct {
		req, err = b.marshalStruct(path, elemValue)
		return
	}
	err = fmt.Errorf("pointer：%v is not a struct", pointer)
	return
}

func (b *Binding) marshalStruct(path string, elemValue reflect.Value) (MarshalRequest, error) {
	req := new(bindRequest)
	req.path = path
	req.params = url.Values{}
	req.query = url.Values{}
	req.body = map[string]interface{}{}
	req.header = http.Header{}
	rte := elemValue.Type()
	for i := 0; i < elemValue.NumField(); i++ {
		for _, tag := range b.options.list {
			tagName := rte.Field(i).Tag.Get(tag)
			if len(tagName) > 0 {
				switch tag {
				case b.options.PathParam:
					req.params.Set(tagName, b.getStringValue(elemValue.Field(i)))
				case b.options.Query:
					req.query.Set(tagName, b.getStringValue(elemValue.Field(i)))
				case b.options.Header:
					req.header.Set(tagName, b.getStringValue(elemValue.Field(i)))
				case b.options.Cookie:
					req.cookie = append(req.cookie, &http.Cookie{
						Name:  tagName,
						Value: b.getStringValue(elemValue.Field(i)),
					})
				case b.options.FormBody:
					req.body[tagName] = elemValue.Field(i).Interface()
					if !req.hasBody {
						req.hasBody = true
					}
				case b.options.jsonBody:
					req.body[tagName] = elemValue.Field(i).Interface()
					if !req.hasBody {
						req.hasBody = true
					}
				case b.options.protobufBody:
					req.body[tagName] = elemValue.Field(i).Interface()
					if !req.hasBody {
						req.hasBody = true
					}
				default:
					req.body[tagName] = elemValue.Field(i).Interface()
					if !req.hasBody {
						req.hasBody = true
					}
				}
				break
			}
		}
	}
	for key, param := range req.params {
		value := ""
		if len(param) > 0 {
			value = param[0]
		}
		req.path = b.options.pathReplace(key, value, req.path)
	}
	return req, nil
}

func (b *Binding) getStringValue(value reflect.Value) string {
	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatInt(value.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Float())
	case reflect.Bool:
		return fmt.Sprintf("%v", value.Bool())
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct, reflect.Interface:
		v, _ := codec.GetCodec(json.Name).Marshal(value.Interface())
		return bytesconv.BytesToString(v)
	default:
		return ""
	}
}

func (b *Binding) receiverValueOf(receiver interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(receiver)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		if v.IsValid() && v.CanAddr() {
			return v, nil
		}
	}
	return v, errors.New("receiver must be a non-nil pointer")
}

// Unmarshal 绑定参数
func (b *Binding) Unmarshal(out interface{}, request Request, params PathParams) error {
	return b.bind.IBind(out, request, params)
}
