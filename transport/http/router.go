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
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/go-fox/fox/errors"
	"github.com/go-fox/fox/internal/bytesconv"
	"github.com/go-fox/fox/middleware"
	"github.com/go-fox/fox/transport"
)

// Router is router definition
type Router interface {
	Routes
	ServeHTTP(ctx *Context) error

	Connect(pattern string, handler Handler, middlewares ...Handler) Router
	Delete(pattern string, handler Handler, middlewares ...Handler) Router
	Get(pattern string, handler Handler, middlewares ...Handler) Router
	Head(pattern string, handler Handler, middlewares ...Handler) Router
	Options(pattern string, handler Handler, middlewares ...Handler) Router
	Patch(pattern string, handler Handler, middlewares ...Handler) Router
	Post(pattern string, handler Handler, middlewares ...Handler) Router
	Put(pattern string, handler Handler, middlewares ...Handler) Router
	Trace(pattern string, handler Handler, middlewares ...Handler) Router

	Static(pattern string, opts ...FsOption) Router

	Mount(pattern string, router Router) Router
	Group(path string, handlers ...Handler) Router
	Any(pattern string, handler Handler, middlewares ...Handler) Router
}

// Routes routers definition
type Routes interface {
	Routes() []Route
}

type router struct {
	prefix      string
	tree        *node
	middlewares []Handler
	srv         *Server
	parent      *router
}

// ServeHTTP serve HTTP
func (r *router) ServeHTTP(ctx *Context) error {
	routePath := ctx.routePath
	if routePath == "" {
		path := bytesconv.BytesToString(ctx.fastCtx.Path())
		if path != "" {
			routePath = path
		}
		if routePath == "" {
			routePath = "/"
		}
	}
	method, ok := methodMap[ctx.method]
	if !ok {
		return ctx.SetStatusCode(StatusBadRequest).SendString(default405Body)
	}
	if _, handler, middlewares := r.tree.FindRoute(ctx, method, routePath); handler != nil {
		ctx.index = -1
		ctx.handlers = append(middlewares, handler)
		tr, ok := transport.FromServerContext(ctx.Context())
		if !ok {
			pathTemplate := ctx.PathTemplate()
			v := &Transport{
				original:     ctx.Path(),
				operation:    pathTemplate,
				pathTemplate: pathTemplate,
				request:      ctx.Request(),
				response:     ctx.Response(),
			}
			if ctx.srv.endpoint != nil {
				v.endpoint = ctx.srv.endpoint.String()
			}
			ctx.SetContext(transport.NewServerContext(ctx.Context(), v))
		} else {
			pathTemplate := ctx.PathTemplate()
			v := tr.(*Transport)
			v.operation = pathTemplate
			v.pathTemplate = pathTemplate
		}
		return ctx.Next()
	}
	if ctx.methodNotAllowed {
		return ctx.SetStatusCode(StatusBadRequest).SendString(default405Body)
	}
	return ctx.SetStatusCode(StatusNotFound).SendString(default404Body)
}

// ServeFastHTTP fast http proxy
func (r *router) ServeFastHTTP(fastCtx *fasthttp.RequestCtx) {
	reqCtx := r.srv.acquireContext(fastCtx)
	defer r.srv.releaseContext(reqCtx)
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	conf := r.srv.Config()
	if r.srv.Config().Timeout > 0 {
		ctx, cancel = context.WithTimeout(reqCtx.Context(), conf.Timeout)
	} else {
		ctx, cancel = context.WithCancel(reqCtx.Context())
	}
	defer cancel()
	reqCtx.SetContext(ctx)
	if err := r.ServeHTTP(reqCtx); err != nil {
		if catch := conf.ene(reqCtx, err); catch != nil {
			_ = reqCtx.SendStatus(StatusInternalServerError)
		}
	}
}

func (s *Server) fastHTTPErrorHandler(fastCtx *fasthttp.RequestCtx, err error) {
	c := s.acquireContext(fastCtx)
	defer s.releaseContext(c)
	var errSmallBuffer *fasthttp.ErrSmallBuffer
	var opError *net.OpError
	if errors.As(err, &errSmallBuffer) {
		err = errors.RequestHeaderFieldsTooLarge("REQUEST_HEADER_FIELDS_TOO_LARGE", statusMessage[StatusRequestHeaderFieldsTooLarge])
	} else if errors.As(err, &opError) && opError.Timeout() {
		err = errors.RequestTimeout("REQUEST_TIMEOUT", statusMessage[StatusRequestTimeout])
	} else if errors.Is(err, fasthttp.ErrBodyTooLarge) {
		err = errors.RequestEntityTooLarge("BODY_TOO_LARGE", statusMessage[StatusRequestEntityTooLarge])
	} else if errors.Is(err, fasthttp.ErrGetOnly) {
		err = errors.MethodNotAllowed("METHOD_NOT_ALLOWED", statusMessage[StatusMethodNotAllowed])
	} else if strings.Contains(err.Error(), "timeout") {
		err = errors.RequestTimeout("REQUEST_TIMEOUT", statusMessage[StatusRequestTimeout])
	} else {
		err = errors.BadRequest("BAD_REQUEST", statusMessage[StatusBadRequest])
	}
	if catch := s.config.ene(c, err); catch != nil {
		_ = c.SendStatus(StatusInternalServerError)
	}
}

// Use registers a middleware route that will match requests
// with the provided prefix (which is optional and defaults to "/").
// Also, you can pass another app instance as a sub-router along a routing path.
// It's very useful to split up a large API as many independent routers and
// compose them as a single service using Use. The fiber's error handler and
// any of the fiber's sub apps are added to the application's error handlers
// to be invoked on errors that happen within the prefix route.
//
//		app.Use(func(c http.Ctx) error {
//		     return c.Next()
//		})
//		app.Use("/api", func(c http.Ctx) error {
//		     return c.Next()
//		})
//		app.Use("/api", handler, func(c http.Ctx) error {
//		     return c.Next()
//		})
//	 	subRoute := http.NewServeMux()
//		app.Use("/mounted-path", subRoute)
//
// This method will match all HTTP verbs: GET, POST, PUT, HEAD etc...
func (r *router) Use(args ...any) Router {
	var (
		subRouter   Router
		prefix      string
		prefixes    []string
		handlers    []Handler
		middlewares []middleware.Middleware
	)
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			prefix = arg
		case Router:
			subRouter = arg
		case []string:
			prefixes = arg
		case Handler:
			handlers = append(handlers, arg)
		case middleware.Middleware:
			middlewares = append(middlewares, arg)
		default:
			panic(fmt.Sprintf("use: invalid handler %v\n", reflect.TypeOf(arg)))
		}
	}
	if len(prefixes) == 0 {
		prefixes = append(prefixes, prefix)
	}
	for _, s := range prefixes {
		if subRouter != nil {
			r.mount(prefix, subRouter)
			return r
		}
		r.handler(mALL, s, nil, handlers...)
	}
	return r
}

// Routes get routers
func (r *router) Routes() []Route {
	return r.tree.routers()
}

// Group returns a new router group.
func (r *router) Group(prefix string, handlers ...Handler) Router {
	prefix = r.clearPath(prefix)
	mws := make([]Handler, len(r.middlewares))
	copy(mws, r.middlewares)
	mws = append(mws, handlers...)
	return &router{
		prefix:      prefix,
		tree:        r.tree,
		middlewares: mws,
		srv:         r.srv,
	}
}

// Any register the handler on all HTTP methods
func (r *router) Any(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mALL, pattern, handler, middlewares...)
	return r
}

// Connect registers a route for CONNECT methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Connect(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mCONNECT, pattern, handler, middlewares...)
	return r
}

// Delete registers a route for DELETE methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Delete(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mDELETE, pattern, handler, middlewares...)
	return r
}

// Get registers a route for GET methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Get(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mGET, pattern, handler, middlewares...)
	return r
}

// Head registers a route for HEAD methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Head(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mHEAD, pattern, handler, middlewares...)
	return r
}

// Options register a route for OPTIONS methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Options(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mOPTIONS, pattern, handler, middlewares...)
	return r
}

// Patch registers a route for PATCH methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Patch(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mPATCH, pattern, handler, middlewares...)
	return r
}

// Post registers a route for POST methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Post(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mPOST, pattern, handler, middlewares...)
	return r
}

// Put registers a route for PUT methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Put(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mPUT, pattern, handler, middlewares...)
	return r
}

// Trace registers a route for TRACE methods that is used to submit an entity to the
// specified resource, often causing a change in state or side effects on the server.
func (r *router) Trace(pattern string, handler Handler, middlewares ...Handler) Router {
	r.handler(mTRACE, pattern, handler, middlewares...)
	return r
}

// Static will create a file server serving static files
func (r *router) Static(pattern string, opts ...FsOption) Router {
	handler := fileServe(opts...)
	r.Head(pattern, handler)
	r.Get(pattern, handler)
	return r
}

// Mount attaches another Router instance as a sub-router along a routing path.
func (r *router) Mount(pattern string, router Router) Router {
	r.mount(pattern, router)
	return r
}

func (r *router) getSever() *Server {
	if r.parent != nil {
		return r.parent.getSever()
	}
	return r.srv
}

func (r *router) mount(pattern string, mux Router) {
	if mux == nil {
		panic(fmt.Sprintf("[HTPP]: attempting to Mount() a nil Router on '%s'", pattern))
	}
	// Provide runtime safety for ensuring a pattern isn't mounted on an existing
	// routing pattern.
	if r.tree.findPattern(pattern+"*") || r.tree.findPattern(pattern+"/*") {
		panic(fmt.Sprintf("[HTTP]: attempting to Mount() a Router on an existing path, '%s'", pattern))
	}

	mountHandler := Handler(func(ctx *Context) error {
		ctx.routePath = r.nextRoutePath(ctx)
		// reset the wildcard URLParam which connects the subrouter
		n := len(ctx.urlParams.keys) - 1
		if n >= 0 && ctx.urlParams.keys[n] == "*" && len(ctx.urlParams.values) > n {
			ctx.urlParams.remove(n)
		}
		return mux.ServeHTTP(ctx)
	})

	if pattern == "" || pattern[len(pattern)-1] != '/' {
		r.handler(mALL|mSTUB, pattern, mountHandler)
		r.handler(mALL|mSTUB, pattern+"/", mountHandler)
		pattern += "/"
	}

	switch v := mux.(type) {
	case *Server:
		v.parent = r
	case *router:
		v.parent = r
	}

	method := mALL
	subRouter := mux

	if subRouter == nil {
		method |= mSTUB
	}
	n := r.handler(method, pattern+"*", mountHandler)
	if subRouter != nil {
		n.subroutes = subRouter
	}
}

func (r *router) nextRoutePath(ctx *Context) string {
	routePath := "/"
	nx := len(ctx.routeParams.keys) - 1 // index of last param in list
	if nx >= 0 && ctx.routeParams.keys[nx] == "*" && len(ctx.routeParams.values) > nx {
		routePath = "/" + ctx.routeParams.values[nx]
	}
	return routePath
}

func (r *router) handler(method methodType, pattern string, handler Handler, middlewares ...Handler) *node {
	pattern = r.clearPath(pattern)
	if len(r.middlewares) > 0 {
		middlewares = append(r.middlewares, middlewares...)
	}
	return r.tree.AddRoute(method, pattern, handler, middlewares...)
}

func (r *router) clearPath(path string) string {
	if len(path) == 0 {
		return r.prefix
	}
	if path[0] != '/' {
		path = "/" + path
	}
	return strings.TrimRight(r.prefix, "/") + path
}

// Deadline impl context.Context
func (ctx *Context) Deadline() (deadline time.Time, ok bool) {
	if ctx.fastCtx == nil {
		return time.Time{}, false
	}
	return ctx.Context().Deadline()
}

// Done impl context.Context
func (ctx *Context) Done() <-chan struct{} {
	if ctx.fastCtx == nil {
		return nil
	}
	return ctx.Context().Done()
}

// Err impl context.Context
func (ctx *Context) Err() error {
	if ctx.fastCtx == nil {
		return context.Canceled
	}
	return ctx.Context().Err()
}

// WithValue impl context.Context
func (ctx *Context) WithValue(key, value any) {
	parant := ctx.Context()
	child := context.WithValue(parant, key, value)
	ctx.SetContext(child)
}

// Value impl context.Context
func (ctx *Context) Value(key any) any {
	if ctx.fastCtx == nil {
		return nil
	}
	return ctx.Context().Value(key)
}
