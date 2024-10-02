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
	"regexp"
	"sort"
	"strings"
)

type (
	methodType uint
	nodeType   uint8
	node       struct {
		parent *node
		// node type: static, regexp, param, catchAll
		nodeType nodeType
		// first byte of the child prefix
		tail byte
		// prefix is the common prefix we ignore
		prefix string
		// first byte of the prefix
		label byte
		// regexp matcher for regexp nodes
		rex       *regexp.Regexp
		endpoints endpoints
		children  [ntCatchAll + 1]nodes
		subroutes Routes
	}
	nodes    []*node
	endpoint struct {
		pattern    string
		paramKeys  []string
		handler    Handler
		middleware HandlersChain
	}
	endpoints map[methodType]*endpoint
	segment   struct {
		nodeType   nodeType
		paramKey   string
		regexp     string
		tail       byte
		startIndex int
		endIndex   int
	}
)

const (
	mSTUB methodType = 1 << iota
	mCONNECT
	mDELETE
	mGET
	mHEAD
	mOPTIONS
	mPATCH
	mPOST
	mPUT
	mTRACE
)

var (
	mALL = mCONNECT | mDELETE | mGET | mHEAD |
		mOPTIONS | mPATCH | mPOST | mPUT | mTRACE
	methodMap = map[string]methodType{
		MethodConnect: mCONNECT,
		MethodDelete:  mDELETE,
		MethodGet:     mGET,
		MethodHead:    mHEAD,
		MethodOptions: mOPTIONS,
		MethodPatch:   mPATCH,
		MethodPost:    mPOST,
		MethodPut:     mPUT,
		MethodTrace:   mTRACE,
		MethodAny:     mALL,
	}
)

const (
	ntStatic   nodeType = iota // /home
	ntRegexp                   // /{id:[0-9]+}
	ntParam                    // /{user}
	ntCatchAll                 // /api/v1/*
)

func (r endpoints) Value(method methodType) *endpoint {
	mh, ok := r[method]
	if !ok {
		mh = &endpoint{}
		r[method] = mh
	}
	return mh
}

// AddRoute add route
func (n *node) AddRoute(method methodType, path string, handler Handler, middleware ...Handler) *node {
	var parent *node
	search := path
	for {
		if len(search) == 0 {
			n.setEndpoint(method, path, handler, middleware...)
			return n
		}
		// We're going to be searching for a wild node next,
		// in this case, we need to get the tail
		var label = search[0]
		var seg segment
		if label == '{' || label == '*' {
			seg = n.nextSegment(search)
		}
		var prefix string
		if seg.nodeType == ntRegexp {
			prefix = seg.regexp
		}

		// Look for the edge to attach to
		parent = n
		n = n.getEdge(seg.nodeType, label, seg.tail, prefix)

		if n == nil {
			child := &node{label: label, tail: seg.tail, prefix: search}
			hn := parent.addChild(child, search)
			hn.setEndpoint(method, path, handler, middleware...)
			return hn
		}

		if n.nodeType > ntStatic {
			// We found a param node, trim the param from the search path and continue.
			// This param/wild pattern segment would already be on the tree from a previous
			// call to addChild when creating a new node.
			search = search[seg.endIndex:]
			continue
		}

		// Static nodes fall below here.
		// Determine longest prefix of the search key on match.
		commonPrefix := longestPrefix(search, n.prefix)
		if commonPrefix == len(n.prefix) {
			// the common prefix is as long as the current node's prefix we're attempting to insert.
			// keep the search going.
			search = search[commonPrefix:]
			continue
		}

		// Split the node
		child := &node{
			nodeType: ntStatic,
			prefix:   search[:commonPrefix],
		}
		parent.replaceChild(search[0], seg.tail, child)

		// Restore the existing node
		n.label = n.prefix[commonPrefix]
		n.prefix = n.prefix[commonPrefix:]
		child.addChild(n, n.prefix)

		// If the new key is a subset, set the method/handler on this node and finish.
		search = search[commonPrefix:]
		if len(search) == 0 {
			child.setEndpoint(method, path, handler, middleware...)
			return child
		}

		// Create a new edge for the node
		subChild := &node{
			nodeType: ntStatic,
			label:    search[0],
			prefix:   search,
		}
		hn := child.addChild(subChild, search)
		hn.setEndpoint(method, path, handler, middleware...)
		return hn
	}
}

// FindRoute find route
func (n *node) FindRoute(ctx *Context, method methodType, path string) (*node, Handler, HandlersChain) {
	// reset data
	ctx.routePath = ""
	ctx.routeParams.Reset()

	rn, mws := n.findRoute(ctx, method, path)
	if rn == nil {
		return nil, nil, mws
	}

	// append params
	ctx.urlParams.keys = append(ctx.urlParams.keys, ctx.routeParams.keys...)
	ctx.urlParams.values = append(ctx.urlParams.values, ctx.routeParams.values...)

	// Record the routing pattern in the request lifecycle
	if rn.endpoints[method].pattern != "" {
		ctx.routePattern = rn.endpoints[method].pattern
		ctx.routePatterns = append(ctx.routePatterns, ctx.routePattern)
		ctx.pathTemplate = ctx.montageRoutePatterns()
	}
	return rn, rn.endpoints.Value(method).handler, append(mws, getMiddleware(rn, method)...)
}

func getMiddleware(cur *node, method methodType) HandlersChain {
	middlewares := HandlersChain{}
	if cur.endpoints != nil {
		middlewares = append(middlewares, cur.endpoints.Value(method).middleware...)
	}
	if cur.parent != nil {
		middlewares = append(middlewares, getMiddleware(cur.parent, method)...)
	}
	return middlewares
}

// findRoute 查找router，并返回路径上中间件
func (n *node) findRoute(ctx *Context, method methodType, path string) (*node, HandlersChain) {
	mws := HandlersChain{}
	curNode := n
	curPath := path
	for t, ns := range curNode.children {
		nType := nodeType(t)
		if len(ns) == 0 {
			continue
		}
		var (
			cur    *node
			label  byte
			search = curPath
		)

		if curPath != "" {
			label = curPath[0]
		}

		switch nType {
		case ntStatic:
			cur = ns.findEdge(label)
			if cur == nil || !strings.HasPrefix(search, cur.prefix) {
				continue
			}
			search = search[len(cur.prefix):]
			mws = append(mws, cur.endpoints.Value(method).middleware...) // 匹配上了就累加
		case ntRegexp, ntParam:
			if search == "" {
				continue
			}
			for i := 0; i < ns.Len(); i++ {
				cur = ns[i]
				index := strings.IndexByte(search, cur.tail)
				if index == -1 {
					if cur.tail == '/' {
						index = len(search)
					} else {
						continue
					}
				} else if nType == ntRegexp && index == 0 {
					continue
				}

				if nType == ntRegexp && cur.rex != nil {
					if !cur.rex.MatchString(search[:index]) {
						continue
					}
				} else if strings.IndexByte(search[:index], '/') != -1 {
					// avoid a match across path segments
					continue
				}

				preValLen := len(ctx.routeParams.values)
				ctx.routeParams.values = append(ctx.routeParams.values, search[:index])
				search = search[index:]

				if len(search) == 0 {
					if cur.isLeaf() {
						if e := cur.endpoints[method]; e != nil && e.handler != nil {
							ctx.routeParams.keys = append(ctx.routeParams.keys, e.paramKeys...)
							return cur, mws
						}
					}
				}

				if find, curMws := cur.findRoute(ctx, method, search); find != nil {
					return find, append(mws, curMws...)
				}

				ctx.routeParams.values = ctx.routeParams.values[:preValLen]
				search = curPath
			}

			ctx.routeParams.values = append(ctx.routeParams.values, "")
		default:
			ctx.routeParams.values = append(ctx.routeParams.values, curPath)
			cur = ns[0]
			search = ""
		}

		if cur == nil {
			continue
		}

		if len(search) == 0 {
			if cur.isLeaf() {
				if e := cur.endpoints[method]; e != nil && e.handler != nil {
					ctx.routeParams.keys = append(ctx.routeParams.keys, e.paramKeys...)
					return cur, mws
				}
				for e := range cur.endpoints {
					if e == mALL || e == mSTUB {
						continue
					}
					ctx.methodsAllowed = append(ctx.methodsAllowed, e)
				}

				// flag that the routing context found a route, but not a corresponding
				// supported method
				ctx.methodNotAllowed = true
			}
		}

		// recursively find the next node..
		if find, curMws := cur.findRoute(ctx, method, search); find != nil {
			return find, append(mws, curMws...)
		}

		// Did not find the final handler, let's remove the param here if it was set
		if cur.nodeType > ntStatic {
			if len(ctx.routeParams.values) > 0 {
				ctx.routeParams.values = ctx.routeParams.values[:len(ctx.routeParams.values)-1]
			}
		}
	}

	return nil, nil
}

func (n *node) findPattern(pattern string) bool {
	nn := n
	for _, nds := range nn.children {
		if len(nds) == 0 {
			continue
		}

		n = nn.findEdge(nds[0].nodeType, pattern[0])
		if n == nil {
			continue
		}

		var idx int
		var xpattern string

		switch n.nodeType {
		case ntStatic:
			idx = longestPrefix(pattern, n.prefix)
			if idx < len(n.prefix) {
				continue
			}

		case ntParam, ntRegexp:
			idx = strings.IndexByte(pattern, '}') + 1

		case ntCatchAll:
			idx = longestPrefix(pattern, "*")

		default:
			panic("chi: unknown node type")
		}

		xpattern = pattern[idx:]
		if len(xpattern) == 0 {
			return true
		}

		return n.findPattern(xpattern)
	}
	return false
}

func (n *node) setEndpoint(method methodType, pattern string, handler Handler, middlewares ...Handler) {
	if n.endpoints == nil {
		n.endpoints = make(endpoints)
	}
	paramKeys := n.getParamKeys(pattern)
	if method&mSTUB == mSTUB {
		r := n.endpoints.Value(mSTUB)
		if handler != nil {
			r.handler = handler
		}
		r.middleware = append(r.middleware, middlewares...)
	}
	if method&mALL == mALL {
		for _, m := range methodMap {
			r := n.endpoints.Value(m)
			r.pattern = pattern
			r.paramKeys = paramKeys
			if handler != nil {
				r.handler = handler
			}
			r.middleware = append(r.middleware, middlewares...)
		}
	} else {
		r := n.endpoints.Value(method)
		r.pattern = pattern
		r.paramKeys = paramKeys
		if handler != nil {
			r.handler = handler
		}
		r.middleware = append(r.middleware, middlewares...)
	}
}

func (n *node) addChild(child *node, prefix string) *node {
	search := prefix
	// set parent
	child.parent = n
	// handler leaf node added to the tree is the child.
	// this may be overridden later down the flow
	hn := child

	// Parse next segment
	seg := n.nextSegment(search)

	// Add child depending on next up segment
	switch seg.nodeType {

	case ntStatic:
		// Search prefix is all static (that is, has no params in path)
		// noop

	default:
		// Search prefix contains a param, regexp or wildcard

		if seg.nodeType == ntRegexp {
			rex, err := regexp.Compile(seg.regexp)
			if err != nil {
				panic(fmt.Sprintf("[HTTP]: invalid regexp pattern '%s' in route param", seg.regexp))
			}
			child.prefix = seg.regexp
			child.rex = rex
		}

		if seg.startIndex == 0 {
			// Route starts with a param
			child.nodeType = seg.nodeType

			if seg.nodeType == ntCatchAll {
				seg.startIndex = -1
			} else {
				seg.startIndex = seg.endIndex
			}
			if seg.startIndex < 0 {
				seg.startIndex = len(search)
			}
			child.tail = seg.tail // for params, we set the tail

			if seg.startIndex != len(search) {
				// add static edge for the remaining part, split the end.
				// its not possible to have adjacent param nodes, so its certainly
				// going to be a static node next.

				search = search[seg.startIndex:] // advance search position

				nn := &node{
					nodeType: ntStatic,
					label:    search[0],
					prefix:   search,
				}
				hn = child.addChild(nn, search)
			}

		} else if seg.startIndex > 0 {
			// Route has some param

			// starts with a static segment
			child.nodeType = ntStatic
			child.prefix = search[:seg.startIndex]
			child.rex = nil

			// add the param edge node
			search = search[seg.startIndex:]

			nn := &node{
				nodeType: seg.nodeType,
				label:    search[0],
				tail:     seg.tail,
			}
			hn = child.addChild(nn, search)

		}
	}

	n.children[child.nodeType] = append(n.children[child.nodeType], child)
	n.children[child.nodeType].Sort()
	return hn
}

func (n *node) getParamKeys(pattern string) []string {
	pat := pattern
	var paramKeys []string
	for {
		s := n.nextSegment(pat)
		if s.nodeType == ntStatic {
			return paramKeys
		}
		for i := 0; i < len(paramKeys); i++ {
			if paramKeys[i] == s.paramKey {
				panic(fmt.Sprintf("[HTTP]: routing pattern '%s' contains duplicate param key, '%s'", pattern, s.paramKey))
			}
		}
		paramKeys = append(paramKeys, s.paramKey)
		pat = pat[s.endIndex:]
	}
}

// patNextSegment returns the next segment details from a pattern:
// node type, param key, regexp string, param tail byte, param starting index, param ending index
func (n *node) nextSegment(pattern string) segment {
	ps := strings.Index(pattern, "{")
	ws := strings.Index(pattern, "*")

	if ps < 0 && ws < 0 {
		return segment{ntStatic, "", "", 0, 0, len(pattern)} // we return the entire thing
	}

	// Sanity check
	if ps >= 0 && ws >= 0 && ws < ps {
		panic("[HTTP]: wildcard '*' must be the last pattern in a route, otherwise use a '{param}'")
	}

	var tail byte = '/' // Default endpoint tail to / byte

	if ps >= 0 {
		// Param/Regexp pattern is next
		nt := ntParam

		// Read to closing } taking into account opens and closes in curl count (cc)
		cc := 0
		pe := ps
		for i, c := range pattern[ps:] {
			if c == '{' {
				cc++
			} else if c == '}' {
				cc--
				if cc == 0 {
					pe = ps + i
					break
				}
			}
		}
		if pe == ps {
			panic("chi: route param closing delimiter '}' is missing")
		}

		key := pattern[ps+1 : pe]
		pe++ // set end to next position

		if pe < len(pattern) {
			tail = pattern[pe]
		}

		var rexpat string
		if idx := strings.Index(key, ":"); idx >= 0 {
			nt = ntRegexp
			rexpat = key[idx+1:]
			key = key[:idx]
		}

		if len(rexpat) > 0 {
			if rexpat[0] != '^' {
				rexpat = "^" + rexpat
			}
			if rexpat[len(rexpat)-1] != '$' {
				rexpat += "$"
			}
		}

		return segment{nt, key, rexpat, tail, ps, pe}
	}

	// Wildcard pattern as finale
	if ws < len(pattern)-1 {
		panic("[HTTP]: wildcard '*' must be the last value in a route. trim trailing text or use a '{param}' instead")
	}
	return segment{ntCatchAll, "*", "", 0, ws, len(pattern)}
}

func (n *node) replaceChild(label, tail byte, child *node) {
	for i := 0; i < len(n.children[child.nodeType]); i++ {
		if n.children[child.nodeType][i].label == label && n.children[child.nodeType][i].tail == tail {
			n.children[child.nodeType][i] = child
			n.children[child.nodeType][i].label = label
			n.children[child.nodeType][i].tail = tail
			return
		}
	}
	panic("[HTTP]: replacing missing child")
}

func (n *node) getEdge(nType nodeType, label, tail byte, prefix string) *node {
	nds := n.children[nType]
	for i := 0; i < len(nds); i++ {
		if nds[i].label == label && nds[i].tail == tail {
			if nType == ntRegexp && nds[i].prefix != prefix {
				continue
			}
			return nds[i]
		}
	}
	return nil
}

func (n *node) findEdge(nType nodeType, label byte) *node {
	nds := n.children[nType]
	num := len(nds)
	idx := 0

	switch nType {
	case ntStatic, ntParam, ntRegexp:
		i, j := 0, num-1
		for i <= j {
			idx = i + (j-i)/2
			if label > nds[idx].label {
				i = idx + 1
			} else if label < nds[idx].label {
				j = idx - 1
			} else {
				i = num // breaks cond
			}
		}
		if nds[idx].label != label {
			return nil
		}
		return nds[idx]

	default: // catch all
		return nds[idx]
	}
}

func (n *node) isLeaf() bool {
	return n.endpoints != nil
}

func (n *node) routers() []Route {
	var rts []Route
	n.walk(func(eps endpoints, subroutes Routes) bool {
		if eps[mSTUB] != nil && eps[mSTUB].handler != nil && subroutes == nil {
			return false
		}

		// Group methodHandlers by unique patterns
		pats := make(map[string]endpoints)

		for mt, h := range eps {
			if h.pattern == "" {
				continue
			}
			p, ok := pats[h.pattern]
			if !ok {
				p = endpoints{}
				pats[h.pattern] = p
			}
			p[mt] = h
		}

		for p, mh := range pats {
			hs := make(map[string]MethodHandler)
			if mh[mALL] != nil && mh[mALL].handler != nil {
				hs["*"] = MethodHandler{
					Handler:     mh[mALL].handler,
					Middlewares: mh[mALL].middleware,
				}
			}

			for mt, h := range mh {
				if h.handler == nil {
					continue
				}
				m := methodTypString(mt)
				if m == "" {
					continue
				}
				hs[m] = MethodHandler{
					Handler:     h.handler,
					Middlewares: h.middleware,
				}
			}

			rt := Route{
				SubRoutes:      subroutes,
				MethodHandlers: hs,
				Pattern:        p,
			}
			rts = append(rts, rt)
		}

		return false
	})
	return rts
}

func (n *node) walk(fn func(eps endpoints, subroutes Routes) bool) bool {
	// Visit the leaf values if any
	if (n.endpoints != nil || n.subroutes != nil) && fn(n.endpoints, n.subroutes) {
		return true
	}

	// Recurse on the children
	for _, ns := range n.children {
		for _, cn := range ns {
			if cn.walk(fn) {
				return true
			}
		}
	}
	return false
}

// longestPrefix finds the length of the shared prefix
// of two strings
func longestPrefix(k1, k2 string) int {
	m := len(k1)
	if l := len(k2); l < m {
		m = l
	}
	var i int
	for i = 0; i < m; i++ {
		if k1[i] != k2[i] {
			break
		}
	}
	return i
}

func methodTypString(method methodType) string {
	for s, t := range methodMap {
		if method == t {
			return s
		}
	}
	return ""
}

func (ns nodes) Sort()              { sort.Sort(ns); ns.tailSort() }
func (ns nodes) Len() int           { return len(ns) }
func (ns nodes) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns nodes) Less(i, j int) bool { return ns[i].label < ns[j].label }

// tailSort pushes nodes with '/' as the tail to the end of the list for param nodes.
// The list order determines the traversal order.
func (ns nodes) tailSort() {
	for i := len(ns) - 1; i >= 0; i-- {
		if ns[i].nodeType > ntStatic && ns[i].tail == '/' {
			ns.Swap(i, len(ns)-1)
			return
		}
	}
}

func (ns nodes) findEdge(label byte) *node {
	num := len(ns)
	idx := 0
	i, j := 0, num-1
	for i <= j {
		idx = i + (j-i)/2
		if label > ns[idx].label {
			i = idx + 1
		} else if label < ns[idx].label {
			j = idx - 1
		} else {
			i = num // breaks cond
		}
	}
	if ns[idx].label != label {
		return nil
	}
	return ns[idx]
}

// Route route definition
type Route struct {
	SubRoutes      Routes
	MethodHandlers map[string]MethodHandler
	Pattern        string
}

// MethodHandler method handlers
type MethodHandler struct {
	Handler     Handler
	Middlewares []Handler
}

// WalkFunc walk func
type WalkFunc func(method string, route string, handler Handler, middlewares ...Handler) error

// Walk walks any router tree that implements Routes interface.
func Walk(r Routes, walkFunc WalkFunc) error {
	return walk(r, walkFunc, "")
}

func walk(r Routes, walkFunc WalkFunc, parentRoute string, routes ...Route) error {
	for _, route := range r.Routes() {
		rts := make([]Route, 0, len(routes))
		copy(rts, routes)
		rts = append(rts, route)
		if route.SubRoutes != nil {
			if err := walk(route.SubRoutes, walkFunc, parentRoute+route.Pattern, rts...); err != nil {
				return err
			}
			continue
		}
		for method, handler := range route.MethodHandlers {
			if method == "*" {
				// Ignore a "catchAll" method, since we pass down all the specific methods for each route.
				continue
			}

			fullRoute := parentRoute + route.Pattern
			fullRoute = strings.Replace(fullRoute, "/*/", "/", -1)
			mws := make([]Handler, 0)
			for _, rt := range rts {
				methodHandler, ok := rt.MethodHandlers[method]
				if ok {
					mws = append(mws, methodHandler.Middlewares...)
				}
			}
			if err := walkFunc(method, fullRoute, handler.Handler, mws...); err != nil {
				return err
			}
		}
	}
	return nil
}
