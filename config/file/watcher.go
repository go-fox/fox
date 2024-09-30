// Package file
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
package file

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"github.com/go-fox/fox/config"
)

type watcher struct {
	f      *file
	ctx    context.Context
	cancel context.CancelFunc
	fw     *fsnotify.Watcher
}

func (w watcher) Next() ([]*config.DataSet, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case event := <-w.fw.Events:
		if event.Op == fsnotify.Rename {
			if _, err := os.Stat(event.Name); err == nil || os.IsExist(err) {
				if err := w.fw.Add(event.Name); err != nil {
					return nil, err
				}
			}
		}
		fi, err := os.Stat(w.f.path.(string))
		if err != nil {
			return nil, err
		}
		path := w.f.path
		if fi.IsDir() {
			path = filepath.Join(w.f.path.(string), filepath.Base(event.Name))
		}
		kv, err := w.f.loadFile(path)
		if err != nil {
			return nil, err
		}
		return []*config.DataSet{kv}, nil
	case err := <-w.fw.Errors:
		return nil, err
	}
}

func (w watcher) Stop() error {
	w.cancel()
	return w.fw.Close()
}

func newWatcher(f *file) (config.Watcher, error) {
	switch f.path.(type) {
	case fs.File:
		return nil, errors.New("fs.File not supported")
	case string:
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	if err = w.Add(f.path.(string)); err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &watcher{
		f:      f,
		fw:     w,
		ctx:    ctx,
		cancel: cancelFunc,
	}, nil
}
