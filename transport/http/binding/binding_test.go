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
	"testing"
)

type Values struct {
	Id      int64    `json:"id" path:"id"`
	Name    string   `json:"name" query:"name"`
	Cookie  string   `json:"cookie" cookie:"cookie"`
	Auth    string   `json:"auth" header:"auth"`
	Float   float64  `json:"float" header:"float"`
	ArrayT  []string `json:"arrayT" header:"arrayT"`
	Anytest struct {
		Ceshi string `json:"ceshi"`
	} `json:"anytest"`
}

func TestMarshal(t *testing.T) {
	bind := New()
	v := &Values{
		Id:     10,
		Name:   "liuqin",
		Auth:   "asdasdfasdasd",
		Cookie: "123456",
		Float:  3.1415926535,
		ArrayT: []string{"a", "b", "c"},
		Anytest: struct {
			Ceshi string `json:"ceshi"`
		}{Ceshi: "aaa"},
	}
	marshal, err := bind.marshal("/user/{id}", v)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Path:%s，body:%s cookie:%v header:%v", marshal.GetPath(), marshal.Body(), marshal.GetCookies(), marshal.GetHeader())
	test := Values{}
	ps := Params{
		{
			Key:   "id",
			Value: "111",
		},
	}
	err = bind.Unmarshal(&test, marshal, ps)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Value:%v", test)
}
