// Package redis
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
package redis

import (
	"context"
	"testing"

	"github.com/go-fox/fox/contrib/clients/redis"
)

func TestCache(t *testing.T) {
	testStringKey := "user"
	testObjKey := "user-info"
	client := redis.New()
	cache := New(WithClient(client), WithPrefix("test"))
	if err := cache.Set(context.Background(), testStringKey, "value", 0); err != nil {
		t.Error(err)
	}
	var value string
	if err := cache.Get(context.Background(), testStringKey, &value); err != nil {
		t.Error(err)
	}
	if value != "value" {
		t.Error("value not equal")
	}
	type Obj struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	objValue := &Obj{
		Name:  "liuqin",
		Value: "10086",
	}
	if err := cache.Set(context.Background(), testObjKey, objValue, 0); err != nil {
		t.Error(err)
	}
	getObjValue := &Obj{}
	if err := cache.Get(context.Background(), testObjKey, getObjValue); err != nil {
		t.Error(err)
	}
	if getObjValue.Name != objValue.Name || getObjValue.Value != objValue.Value {
		t.Error("value not equal")
	}

}
