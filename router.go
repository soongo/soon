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

type Handle func(*Response, *http.Request, func())

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
	panicHandler func(*Response, *http.Request, interface{})
}

const (
	HTTPMethodAll = "ALL"
)

var _ http.Handler = NewRouter()

func defaultPanic(w http.ResponseWriter, _ *http.Request, v interface{}) {
	text := http.StatusText(http.StatusInternalServerError)
	switch err := v.(type) {
	case error:
		text = err.Error()
	case string:
		text = err
	}
	http.Error(w, text, http.StatusInternalServerError)
}

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

func (r *Router) recv(res *Response, req *http.Request) {
	if rcv := recover(); rcv != nil {
		if r.panicHandler != nil {
			r.panicHandler(res, req, rcv)
			return
		}
		defaultPanic(res, req, rcv)
	}
}

func (r *Router) Use(params ...interface{}) {
	length := len(params)
	if length == 2 {
		if route, ok := params[0].(string); ok {
			if router, ok := params[1].(*Router); ok {
				r.mount(route, router)
				return
			} else if middleware, ok := params[1].(func(*Response, *http.Request, func())); ok {
				r.useMiddleware(route, middleware)
				return
			}
			panic("second param should be middleware function or Router")
		}
		panic("route should be string")
	}

	if length == 1 {
		if middleware, ok := params[0].(func(*Response, *http.Request, func())); ok {
			r.useMiddleware("/", middleware)
			return
		}

		if router, ok := params[0].(*Router); ok {
			r.mount("/", router)
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

func (r *Router) GET(route string, handle Handle) {
	r.Handle(http.MethodGet, route, handle)
}

func (r *Router) HEAD(route string, handle Handle) {
	r.Handle(http.MethodHead, route, handle)
}

func (r *Router) POST(route string, handle Handle) {
	r.Handle(http.MethodPost, route, handle)
}

func (r *Router) PUT(route string, handle Handle) {
	r.Handle(http.MethodPut, route, handle)
}

func (r *Router) PATCH(route string, handle Handle) {
	r.Handle(http.MethodPatch, route, handle)
}

func (r *Router) DELETE(route string, handle Handle) {
	r.Handle(http.MethodDelete, route, handle)
}

func (r *Router) OPTIONS(route string, handle Handle) {
	r.Handle(http.MethodOptions, route, handle)
}

func (r *Router) ALL(route string, handle Handle) {
	r.Handle(HTTPMethodAll, route, handle)
}

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

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	res := &Response{w}
	defer r.recv(res, req)

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
			node.handle(res, req, next)
			return
		}

		next()
	}

	next()
}
