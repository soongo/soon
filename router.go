// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"

	"github.com/dlclark/regexp2"
	pathToRegexp "github.com/soongo/path-to-regexp"
)

type node struct {
	method       string
	route        string
	regexp       *regexp2.Regexp
	isMiddleware bool
	handle       Handle
	options      *pathToRegexp.Options
}

func (n *node) match(path string) bool {
	m, err := n.regexp.MatchString(path)
	if err != nil {
		panic(err)
	}
	return m
}

// Handle is the handler function of router, router use it to handle matched
// http request, and dispatch req, res into the handler.
type Handle func(*http.Request, *Response, func())

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions
type Router struct {
	routes []*node

	// When true the regexp will be case sensitive. (default: false)
	Sensitive bool

	// When true the regexp allows an optional trailing delimiter to match. (default: false)
	Strict bool

	options *pathToRegexp.Options

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handle can be used to keep your server from crashing because of
	// unrecovered panics.
	panicHandler func(*http.Request, *Response, interface{})
}

const (
	// HTTPMethodAll means any http method.
	HTTPMethodAll = "ALL"
)

var _ http.Handler = NewRouter()

func defaultPanic(_ *http.Request, w http.ResponseWriter, v interface{}) {
	text := http.StatusText(http.StatusInternalServerError)
	switch err := v.(type) {
	case error:
		text = err.Error()
	case string:
		text = err
	}
	http.Error(w, text, http.StatusInternalServerError)
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

func (r *Router) recv(req *http.Request, res *Response) {
	if rcv := recover(); rcv != nil {
		if r.panicHandler != nil {
			r.panicHandler(req, res, rcv)
			return
		}
		defaultPanic(req, res, rcv)
	}
}

// Use the given middleware function, or mount another router,
// with optional path, defaulting to "/".
func (r *Router) Use(params ...interface{}) {
	length := len(params)
	if length == 2 {
		if route, ok := params[0].(string); ok {
			if router, ok := params[1].(*Router); ok {
				r.mount(route, router)
				return
			} else if middleware, ok := params[1].(func(*http.Request, *Response, func())); ok {
				r.useMiddleware(route, middleware)
				return
			}
			panic("second param should be middleware function or Router")
		}
		panic("route should be string")
	}

	if length == 1 {
		if router, ok := params[0].(*Router); ok {
			r.mount("/", router)
			return
		}

		if middleware, ok := params[0].(func(*http.Request, *Response, func())); ok {
			r.useMiddleware("/", middleware)
			return
		}

		panic("params should be middleware function or Router")
	}

	panic("params count should be 1 or 2")
}

func (r *Router) useMiddleware(route string, middleware Handle) {
	r.initOptions()
	route = routeJoin(route, "/(.*)")
	regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp(route, nil, r.options))
	r.routes = append(r.routes, &node{
		route:        route,
		regexp:       regexp,
		isMiddleware: true,
		handle:       middleware,
		options:      r.options,
	})
}

func (r *Router) mount(mountPoint string, router *Router) {
	for _, v := range router.routes {
		route := routeJoin(mountPoint, v.route)
		regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp(route, nil, v.options))
		r.routes = append(r.routes, &node{
			method:       v.method,
			route:        route,
			regexp:       regexp,
			isMiddleware: v.isMiddleware,
			handle:       v.handle,
			options:      v.options,
		})
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

// Handle registers the handler for the http request which matched the method and route.
// And dispatch a req, res into the handler.
func (r *Router) Handle(method, route string, handle Handle) {
	r.initOptions()
	route = addPrefixSlash(route)
	regexp := pathToRegexp.Must(pathToRegexp.PathToRegexp(route, nil, r.options))
	r.routes = append(r.routes, &node{
		method:       method,
		route:        route,
		regexp:       regexp,
		isMiddleware: false,
		handle:       handle,
		options:      r.options,
	})
}

// ServeHTTP writes reply headers and data to the ResponseWriter and then return.
// Router implements the interface http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	res := &Response{w}
	defer r.recv(req, res)

	i, urlPath := -1, req.URL.Path
	var next func()
	next = func() {
		if i++; i >= len(r.routes) {
			http.NotFound(w, req)
			return
		}

		node := r.routes[i]
		if node.match(urlPath) && (node.isMiddleware ||
			node.method == HTTPMethodAll || node.method == req.Method) {
			node.handle(req, res, next)
			return
		}

		next()
	}

	next()
}
