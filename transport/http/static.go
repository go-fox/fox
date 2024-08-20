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

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/internal/bytesconv"
)

// fileServe file serve handler
func fileServe(opts ...FsOption) Handler {
	o := defaultFSOptions()
	for _, opt := range opts {
		opt(o)
	}
	var root = o.root
	var createFS sync.Once
	var fileHandler fasthttp.RequestHandler
	var cacheControlValue string
	// adjustments for io/fs compatibility
	if o.fs != nil && root == "" {
		root = "."
	}
	return func(ctx *Context) error {
		if o.next != nil && o.next(ctx) {
			return ctx.Next()
		}
		method := ctx.Method()
		if method != MethodHead && method != MethodGet {
			return ctx.Next()
		}
		// Initialize FS
		createFS.Do(func() {
			prefix := ctx.PathTemplate()

			// Is prefix a partial wildcard?
			if strings.Contains(prefix, "*") {
				// /john* -> /john
				prefix = strings.Split(prefix, "*")[0]
			}

			prefixLen := len(prefix)
			if prefixLen > 1 && prefix[prefixLen-1:] == "/" {
				// /john/ -> /john
				prefixLen--
			}

			fs := &fasthttp.FS{
				Root:                   root,
				FS:                     o.fs,
				AllowEmptyRoot:         true,
				GenerateIndexPages:     o.browse,
				AcceptByteRange:        o.byteRange,
				Compress:               o.compress,
				CompressBrotli:         o.compress, // Brotli compression won't work without this
				CompressedFileSuffixes: o.compressedFileSuffix,
				CacheDuration:          o.cacheDuration,
				SkipCache:              o.cacheDuration < 0,
				IndexNames:             o.indexNames,
				PathNotFound: func(fctx *fasthttp.RequestCtx) {
					fctx.Response.SetStatusCode(StatusNotFound)
				},
			}

			fs.PathRewrite = func(fctx *fasthttp.RequestCtx) []byte {
				path := fctx.Path()

				if len(path) >= prefixLen {
					checkFile, err := isFile(root, fs.FS)
					if err != nil {
						return path
					}

					// If the root is a file, we need to reset the path to "/" always.
					switch {
					case checkFile && fs.FS == nil:
						path = []byte("/")
					case checkFile && fs.FS != nil:
						path = bytesconv.StringToBytes(root)
					default:
						path = path[prefixLen:]
						if len(path) == 0 || path[len(path)-1] != '/' {
							path = append(path, '/')
						}
					}
				}

				if len(path) > 0 && path[0] != '/' {
					path = append([]byte("/"), path...)
				}

				return path
			}

			maxAge := o.maxAge
			if maxAge > 0 {
				cacheControlValue = "public, max-age=" + strconv.Itoa(maxAge)
			}
			fileHandler = fs.NewRequestHandler()
		})

		// Serve file
		fileHandler(ctx.FastCtx())

		// Sets the response Content-Disposition header to attachment if the Download option is true
		if o.download {
			ctx.Attachment()
		}

		// Return request if found and not forbidden
		status := ctx.Response().StatusCode()

		if status != StatusNotFound && status != StatusForbidden {
			if len(cacheControlValue) > 0 {
				ctx.Response().Header.Set(HeaderCacheControl, cacheControlValue)
			}

			if o.modifyResponse != nil {
				return o.modifyResponse(ctx)
			}

			return nil
		}

		// Return custom 404 handler if provided.
		if o.notFoundHandler != nil {
			return o.notFoundHandler(ctx)
		}

		// Reset response to default
		ctx.FastCtx().SetContentType("") // Issue #420
		ctx.FastCtx().Response.SetStatusCode(StatusOK)
		ctx.FastCtx().Response.SetBodyString("")

		// Next middleware
		return ctx.Next()
	}
}

// isFile checks if the root is a file.
func isFile(root string, filesystem fs.FS) (bool, error) {
	var file fs.File
	var err error

	if filesystem != nil {
		file, err = filesystem.Open(root)
		if err != nil {
			return false, fmt.Errorf("static: %w", err)
		}
	} else {
		file, err = os.Open(filepath.Clean(root))
		if err != nil {
			return false, fmt.Errorf("static: %w", err)
		}
	}

	stat, err := file.Stat()
	if err != nil {
		return false, fmt.Errorf("static: %w", err)
	}

	return stat.Mode().IsRegular(), nil
}
