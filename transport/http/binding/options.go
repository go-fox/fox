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
	"github.com/andeya/goutil"
	"strings"
)

var (
	defaultTagPath   = "path"
	defaultTagQuery  = "query"
	defaultTagHeader = "header"
	defaultTagCookie = "cookie"
	defaultTagForm   = "form"
	tagProtobuf      = "protobuf"
	tagJSON          = "json"
)

type Option func(o *Options)

type Options struct {
	// LooseZeroMode if set to true,
	// the empty string request parameter is bound to the zero value of parameter.
	// NOTE: Suitable for these parameter types: query/header/cookie/form .
	LooseZeroMode bool
	// PathParam use 'path' by default when empty
	PathParam string
	// Query use 'query' by default when empty
	Query string
	// Header use 'header' by default when empty
	Header string
	// Cookie use 'cookie' by default when empty
	Cookie string
	// RawBody use 'raw' by default when empty
	RawBody string
	// FormBody use 'form' by default when empty
	FormBody string
	// Validator use 'vd' by default when empty
	Validator string
	// protobufBody use 'protobuf' by default when empty
	protobufBody string
	// jsonBody use 'json' by default when empty
	jsonBody string
	// defaultVal use 'default' by default when empty
	defaultVal  string
	pathReplace PathReplace
	list        []string
}

func DefaultOptions() *Options {
	options := new(Options)
	options.list = []string{
		goutil.InitAndGetString(&options.PathParam, defaultTagPath),
		goutil.InitAndGetString(&options.Query, defaultTagQuery),
		goutil.InitAndGetString(&options.Header, defaultTagHeader),
		goutil.InitAndGetString(&options.Cookie, defaultTagCookie),
		goutil.InitAndGetString(&options.FormBody, defaultTagForm),
		goutil.InitAndGetString(&options.protobufBody, tagProtobuf),
		goutil.InitAndGetString(&options.jsonBody, tagJSON),
	}
	options.pathReplace = func(name, value string, path string) string {
		return strings.Replace(path, ":"+name, value, -1)
	}
	return options
}
