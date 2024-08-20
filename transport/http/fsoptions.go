// Package static
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

import (
	"io/fs"
	"time"

	"github.com/valyala/fasthttp"
)

type fsOptions struct {
	// FS is the file system to serve the static files from.
	// You can use interfaces compatible with fs.FS like embed.FS, os.DirFS etc.
	//
	// Optional. Default: nil
	fs fs.FS
	// root file root path
	//
	// Optional. Default: "."
	root string
	// next defines a function to skip this middleware when returned true.
	//
	// Optional. Default: nil
	next func(c *Context) bool

	// ModifyResponse defines a function that allows you to alter the response.
	//
	// Optional. Default: nil
	modifyResponse Handler

	// NotFoundHandler defines a function to handle when the path is not found.
	//
	// Optional. Default: nil
	notFoundHandler Handler

	// The names of the index files for serving a directory.
	//
	// Optional. Default: []string{"index.html"}.
	indexNames []string

	// Expiration duration for inactive file handlers.
	// Use a negative time.Duration to disable it.
	//
	// Optional. Default: 10 * time.Second.
	cacheDuration time.Duration

	// The value for the Cache-Control HTTP-header
	// that is set on the file response. MaxAge is defined in seconds.
	//
	// Optional. Default: 0.
	maxAge int

	// When set to true, the server tries minimizing CPU usage by caching compressed files.
	//
	// Optional. Default: false
	compress bool

	// CompressedFileSuffixes adds suffix to the original file name and
	// tries saving the resulting compressed file under the new file name.
	//
	// Default: map[string]string{"gzip": ".fox.gz", "br": ".fox.br", "zstd": ".fox.zst"}
	compressedFileSuffix map[string]string

	// When set to true, enables byte range requests.
	//
	// Optional. Default: false
	byteRange bool

	// When set to true, enables directory browsing.
	//
	// Optional. Default: false.
	browse bool

	// When set to true, enables direct download.
	//
	// Optional. Default: false.
	download bool
}

type fsInstance struct {
	opts              *fsOptions
	cacheControlValue string
	handler           fasthttp.RequestHandler
}

func (f *fsOptions) compareOptions(o *fsOptions) bool {
	if o.fs != nil {
		return false
	}

	if f.root != o.root {
		return false
	}

	if f.compress != o.compress {
		return false
	}

	if f.byteRange != o.byteRange {
		return false
	}

	if f.download != o.download {
		return false
	}

	if f.cacheDuration != o.cacheDuration {
		return false
	}

	if f.maxAge != o.maxAge {
		return false
	}
	return true
}

func defaultFSOptions() *fsOptions {
	return &fsOptions{
		root:          ".",
		indexNames:    []string{"index.html"},
		cacheDuration: 10 * time.Second,
		compressedFileSuffix: map[string]string{
			"gzip": ".fox.gz",
			"br":   ".fox.br",
			"zstd": ".fox.zst",
		},
	}
}

// FsOption static options
type FsOption func(o *fsOptions)

// FSDownload set download option
func FSDownload(download bool) FsOption {
	return func(o *fsOptions) {
		o.download = download
	}
}

// FSRootPath file server root path
func FSRootPath(root string) FsOption {
	return func(o *fsOptions) {
		o.root = root
	}
}
