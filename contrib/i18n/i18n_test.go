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
	"log/slog"
	"testing"
	"time"
)

const (
	_testZhCnJSON = `{
	"user":{
		"username":"用户名称：{{.name}}"
	}
}`
	_testEnToml = `[user]
username="username：{{.name}}"`
)

type testZhCnJSONSource struct {
}

func (t *testZhCnJSONSource) Load() ([]*DataSet, error) {
	return []*DataSet{
		{
			Locale:    "zh_CN",
			Value:     []byte(_testZhCnJSON),
			Format:    "json",
			Timestamp: time.Now(),
		},
	}, nil
}

type testEnToml struct {
}

func (t *testEnToml) Load() ([]*DataSet, error) {
	return []*DataSet{
		{
			Locale: "en_US",
			Value:  []byte(_testEnToml),
			Format: "toml",
		},
	}, nil
}

func TestI18n(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	n, err := New(
		WithSources(&testZhCnJSONSource{}, &testEnToml{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	username := n.T("en_US", "user.username", map[string]any{
		"name": "en_US",
	})
	t.Logf("translate key:username value:%s", username)

	locale := n.Locale("zh_CN")
	scope := locale.Scope("user")
	s := scope.T("username", map[string]any{
		"name": "zh_CN",
	})
	t.Logf("translate key:username value:%s", s)
}
