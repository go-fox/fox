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
	"fmt"
)

var _ Locale = (*localeImpl)(nil)

// Locale locale translate interface
type Locale interface {
	Locale(locale string) Locale
	Scope(scope string) Locale
	T(key string, args ...any) string
}

type localeImpl struct {
	locale string
	scope  string
	i18n   *i18n
}

// Locale with locale
func (l *localeImpl) Locale(locale string) Locale {
	return &localeImpl{
		locale: locale,
		i18n:   l.i18n,
		scope:  l.scope,
	}
}

// Scope with scope
func (l *localeImpl) Scope(scope string) Locale {
	if len(l.scope) > 0 {
		scope = fmt.Sprintf("%s.%s", l.scope, scope)
	}
	return &localeImpl{
		locale: l.locale,
		scope:  scope,
		i18n:   l.i18n,
	}
}

// T translate with a key
func (l *localeImpl) T(key string, args ...any) string {
	if len(l.scope) > 0 {
		key = fmt.Sprintf("%s.%s", l.scope, key)
	}
	return l.i18n.T(l.locale, key, args...)
}
