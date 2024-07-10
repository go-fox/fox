// Package matcher
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
package matcher

import (
	"sort"
	"strings"

	"github.com/go-fox/fox/middleware"
)

// Matcher is a middleware matcher.
type Matcher interface {
	Use(ms ...middleware.Middleware)
	Match(operation string) []middleware.Middleware
	Add(selector string, ms ...middleware.Middleware)
}

// New return a middleware matcher
//
//	@return Matcher
//	@player
func New() Matcher {
	return &matcher{
		matchs: make(map[string][]middleware.Middleware),
	}
}

type matcher struct {
	prefix   []string
	defaults []middleware.Middleware
	matchs   map[string][]middleware.Middleware
}

func (m *matcher) Use(ms ...middleware.Middleware) {
	m.defaults = ms
}

func (m *matcher) Add(selector string, ms ...middleware.Middleware) {
	if strings.HasSuffix(selector, "*") {
		selector = strings.TrimSuffix(selector, "*")
		m.prefix = append(m.prefix, selector)
		// sort the prefix:
		//  - /foo/bar
		//  - /foo
		sort.Slice(m.prefix, func(i, j int) bool {
			return m.prefix[i] > m.prefix[j]
		})
	}
	m.matchs[selector] = ms
}

func (m *matcher) Match(operation string) []middleware.Middleware {
	ms := make([]middleware.Middleware, 0, len(m.defaults))
	if len(m.defaults) > 0 {
		ms = append(ms, m.defaults...)
	}
	if next, ok := m.matchs[operation]; ok {
		return append(ms, next...)
	}
	for _, prefix := range m.prefix {
		if strings.HasPrefix(operation, prefix) {
			return append(ms, m.matchs[prefix]...)
		}
	}
	return ms
}
