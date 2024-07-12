// Package toml
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
package toml

import (
	"bytes"
	"sync"

	"github.com/BurntSushi/toml"

	"github.com/go-fox/fox/codec"
)

// Name toml编码器名称
const Name = "toml"

var initOnce sync.Once

// Codec toml编解码器实现
type Codec struct{}

func init() {
	initOnce.Do(func() {
		codec.RegisterCodec(Codec{})
	})
}

// Name codec name
func (Codec) Name() string {
	return Name
}

// Marshal returns the toml encoding of v.
func (Codec) Marshal(v interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte(""))
	err := toml.NewEncoder(buf).Encode(v)
	return buf.Bytes(), err
}

// Unmarshal parses the toml-encoded data and stores the result
func (Codec) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}
