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
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type file struct {
	path any
}

// NewSource file source
func NewSource(path any) Source {
	switch path.(type) {
	case string:
	case fs.File:
	default:
		panic("path must be string or fs.File")
	}
	return &file{path: path}
}

// Load load
func (f *file) Load() ([]*DataSet, error) {
	var dataSet *DataSet
	switch path := f.path.(type) {
	case string:
		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			return f.loadDir(path, make([]*DataSet, 0))
		}
		dataSet, err = f.loadFile(path)
		if err != nil {
			return nil, err
		}
	case fs.File:
		var err error
		dataSet, err = f.loadFile(path)
		if err != nil {
			return nil, err
		}
		return []*DataSet{dataSet}, nil
	}
	return []*DataSet{dataSet}, nil
}

// loadDir load dir
func (f *file) loadDir(path string, dataSets []*DataSet) ([]*DataSet, error) {
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

// loadFile load file
func (f *file) loadFile(path any) (*DataSet, error) {
	var fsFile fs.File
	var err error
	switch p := path.(type) {
	case string:
		fsFile, err = os.Open(p)
		if err != nil {
			return nil, err
		}
	case fs.File:
		fsFile = p
	}
	defer fsFile.Close()
	data, err := io.ReadAll(fsFile)
	if err != nil {
		return nil, err
	}
	info, err := fsFile.Stat()
	if err != nil {
		return nil, err
	}
	return &DataSet{
		Locale:    strings.TrimSuffix(filepath.Base(info.Name()), filepath.Ext(info.Name())),
		Format:    strings.TrimPrefix(filepath.Ext(info.Name()), "."),
		Value:     data,
		Timestamp: info.ModTime(),
	}, nil
}
