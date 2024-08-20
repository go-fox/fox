// Package http
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
package http

// RouteParams route params
type RouteParams struct {
	keys, values []string
}

// Add will append a URL parameter to the end of the route param
func (r *RouteParams) Add(key, value string) {
	r.keys = append(r.keys, key)
	r.values = append(r.values, value)
}

// remove remove
func (r *RouteParams) remove(index int) {
	r.keys = append(r.keys[:index], r.keys[index+1:]...)
}

// Get will get path param value
func (r *RouteParams) Get(key string) (string, bool) {
	for i, s := range r.keys {
		if s == key {
			return r.values[i], true
		}
	}
	return "", false
}

// Keys get all keys
func (r *RouteParams) Keys() []string {
	return r.keys
}

// Reset Reset Data
func (r *RouteParams) Reset() {
	r.keys = r.keys[:0]
	r.values = r.values[:0]
}
