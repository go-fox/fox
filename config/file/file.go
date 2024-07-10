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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-fox/fox/config"
)

type file struct {
	path string
}

// NewSource 创建文件资源管理器
func NewSource(path string) config.Source {
	return &file{path: path}
}

// Load 加载资源
func (f *file) Load() ([]*config.DataSet, error) {
	stat, err := os.Stat(f.path)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return f.loadDir(f.path, make([]*config.DataSet, 0))
	}
	dataSet, err := f.loadFile(f.path)
	if err != nil {
		return nil, err
	}
	return []*config.DataSet{dataSet}, nil
}

// Watch 启动文件监控
func (f *file) Watch() (config.Watcher, error) {
	return newWatcher(f)
}

// loadDir 加载文件夹文件
func (f *file) loadDir(path string, dataSets []*config.DataSet) ([]*config.DataSet, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range files {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		dataSet, err := f.loadFile(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, err
		}
		dataSets = append(dataSets, dataSet)
	}
	return dataSets, nil
}

// loadFile 加载文件
func (f *file) loadFile(path string) (*config.DataSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return &config.DataSet{
		Key:       info.Name(),
		Format:    strings.TrimPrefix(filepath.Ext(info.Name()), "."),
		Value:     data,
		Timestamp: info.ModTime(),
	}, nil
}
