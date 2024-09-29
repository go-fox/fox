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
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-fox/sugar/container/spool"
	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/json"
	"github.com/go-fox/fox/codec/xml"
	"github.com/go-fox/fox/errors"
	"github.com/go-fox/fox/internal/bytesconv"
	"github.com/go-fox/fox/internal/httputil"
	"github.com/go-fox/fox/internal/matcher"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/transport"
)

var ctxPool = spool.New[*Context](func() *Context {
	return new(Context)
}, func(ctx *Context) {
	ctx.reset()
})

// Context http request context
type Context struct {
	index            int8
	fastCtx          *fasthttp.RequestCtx
	handlers         HandlersChain
	middleware       matcher.Matcher
	method           string
	routePath        string
	routePattern     string
	routePatterns    []string
	srv              *Server
	methodsAllowed   []methodType // allowed methods in case of a 405
	methodNotAllowed bool
	methodType       methodType
	originalPath     string
	routeParams      *RouteParams
	urlParams        *RouteParams
	pathTemplate     string
}

func (s *Server) acquireContext(fastCtx *fasthttp.RequestCtx) *Context {
	ctx := ctxPool.Get()
	ctx.reset()
	ctx.fastCtx = fastCtx
	ctx.srv = s
	ctx.middleware = s.Config().middlewares
	ctx.method = bytesconv.BytesToString(fastCtx.Method())
	ctx.originalPath = bytesconv.BytesToString(fastCtx.URI().PathOriginal())
	m, ok := methodMap[ctx.method]
	if ok {
		ctx.methodType = m
	}
	ctx.SetContext(context.Background())
	ctx.WithValue(httpCtxKey{}, ctx)
	return ctx
}

func (s *Server) releaseContext(c *Context) {
	ctxPool.Put(c)
}

func (ctx *Context) reset() {
	ctx.index = -1
	ctx.fastCtx = nil
	ctx.middleware = nil
	ctx.srv = nil
	ctx.method = ""
	ctx.routePath = ""
	ctx.routePattern = ""
	ctx.routePatterns = []string{}
	ctx.methodsAllowed = []methodType{}
	ctx.handlers = HandlersChain{}
	ctx.routeParams = &RouteParams{}
	ctx.urlParams = &RouteParams{}
}

// Next execute the next method in the stack that matches the current route
func (ctx *Context) Next() (err error) {
	ctx.index++
	if ctx.index < int8(len(ctx.handlers)) {
		err = ctx.handlers[ctx.index](ctx)
	}
	return err
}

// Context get user context
func (ctx *Context) Context() context.Context {
	return ctx.fastCtx.Value(userContextKey{}).(context.Context)
}

// SetContext set context
func (ctx *Context) SetContext(c context.Context) {
	ctx.fastCtx.SetUserValue(userContextKey{}, c)
}

// Middleware executes middlewares
func (ctx *Context) Middleware(h middleware.Handler) middleware.Handler {
	tr, ok := transport.FromServerContext(ctx.Context())
	if ok {
		return middleware.Chain(ctx.middleware.Match(tr.Operation())...)(h)
	}
	return middleware.Chain(ctx.middleware.Match(ctx.PathTemplate())...)(h)
}

// Method get request method
func (ctx *Context) Method() string {
	return ctx.method
}

// Param get path param
func (ctx *Context) Param(key string) string {
	v, _ := ctx.urlParams.Get(key)
	return v
}

// Params get all path param
func (ctx *Context) Params() *RouteParams {
	return ctx.urlParams
}

// ShouldBind binding struct
func (ctx *Context) ShouldBind(out interface{}) error {
	req := warpRequest(ctx.Request())
	defer releaseBindRequest(req)
	return defaultBinding.Unmarshal(out, req, ctx.urlParams)
}

// MultipartForm get all multipart.Form
func (ctx *Context) MultipartForm() (*multipart.Form, error) {
	return ctx.fastCtx.MultipartForm()
}

// montageRoutePatterns get route template
func (ctx *Context) montageRoutePatterns() string {
	routePattern := strings.Join(ctx.routePatterns, "")
	routePattern = replaceWildcards(routePattern)
	if routePattern != "/" {
		routePattern = strings.TrimSuffix(routePattern, "//")
		routePattern = strings.TrimSuffix(routePattern, "/")
	}
	return routePattern
}

// Path get path
func (ctx *Context) Path() string {
	return ctx.originalPath
}

// ContentType get request Content-Type
func (ctx *Context) ContentType() string {
	return bytesconv.BytesToString(ctx.Response().Header.ContentType())
}

// PathTemplate return route path template
func (ctx *Context) PathTemplate() string {
	return ctx.pathTemplate
}

// replaceWildcards takes a route pattern and recursively replaces all
// occurrences of "/*/" to "/".
func replaceWildcards(p string) string {
	if strings.Contains(p, "/*/") {
		return replaceWildcards(strings.Replace(p, "/*/", "/", -1))
	}
	return p
}

// Send http write bytes to body
func (ctx *Context) Send(p []byte) {
	ctx.fastCtx.Response.AppendBody(p)
}

// FastCtx get fast http context
func (ctx *Context) FastCtx() *fasthttp.RequestCtx {
	return ctx.fastCtx
}

// Request return fasthttp.Request
func (ctx *Context) Request() *Request {
	return &ctx.fastCtx.Request
}

// Response return fasthttp.Response
func (ctx *Context) Response() *fasthttp.Response {
	return &ctx.fastCtx.Response
}

// RequestBodyStream get request body stream
func (ctx *Context) RequestBodyStream() io.Reader {
	return ctx.fastCtx.RequestBodyStream()
}

// GetRequestHeader get request header
func (ctx *Context) GetRequestHeader(key string) string {
	return bytesconv.BytesToString(ctx.Request().Header.Peek(key))
}

// GetResponseHeader get response header
func (ctx *Context) GetResponseHeader(key string) string {
	return bytesconv.BytesToString(ctx.Response().Header.Peek(key))
}

// SetStatusCode set status code
func (ctx *Context) SetStatusCode(code int) *Context {
	ctx.Response().SetStatusCode(code)
	return ctx
}

// SetResponseHeader set http response header
func (ctx *Context) SetResponseHeader(key string, value string) {
	ctx.Response().Header.Set(key, value)
}

// Redirect redirect rul
func (ctx *Context) Redirect(statusCode int, uri []byte) error {
	ctx.redirect(uri, statusCode)
	return nil
}

// redirect redirect url
func (ctx *Context) redirect(uri []byte, statusCode int) {
	ctx.Response().Header.SetCanonical(bytesconv.StringToBytes(HeaderLocation), uri)
	statusCode = ctx.getRedirectStatusCode(statusCode)
	ctx.Response().SetStatusCode(statusCode)
}

// getRedirectStatusCode get redirect status code
func (ctx *Context) getRedirectStatusCode(statusCode int) int {
	if statusCode == StatusMovedPermanently || statusCode == StatusFound ||
		statusCode == StatusSeeOther || statusCode == StatusTemporaryRedirect ||
		statusCode == StatusPermanentRedirect {
		return statusCode
	}
	return StatusFound
}

// SendString send http body
func (ctx *Context) SendString(body string) error {
	ctx.fastCtx.SetBodyString(body)
	return nil
}

// SendStatus send http status
func (ctx *Context) SendStatus(status int) error {
	ctx.Response().SetStatusCode(status)
	return nil
}

// Returns send any data
func (ctx *Context) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return ctx.srv.Config().enc(ctx, v)
}

// Result send any data
func (ctx *Context) Result(code int, v interface{}) error {
	ctx.SetStatusCode(code)
	return ctx.srv.Config().enc(ctx, v)
}

// OkResult send data with a status of 200
func (ctx *Context) OkResult(v any) error {
	ctx.SetStatusCode(http.StatusOK)
	return ctx.srv.Config().enc(ctx, v)
}

// JSON wire json data
func (ctx *Context) JSON(code int, v interface{}) error {
	ctx.SetResponseHeader("Content-Type", "application/json")
	ctx.SetStatusCode(code)
	body, err := codec.GetCodec(json.Name).Marshal(v)
	if err != nil {
		return err
	}
	ctx.Send(body)
	return nil
}

// XML wire xml data
func (ctx *Context) XML(code int, v interface{}) error {
	ctx.SetResponseHeader("Content-Type", "application/xml")
	ctx.SetStatusCode(code)
	body, err := codec.GetCodec(xml.Name).Marshal(v)
	if err != nil {
		return err
	}
	ctx.Send(body)
	return nil
}

// SendFile send file
func (ctx *Context) SendFile(file string, opts ...FsOption) error {
	o := defaultFSOptions()
	for _, opt := range opts {
		opt(o)
	}
	filename := file
	if o.cacheDuration == 0 {
		o.cacheDuration = 10 * time.Second
	}
	var fsHandler fasthttp.RequestHandler
	var cacheControlValue string

	ctx.srv.fsInstanceMux.Lock()
	for _, instance := range ctx.srv.fsInstances {
		if instance.opts.compareOptions(o) {
			fsHandler = instance.handler
			cacheControlValue = instance.cacheControlValue
		}
	}
	ctx.srv.fsInstanceMux.Unlock()

	if fsHandler == nil {
		fasthttpFS := &fasthttp.FS{
			Root:                   "",
			FS:                     o.fs,
			AllowEmptyRoot:         true,
			GenerateIndexPages:     false,
			AcceptByteRange:        o.byteRange,
			Compress:               o.compress,
			CompressBrotli:         o.compress,
			CompressedFileSuffixes: o.compressedFileSuffix,
			CacheDuration:          o.cacheDuration,
			SkipCache:              o.cacheDuration < 0,
			IndexNames:             []string{"index.html"},
			PathNotFound: func(ctx *fasthttp.RequestCtx) {
				ctx.Response.SetStatusCode(StatusNotFound)
			},
		}

		if o.fs != nil {
			fasthttpFS.Root = "."
		}

		sf := &fsInstance{
			opts:    o,
			handler: fasthttpFS.NewRequestHandler(),
		}

		maxAge := o.maxAge
		if maxAge > 0 {
			sf.cacheControlValue = "public, max-age=" + strconv.Itoa(maxAge)
		}

		// set vars
		fsHandler = sf.handler
		cacheControlValue = sf.cacheControlValue

		ctx.srv.fsInstanceMux.Lock()
		ctx.srv.fsInstances = append(ctx.srv.fsInstances, sf)
		ctx.srv.fsInstanceMux.Unlock()
	}

	// Delete the Accept-Encoding header if compression is disabled
	if !o.compress {
		// https://github.com/valyala/fasthttp/blob/7cc6f4c513f9e0d3686142e0a1a5aa2f76b3194a/fs.go#L55
		ctx.fastCtx.Request.Header.Del(HeaderAcceptEncoding)
	}

	// copy of https://github.com/valyala/fasthttp/blob/7cc6f4c513f9e0d3686142e0a1a5aa2f76b3194a/fs.go#L103-L121 with small adjustments
	if len(file) == 0 || (!filepath.IsAbs(file) && o.fs == nil) {
		// extend relative path to absolute path
		hasTrailingSlash := len(file) > 0 && (file[len(file)-1] == '/' || file[len(file)-1] == '\\')

		var err error
		file = filepath.FromSlash(file)
		if file, err = filepath.Abs(file); err != nil {
			return errors.InternalServer("FILEPATH_ABS", "failed to determine abs file path").WithCause(err)
		}
		if hasTrailingSlash {
			file += "/"
		}
	}

	// convert the path to forward slashes regardless the OS in order to set the URI properly
	// the handler will convert back to OS path separator before opening the file
	file = filepath.ToSlash(file)

	// Restore the original requested URL
	originalURL := bytesconv.StringToBytes(ctx.originalPath)
	defer ctx.fastCtx.Request.SetRequestURI(string(originalURL))

	// Set new URI for fileHandler
	ctx.fastCtx.Request.SetRequestURI(file)

	// Save status code
	status := ctx.fastCtx.Response.StatusCode()

	// Serve file
	fsHandler(ctx.fastCtx)

	// Sets the response Content-Disposition header to attachment if the Download option is true
	if o.download {
		ctx.Attachment()
	}

	// Get the status code which is set by fasthttp
	fsStatus := ctx.fastCtx.Response.StatusCode()

	// Check for error
	if status != StatusNotFound && fsStatus == StatusNotFound {
		return errors.NotFound("NOT_FOUND_FILE", fmt.Sprintf("file %s not found", filename))
	}

	// Set the status code set by the user if it is different from the fasthttp status code and 200
	if status != fsStatus && status != StatusOK {
		ctx.SetStatusCode(status)
	}

	// Apply cache control header
	if status != StatusNotFound && status != StatusForbidden {
		if len(cacheControlValue) > 0 {
			ctx.fastCtx.Response.Header.Set(HeaderCacheControl, cacheControlValue)
		}
		return nil
	}

	return nil
}

// Cookie get cookie value
func (ctx *Context) Cookie(key string) string {
	return bytesconv.BytesToString(ctx.Request().Header.Cookie(key))
}

// Type sets the Content-Type HTTP header to the MIME type specified by the file extension.
func (ctx *Context) Type(extension string, charset ...string) *Context {
	if len(charset) > 0 {
		ctx.fastCtx.Response.Header.SetContentType(httputil.GetMIME(extension) + "; charset=" + charset[0])
	} else {
		ctx.fastCtx.Response.Header.SetContentType(httputil.GetMIME(extension))
	}
	return ctx
}

// SetCookie 设置cookie
func (ctx *Context) SetCookie(name, value string, opts ...CookieOption) *Context {
	ck := fasthttp.AcquireCookie()
	defer fasthttp.ReleaseCookie(ck)
	for _, opt := range opts {
		opt(ck)
	}
	ck.SetKey(name)
	ck.SetValue(value)
	ctx.Response().Header.SetCookie(ck)
	return ctx
}

// IsTLS is TLS
func (ctx *Context) IsTLS() bool {
	return ctx.fastCtx.IsTLS()
}

// Attachment sets the HTTP response Content-Disposition header field to attachment.
func (ctx *Context) Attachment(filename ...string) {
	if len(filename) > 0 {
		fname := filepath.Base(filename[0])
		ctx.Type(filepath.Ext(fname))

		ctx.SetCanonical(HeaderContentDisposition, `attachment; filename="`+fname+`"`)
		return
	}
	ctx.SetCanonical(HeaderContentDisposition, "attachment")
}

// SetCanonical set Content-Disposition
func (ctx *Context) SetCanonical(key, val string) {
	ctx.fastCtx.Response.Header.SetCanonical(bytesconv.StringToBytes(key), bytesconv.StringToBytes(val))
}
