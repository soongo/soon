// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"

	"github.com/soongo/soon/util"

	"github.com/dlclark/regexp2"
	pathToRegexp "github.com/soongo/path-to-regexp"
)

// Handle is the handler function of router, router use it to handle matched
// http request, and dispatch a context object into the handler.
type Handle func(*Context)

// ErrorHandle handles the error generated in route handler, and dispatch error
// and context objects into the error handler.
type ErrorHandle func(interface{}, *Context)

type node struct {
	method       string
	route        string
	regexp       *regexp2.Regexp
	isMiddleware bool
	handle       Handle
	errorHandle  ErrorHandle
	options      *pathToRegexp.Options
	tokens       []pathToRegexp.Token
}

func (n *node) initRegexp() {
	n.regexp = pathToRegexp.Must(pathToRegexp.PathToRegexp(
		n.route, &n.tokens, n.options))
}

func (n *node) buildRequestParams(c *Context) {
	if len(n.tokens) > 0 {
		c.resetParams()
		match, err := n.regexp.FindStringMatch(c.URL.Path)
		if err != nil {
			panic(err)
		}
		for i, g := range match.Groups() {
			if i > 0 {
				c.Params.Set(n.tokens[i-1].Name, g.String())
			}
		}
	}
}

func (n *node) match(path string) bool {
	m, err := n.regexp.MatchString(path)
	if err != nil {
		panic(err)
	}
	return m
}

func (n *node) isErrorHandler() bool {
	return n.errorHandle != nil
}

// Router is a http.Handler which can be used to dispatch requests to
// different handler functions.
type Router struct {
	routes []*node

	// When true the regexp will be case sensitive. (default: false)
	Sensitive bool

	// When true the regexp allows an optional trailing delimiter to match. (default: false)
	Strict bool

	options *pathToRegexp.Options
}

const (
	// HTTPMethodAll means any http method.
	HTTPMethodAll = "ALL"
)

var _ http.Handler = NewRouter()

// Function to handle error when no other error handlers.
func defaultErrorHandler(v interface{}, c *Context) {
	text := http.StatusText(http.StatusInternalServerError)
	switch err := v.(type) {
	case error:
		text = err.Error()
	case string:
		text = err
	}
	http.Error(c.ResponseWriter, text, http.StatusInternalServerError)
}

// NewRouter returns a new initialized Router with default configuration.
// Sensitive and Strict is false by default.
func NewRouter() *Router {
	return &Router{}
}

func (r *Router) hasRoute(method, route string) bool {
	for _, v := range r.routes {
		if !v.isMiddleware && v.method == method && v.match(route) {
			return true
		}
	}
	return false
}

func (r *Router) initOptions() {
	if r.options == nil {
		r.options = &pathToRegexp.Options{
			Sensitive: r.Sensitive,
			Strict:    r.Strict,
		}
	}
}

func (r *Router) recv(c *Context) {
	if rcv := recover(); rcv != nil {
		c.next(rcv)
	}
}

// Use the given middleware, or error handler, or mount another router,
// with optional path, defaulting to "/".
func (r *Router) Use(params ...interface{}) {
	length := len(params)
	if length > 2 || length == 0 {
		panic("params count should be 1 or 2")
	}

	if length == 2 {
		if _, ok := params[0].(string); !ok {
			panic("route should be string")
		}
	}

	route := "/"
	if v, ok := params[0].(string); ok {
		route = v
	}

	var handle = params[length-1]

	if router, ok := handle.(*Router); ok {
		r.mount(route, router)
		return
	}

	if m, ok := handle.(func(*Context)); ok {
		r.useMiddleware(route, m)
		return
	}

	if h, ok := handle.(func(interface{}, *Context)); ok {
		r.useErrorHandle(route, h)
		return
	}

	msg := "params should be middleware or error handler or router"
	if length == 2 {
		msg = "second " + msg
	}
	panic(msg)
}

func (r *Router) useMiddleware(route string, h Handle) {
	r.initOptions()
	route = util.RouteJoin(route, "/(.*)")
	node := &node{
		route:        route,
		isMiddleware: true,
		handle:       h,
		options:      r.options,
	}
	node.initRegexp()
	r.routes = append(r.routes, node)
}

func (r *Router) useErrorHandle(route string, h ErrorHandle) {
	r.initOptions()
	route = util.RouteJoin(route, "/(.*)")
	node := &node{
		route:       route,
		errorHandle: h,
		options:     r.options,
	}
	node.initRegexp()
	r.routes = append(r.routes, node)
}

func (r *Router) mount(mountPoint string, router *Router) {
	for _, v := range router.routes {
		route := util.RouteJoin(mountPoint, v.route)
		node := &node{
			method:       v.method,
			route:        route,
			isMiddleware: v.isMiddleware,
			handle:       v.handle,
			errorHandle:  v.errorHandle,
			options:      v.options,
		}
		node.initRegexp()
		r.routes = append(r.routes, node)
	}
}

// GET is a shortcut for router.Handle(http.MethodGet, route, handle)
func (r *Router) GET(route string, handle Handle) {
	r.Handle(http.MethodGet, route, handle)
}

// HEAD is a shortcut for router.Handle(http.MethodHead, route, handle)
func (r *Router) HEAD(route string, handle Handle) {
	r.Handle(http.MethodHead, route, handle)
}

// POST is a shortcut for router.Handle(http.MethodPost, route, handle)
func (r *Router) POST(route string, handle Handle) {
	r.Handle(http.MethodPost, route, handle)
}

// PUT is a shortcut for router.Handle(http.MethodPut, route, handle)
func (r *Router) PUT(route string, handle Handle) {
	r.Handle(http.MethodPut, route, handle)
}

// PATCH is a shortcut for router.Handle(http.MethodPatch, route, handle)
func (r *Router) PATCH(route string, handle Handle) {
	r.Handle(http.MethodPatch, route, handle)
}

// DELETE is a shortcut for router.Handle(http.MethodDelete, route, handle)
func (r *Router) DELETE(route string, handle Handle) {
	r.Handle(http.MethodDelete, route, handle)
}

// OPTIONS is a shortcut for router.Handle(http.MethodOptions, route, handle)
func (r *Router) OPTIONS(route string, handle Handle) {
	r.Handle(http.MethodOptions, route, handle)
}

// ALL means any http method, so this is a shortcut for
// router.Handle(http.MethodAny, route, handle)
func (r *Router) ALL(route string, handle Handle) {
	r.Handle(HTTPMethodAll, route, handle)
}

// Handle registers the handler for the http request which matched the method
// and route, and dispatch a context object into the handler.
func (r *Router) Handle(method, route string, handle Handle) {
	r.initOptions()
	route = util.AddPrefixSlash(route)
	node := &node{
		method:       method,
		route:        route,
		isMiddleware: false,
		handle:       handle,
		options:      r.options,
	}
	node.initRegexp()
	r.routes = append(r.routes, node)
}

// ServeHTTP writes reply headers and data to the ResponseWriter and then return.
// Router implements the interface http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := &Context{Request: &Request{Request: req}, ResponseWriter: w}
	i, urlPath := -1, req.URL.Path
	c.next = func(v ...interface{}) {
		if i++; i >= len(r.routes) {
			if len(v) > 0 && v[0] != nil {
				defaultErrorHandler(v[0], c)
			} else {
				http.NotFound(w, req)
			}
			return
		}

		node := r.routes[i]
		if node.match(urlPath) {
			if len(v) > 0 && v[0] != nil {
				if node.isErrorHandler() {
					node.buildRequestParams(c)
					node.errorHandle(v[0], c)
					return
				}
				c.next(v[0])
				return
			}

			if node.isMiddleware || node.method == HTTPMethodAll || node.method == req.Method {
				node.buildRequestParams(c)
				node.handle(c)
				return
			}
		}

		c.next()
	}

	defer r.recv(c)

	c.next()
}
