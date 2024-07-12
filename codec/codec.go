// Package codec
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
package codec

import (
	"fmt"
	"strings"

	"github.com/go-fox/sugar/container/smap"
)

var registeredCodecs = smap.New[string, Codec](true)

// Codec codec interface
type Codec interface {
	Name() string
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// RegisterCodec register codec
func RegisterCodec(codec Codec) {
	if codec == nil {
		panic("codec: Register codec is nil")
	}
	if codec.Name() == "" {
		panic("codec: Register codec name is empty")
	}
	name := strings.ToLower(codec.Name())
	registeredCodecs.Set(name, codec)
}

// GetCodec load codec
func GetCodec(name string) (Codec, error) {
	name = strings.ToLower(name)
	codec, ok := registeredCodecs.Get(name)
	if !ok {
		return nil, fmt.Errorf("codec: codec `%s` not found", name)
	}
	return codec, nil
}
