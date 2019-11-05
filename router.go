// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"
	"path"
)

type TrailingSlashPolicy int

type node struct {
	method       string
	route        string
	isMiddleware bool
	handle       Handle
}

type Handle func(http.ResponseWriter, *http.Request, func())

type Router struct {
	routes []*node

	// How to handle the trailing slash in URL
	// TrailingSlashPolicyStatic (default) or TrailingSlashPolicyRedirect
	// or TrailingSlashPolicyNone
	TrailingSlashPolicy TrailingSlashPolicy

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handle can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

const (
	// Enables automatic use similar route if the current route can't be
	// matched but a handle for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo,
	// the handler for /foo will be used to handle /foo/ as well.
	TrailingSlashPolicyStatic = TrailingSlashPolicy(iota)

	// Enables automatic redirection if the current route can't be matched but a
	// handle for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301.
	TrailingSlashPolicyRedirect

	// Disable automatic handle the trailing slash in URL.
	TrailingSlashPolicyNone

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
	return &Router{TrailingSlashPolicy: TrailingSlashPolicyStatic}
}

func (r *Router) hasRoute(method, route string) bool {
	for _, v := range r.routes {
		if !v.isMiddleware && v.method == method && v.route == route {
			return true
		}
	}
	return false
}

func (r *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		if r.PanicHandler != nil {
			r.PanicHandler(w, req, rcv)
			return
		}
		defaultPanic(w, req, rcv)
	}
}

func (r *Router) Use(middleware Handle) {
	r.routes = append(r.routes, &node{
		route:        "/",
		isMiddleware: true,
		handle:       middleware,
	})
}

func (r *Router) Get(route string, handle Handle) {
	r.Handle(http.MethodGet, route, handle)
}

func (r *Router) Head(route string, handle Handle) {
	r.Handle(http.MethodHead, route, handle)
}

func (r *Router) Post(route string, handle Handle) {
	r.Handle(http.MethodPost, route, handle)
}

func (r *Router) Put(route string, handle Handle) {
	r.Handle(http.MethodPut, route, handle)
}

func (r *Router) Patch(route string, handle Handle) {
	r.Handle(http.MethodPatch, route, handle)
}

func (r *Router) Delete(route string, handle Handle) {
	r.Handle(http.MethodDelete, route, handle)
}

func (r *Router) Options(route string, handle Handle) {
	r.Handle(http.MethodOptions, route, handle)
}

func (r *Router) All(route string, handle Handle) {
	r.routes = append(r.routes, &node{
		method:       HTTPMethodAll,
		route:        addPrefixSlash(route),
		isMiddleware: false,
		handle:       handle,
	})
}

func (r *Router) Handle(method, route string, handle Handle) {
	r.routes = append(r.routes, &node{
		method:       method,
		route:        addPrefixSlash(route),
		isMiddleware: false,
		handle:       handle,
	})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer r.recv(w, req)

	i, urlPath := -1, req.URL.Path
	var next func()
	next = func() {
		if i++; i >= len(r.routes) {
			http.NotFound(w, req)
			return
		}

		node := r.routes[i]
		if (node.isMiddleware && isAncestor(node.route, urlPath)) || (node.route ==
			urlPath && (node.method == HTTPMethodAll || node.method == req.Method)) {
			node.handle(w, req, next)
			return
		}

		if node.method != req.Method || r.TrailingSlashPolicy == TrailingSlashPolicyNone {
			next()
			return
		}

		if similar(node.route, urlPath) {
			if r.hasRoute(node.method, urlPath) {
				next()
				return
			}

			if r.TrailingSlashPolicy == TrailingSlashPolicyRedirect {
				p := path.Clean(node.route)
				if similar(urlPath, p) && !r.hasRoute(node.method, p) {
					next()
					return
				}
				req.URL.Path = node.route
				http.Redirect(w, req, req.URL.String(), http.StatusMovedPermanently)
				return
			}

			node.handle(w, req, next)
			return
		}

		next()
	}

	next()
}
