// Package xml
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
package xml

import (
	"encoding/xml"
	"sync"

	"github.com/go-fox/fox/codec"
)

// Name xml编解码器名称
const Name = "xml"

var initOnce sync.Once

func init() {
	initOnce.Do(func() {
		codec.RegisterCodec(Codec{})
	})
}

// Codec xml编解码器的实现
type Codec struct{}

// Marshal returns the xml encoding of v.
func (Codec) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

// Unmarshal parses the xml-encoded data and stores the result
func (Codec) Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// Name codec name
func (Codec) Name() string {
	return Name
}
