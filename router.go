// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/soongo/soon/util"

	pathToRegexp "github.com/soongo/path-to-regexp"
)

// Handle is the handler function of router, router use it to handle matched
// http request, and dispatch a context object into the handler.
type Handle func(*Context)

type paramHandle func(*Context, string)

// ErrorHandle handles the error generated in route handler, and dispatch error
// and context objects into the error handler.
type ErrorHandle func(interface{}, *Context)

type node struct {
	method         string
	route          string
	originalRoute  string
	regexp         *regexp2.Regexp
	baseUrlRegexp  *regexp2.Regexp
	isMiddleware   bool
	appendWildcard bool
	handle         Handle
	errorHandle    ErrorHandle
	tokens         []pathToRegexp.Token
	originalTokens []pathToRegexp.Token
	router         *Router
}

func (n *node) initRegexp() {
	var options *pathToRegexp.Options
	if n.router.routerOption != nil {
		options = n.router.routerOption.toPathToRegexpOption()
	}
	n.regexp = pathToRegexp.Must(pathToRegexp.PathToRegexp(n.route, &n.tokens, options))
	baseUrlRoute := strings.TrimSuffix(n.route, n.originalRoute)
	if baseUrlRoute != "" {
		n.baseUrlRegexp = pathToRegexp.Must(pathToRegexp.PathToRegexp(baseUrlRoute, nil, options))
		ro := regexp2.None
		if options == nil || !options.Sensitive {
			ro = regexp2.IgnoreCase
		}
		route := "(" + strings.TrimSuffix(n.baseUrlRegexp.String(), "$") + ")/(.*)"
		n.baseUrlRegexp = regexp2.MustCompile(route, ro)
	}
	pathToRegexp.Must(pathToRegexp.PathToRegexp(n.originalRoute, &n.originalTokens, options))
}

func (n *node) buildRequestProperties(c *Context, urlPath string) {
	match, err := n.regexp.FindStringMatch(urlPath)
	nGroup := match.GroupCount()

	if n.router.routerOption != nil && n.router.routerOption.MergeParams {
		if err == nil && match != nil && len(n.tokens) > 0 {
			for i, g := range match.Groups() {
				if i > 0 && (!n.appendWildcard || i < nGroup-1) {
					c.Request.Params.Set(n.tokens[i-1].Name, g.String())
				}
			}
		}
	} else {
		c.Request.resetParams()
		nToken := len(n.originalTokens)
		if err == nil && match != nil && nToken > 0 {
			for i, g := range match.Groups() {
				if i > 0 && (!n.appendWildcard || i < nGroup-1) && i >= nGroup-nToken {
					c.Request.Params.Set(n.originalTokens[i-nGroup+nToken].Name, g.String())
				}
			}
		}
	}

	c.Request.BaseUrl = ""
	if n.baseUrlRegexp != nil {
		baseUrlMatch, err := n.baseUrlRegexp.FindStringMatch(urlPath)
		if err == nil && baseUrlMatch != nil && baseUrlMatch.GroupCount() > 1 {
			c.Request.BaseUrl = baseUrlMatch.GroupByNumber(1).String()
		}
	}
}

func (n *node) match(path string) bool {
	m, err := n.regexp.MatchString(path)
	return err == nil && m
}

func (n *node) isErrorHandler() bool {
	return n.errorHandle != nil
}

// RouterOption contains options for router, such as `Sensitive` and `Strict`
type RouterOption struct {
	// When true the regexp will be case sensitive. (default: false)
	Sensitive bool

	// Preserve the req.params values from the parent router.
	// If the parent and the child have conflicting param names, the child’s value take precedence.
	MergeParams bool

	// When true the regexp won't allow an optional trailing delimiter to match. (default: false)
	Strict bool
}

func (o *RouterOption) toPathToRegexpOption() *pathToRegexp.Options {
	return &pathToRegexp.Options{Sensitive: o.Sensitive, Strict: o.Strict}
}

// Router is a http.Handler which can be used to dispatch requests to
// different handler functions.
type Router struct {
	routes []*node

	routerOption *RouterOption

	paramHandles map[string][]paramHandle
}

const (
	// HTTPMethodAll means any http method.
	HTTPMethodAll = "ALL"
)

var _ http.Handler = NewRouter()

// Function to handle error when no other error handlers.
func defaultErrorHandler(v interface{}, c *Context) {
	if !c.finished {
		status := http.StatusInternalServerError
		text := http.StatusText(status)
		switch err := v.(type) {
		case httpError:
			text, status = err.Error(), err.status()
		case error:
			text = err.Error()
		case string:
			text = err
		}

		http.Error(c.Writer, text, status)
		c.finished = true
	}
}

// NewRouter returns a new initialized Router with default configuration.
// Sensitive and Strict is false by default.
func NewRouter(options ...*RouterOption) *Router {
	router := &Router{paramHandles: make(map[string][]paramHandle, 0)}
	if len(options) > 0 {
		router.routerOption = options[0]
	}
	return router
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
	appendWildcard := false
	if !strings.HasSuffix(route, "/(.*)") && !strings.HasSuffix(route, "/(.*)/") {
		route = util.RouteJoin(route, "/(.*)")
		appendWildcard = true
	}
	node := &node{
		route:          route,
		originalRoute:  route,
		isMiddleware:   true,
		handle:         h,
		router:         r,
		appendWildcard: appendWildcard,
	}
	node.initRegexp()
	r.routes = append(r.routes, node)
}

func (r *Router) useErrorHandle(route string, h ErrorHandle) {
	appendWildcard := false
	if !strings.HasSuffix(route, "/(.*)") && !strings.HasSuffix(route, "/(.*)/") {
		route = util.RouteJoin(route, "/(.*)")
		appendWildcard = true
	}
	node := &node{
		route:          route,
		originalRoute:  route,
		errorHandle:    h,
		router:         r,
		appendWildcard: appendWildcard,
	}
	node.initRegexp()
	r.routes = append(r.routes, node)
}

func (r *Router) mount(mountPoint string, router *Router) {
	if router.routerOption == nil {
		router.routerOption = r.routerOption
	}

	for _, v := range router.routes {
		route := util.RouteJoin(mountPoint, v.route)
		route = util.AddPrefixSlash(strings.TrimSuffix(route, "/"))
		node := &node{
			method:         v.method,
			route:          route,
			originalRoute:  v.originalRoute,
			isMiddleware:   v.isMiddleware,
			appendWildcard: v.appendWildcard,
			handle:         v.handle,
			errorHandle:    v.errorHandle,
			router:         v.router,
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
	route = util.AddPrefixSlash(strings.TrimSuffix(route, "/"))
	node := &node{
		method:        method,
		route:         route,
		originalRoute: route,
		isMiddleware:  false,
		handle:        handle,
		router:        r,
	}
	node.initRegexp()
	r.routes = append(r.routes, node)
}

// Param registers a handler on router, and the handler will be triggered
// only by route parameters defined on router routes.
//
// The handler will be called only once in a request-response cycle,
// even if the parameter is matched in multiple routes
func (r *Router) Param(name string, handle paramHandle) {
	if r.paramHandles[name] == nil {
		r.paramHandles[name] = make([]paramHandle, 0)
	}
	r.paramHandles[name] = append(r.paramHandles[name], handle)
}

// ServeHTTP writes reply headers and data to the ResponseWriter and then return.
// Router implements the interface http.Handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c, i, paramCalled := NewContext(req, w), -1, make(map[string]string)

	c.next = func(v ...interface{}) {
		defer r.recv(c)

		if i++; i >= len(r.routes) {
			if len(v) > 0 && v[0] != nil {
				defaultErrorHandler(v[0], c)
			} else {
				http.NotFound(w, req)
			}
			return
		}

		node, urlPath := r.routes[i], req.URL.Path
		isMiddleware, isErrorHandler := node.isMiddleware, node.isErrorHandler()
		match := node.match(urlPath)
		if !match && (isMiddleware || isErrorHandler) && !strings.HasSuffix(urlPath, "/") {
			urlPath += "/"
			match = node.match(urlPath)
		}
		if match {
			if len(v) > 0 && v[0] != nil {
				if isErrorHandler {
					node.buildRequestProperties(c, urlPath)
					node.errorHandle(v[0], c)
					return
				}
				c.next(v[0])
				return
			}

			if node.isMiddleware || node.method == HTTPMethodAll || node.method == req.Method {
				node.buildRequestProperties(c, urlPath)

				if len(node.router.paramHandles) > 0 {
					for n, handles := range node.router.paramHandles {
						if v, ok := c.Request.Params[n]; ok {
							if paramCalled[n] != v {
								paramCalled[n] = v
								for _, h := range handles {
									h(c, v)
								}
							}
						}
					}
				}

				node.handle(c)
				return
			}
		}

		c.next(v...)
	}

	c.next()
}

// Route returns an instance of a single route which you can then use to handle
// HTTP verbs with optional middleware.
// Use router.route() to avoid duplicate route naming and thus typing errors.
func (r *Router) Route(route string) *routerProxy {
	return &routerProxy{router: r, route: route}
}

type routerProxy struct {
	router *Router
	route  string
}

// GET is a shortcut for Handle("GET", handle)
func (r *routerProxy) GET(h Handle) *routerProxy {
	return r.Handle(http.MethodGet, h)
}

// HEAD is a shortcut for Handle("HEAD", handle)
func (r *routerProxy) HEAD(h Handle) *routerProxy {
	return r.Handle(http.MethodHead, h)
}

// POST is a shortcut for Handle("POST", handle)
func (r *routerProxy) POST(h Handle) *routerProxy {
	return r.Handle(http.MethodPost, h)
}

// PUT is a shortcut for Handle("PUT", handle)
func (r *routerProxy) PUT(h Handle) *routerProxy {
	return r.Handle(http.MethodPut, h)
}

// PATCH is a shortcut for Handle("PATCH", handle)
func (r *routerProxy) PATCH(h Handle) *routerProxy {
	return r.Handle(http.MethodPatch, h)
}

// DELETE is a shortcut for Handle("DELETE", handle)
func (r *routerProxy) DELETE(h Handle) *routerProxy {
	return r.Handle(http.MethodDelete, h)
}

// OPTIONS is a shortcut for Handle("OPTIONS", handle)
func (r *routerProxy) OPTIONS(h Handle) *routerProxy {
	return r.Handle(http.MethodOptions, h)
}

// ALL means any http method
func (r *routerProxy) ALL(h Handle) *routerProxy {
	return r.Handle(HTTPMethodAll, h)
}

// Handle registers the handler for matched method
func (r *routerProxy) Handle(method string, h Handle) *routerProxy {
	r.router.Handle(method, r.route, h)
	return r
}
