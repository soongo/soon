// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type header map[string]string

type test struct {
	route             string
	middlewareRoute   string
	errorHandlerRoute string
	routerOption      *RouterOption
	path              string
	statusCode        int
	header            header
	body              string
	middleware        func(c *Context)
	err               interface{}
	errorHandle       func(v interface{}, c *Context)
}

const (
	body200 = `{"alive": true}`
	body404 = "404 page not found"
)

var methods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
}

func makeHandle(t test) Handle {
	return func(c *Context) {
		if t.err != nil {
			panic(t.err)
		}

		for k, v := range t.header {
			c.Set(k, v)
		}
		if t.statusCode != 0 {
			c.Status(t.statusCode)
		}
		if t.body != "" {
			c.Send(t.body)
		}
	}
}

func request(method, url string, h http.Header) (statusCode int,
	header http.Header, body string, err error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}
	req.Header = h
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	header = resp.Header

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	body = strings.TrimSpace(string(byteArray))

	return
}

func TestRouter(t *testing.T) {
	tests := []test{
		{route: "/", path: "/", statusCode: 200, body: body200},
		{route: "/", path: "", statusCode: 200, body: body200},
		{route: "//", path: "/", statusCode: 404, body: body404},
		{route: "//", path: "//", statusCode: 200, body: body200},
		{route: "//", path: "///", statusCode: 404, body: body404},
		{route: "", path: "/", statusCode: 200, body: body200},
		{route: "", path: "", statusCode: 200, body: body200},
		{
			route:      "/health-check/",
			path:       "/health-check/",
			statusCode: 200,
			header:     header{"User-Agent": "go-http-client"},
			body:       body200,
		},
		{
			route:        "/HEALTH-check/",
			routerOption: &RouterOption{true, true},
			path:         "/health-check/",
			statusCode:   404,
			body:         body404,
		},
		{route: "/health-check/", path: "/health-check", statusCode: 404, body: body404},
		{route: "/health-check//", path: "/health-check/", statusCode: 404, body: body404},
		{route: "/health-check//", path: "/health-check//", statusCode: 200, body: body200},
		{route: "/health-check//", path: "/health-check///", statusCode: 404, body: body404},
		{route: "/health-check", path: "/health-check/", statusCode: 200, body: body200},
		{
			route:        "/health-check",
			routerOption: &RouterOption{false, true},
			path:         "/health-check/",
			statusCode:   404,
			body:         body404,
		},
		{
			route:        "/health-check/",
			routerOption: &RouterOption{false, true},
			path:         "/health-check/",
			statusCode:   404,
			body:         body404,
		},
		{route: "/health-check", path: "/health-check", statusCode: 200, body: body200},
	}

	t.Run("one-by-one", func(t *testing.T) {
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				assert := assert.New(t)
				router := NewRouter(tt.routerOption)
				router.GET(tt.route, makeHandle(tt))
				server := httptest.NewServer(router)
				defer server.Close()

				statusCode, header, body, err := request("GET", server.URL+tt.path, nil)
				assert.Nil(err)
				assert.Equal(tt.statusCode, statusCode)
				for k, v := range tt.header {
					assert.Equal(v, header.Get(k))
				}
				assert.Equal(tt.body, body)

				statusCode, _, _, err = request("HEAD", server.URL+tt.path, nil)
				assert.Nil(err)
				assert.Equal(404, statusCode)
			})
		}
	})

	tests = []test{
		{route: "/", path: "/", statusCode: 200, body: body200},
		{route: "/", path: "", statusCode: 200, body: body200},
		{route: "//", path: "/", statusCode: 200, body: body200},
		{route: "//", path: "//", statusCode: 200, body: body200},
		{route: "//", path: "///", statusCode: 404, body: body404},
		{route: "", path: "/", statusCode: 200, body: body200},
		{route: "", path: "", statusCode: 200, body: body200},
		{
			route:      "/health-check/",
			path:       "/health-check/",
			statusCode: 200,
			header:     header{"User-Agent": "go-http-client"},
			body:       body200,
		},
		{route: "/health-check/", path: "/health-check", statusCode: 200, body: body200},
		{route: "/health-check//", path: "/health-check/", statusCode: 200, body: body200},
		{route: "/health-check//", path: "/health-check//", statusCode: 200, body: body200},
		{route: "/health-check//", path: "/health-check///", statusCode: 404, body: body404},
		{route: "/health-check", path: "/health-check/", statusCode: 200, body: body200},
		{route: "/health-check", path: "/health-check", statusCode: 200, body: body200},
	}

	t.Run("all", func(t *testing.T) {
		router := NewRouter()
		for _, tt := range tests {
			router.GET(tt.route, makeHandle(tt))
		}
		server := httptest.NewServer(router)
		defer server.Close()

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				assert := assert.New(t)
				statusCode, header, body, err := request("GET", server.URL+tt.path, nil)
				assert.Nil(err)
				assert.Equal(tt.statusCode, statusCode)
				for k, v := range tt.header {
					assert.Equal(v, header.Get(k))
				}
				assert.Equal(tt.body, body)
			})
		}
	})

	t.Run("sub-router-with-custom-options", func(t *testing.T) {
		assert := assert.New(t)
		router := NewRouter(&RouterOption{true, true})
		router_1 := NewRouter()
		router_1.GET("/1-a", func(c *Context) {
			c.String(body200)
		})
		router_1_1 := NewRouter(&RouterOption{false, false})
		router_1_1.GET("/1-1-a", func(c *Context) {
			c.String(body200)
		})
		router_1_1_1 := NewRouter()
		router_1_1_1.GET("/1-1-1-a", func(c *Context) {
			c.String(body200)
		})
		router_1_1_1.GET("/1-1-1-a/b", func(c *Context) {
			c.Next()
		})
		router_1_1_1_1 := NewRouter(&RouterOption{true, true})
		router_1_1_1_1.GET("/1-1-1-a/b", func(c *Context) {
			c.String(body200)
		})
		router_1_1_1.Use("/", router_1_1_1_1)
		router_1_1.Use("/1-1", router_1_1_1)
		router_1.Use("/1", router_1_1)
		router.Use("/", router_1)

		server := httptest.NewServer(router)
		defer server.Close()

		tests := []struct {
			path       string
			statusCode int
			body       string
		}{
			{"/1-a", 200, body200},
			{"/1-a/", 404, body404},
			{"/1-A", 404, body404},
			{"/1-A/", 404, body404},
			{"/1/1-1-a", 200, body200},
			{"/1/1-1-a/", 200, body200},
			{"/1/1-1-A", 200, body200},
			{"/1/1-1-A/", 200, body200},
			{"/1/1-1/1-1-1-a", 200, body200},
			{"/1/1-1/1-1-1-a/", 200, body200},
			{"/1/1-1/1-1-1-A", 200, body200},
			{"/1/1-1/1-1-1-A/", 200, body200},
			{"/1/1-1/1-1-1-a/b", 200, body200},
			{"/1/1-1/1-1-1-a/b/", 404, body404},
			{"/1/1-1/1-1-1-A/B", 404, body404},
			{"/1/1-1/1-1-1-A/B/", 404, body404},
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				statusCode, _, body, err := request("GET", server.URL+tt.path, nil)
				assert.Nil(err)
				assert.Equal(tt.statusCode, statusCode)
				assert.Equal(tt.body, body)
			})
		}
	})
}

func TestRouterMiddleware(t *testing.T) {
	tests := []test{
		{
			route:      "/",
			path:       "/",
			statusCode: 200,
			middleware: func(c *Context) {
				c.Send(body200)
			},
		},
		{
			route:      "/foo/:foo",
			path:       "/foo/foo1",
			statusCode: 200,
			middleware: func(c *Context) {
				c.Send(body200)
			},
		},
		{
			route:           "/foo/:foo/bar/:bar",
			path:            "/foo/foo1/bar/bar1",
			middlewareRoute: "/foo/:foo",
			statusCode:      200,
			middleware: func(c *Context) {
				c.Send(body200)
			},
		},
		{
			route:           "/foo/:foo/bar/:bar",
			path:            "/foo/foo1/bar/bar1",
			middlewareRoute: "/foo/bar/:foo",
			statusCode:      200,
			body:            body200,
			middleware: func(c *Context) {
				c.Send(body404)
			},
		},
		{
			route:      "/",
			path:       "/404",
			statusCode: 404,
			middleware: func(c *Context) {
				http.Error(c.Writer, body200, 404)
			},
		},
		{
			route:      "/",
			path:       "/",
			statusCode: 500,
			middleware: func(c *Context) {
				panic(body200)
			},
		},
		{
			route:      "/",
			path:       "/",
			statusCode: 500,
			middleware: func(c *Context) {
				c.Next(body200)
			},
		},
		{
			route:      "/",
			path:       "/",
			statusCode: 500,
			err:        body200,
		},
		{
			route:      "/",
			path:       "/",
			statusCode: 200,
			err:        errors.New(body200),
			errorHandle: func(v interface{}, c *Context) {
				c.Send(v.(error).Error())
			},
		},
		{
			route:      "/foo/:foo",
			path:       "/foo/foo1",
			statusCode: 200,
			err:        errors.New(body200),
			errorHandle: func(v interface{}, c *Context) {
				c.Send(v.(error).Error())
			},
		},
		{
			route:             "/foo/:foo/bar/:bar",
			path:              "/foo/foo1/bar/bar1",
			errorHandlerRoute: "/foo/:bar",
			statusCode:        200,
			err:               errors.New(body200),
			errorHandle: func(v interface{}, c *Context) {
				c.Send(v.(error).Error())
			},
		},
		{
			route:             "/foo/:foo/bar/:bar",
			path:              "/foo/foo1/bar/bar1",
			errorHandlerRoute: "/foo/bar/:foo",
			statusCode:        500,
			err:               errors.New(body200),
			errorHandle: func(v interface{}, c *Context) {
				c.Send(v.(error).Error())
			},
		},
		{
			route:      "/",
			path:       "/",
			statusCode: 500,
			err:        errors.New(body200),
			errorHandle: func(v interface{}, c *Context) {
				panic(v)
			},
		},
		{
			route:      "/",
			path:       "/",
			statusCode: 500,
			err:        errors.New(body200),
			errorHandle: func(v interface{}, c *Context) {
				c.Next(v)
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router := NewRouter()
			if tt.middleware != nil {
				router.Use(stringOr(tt.middlewareRoute, tt.route), tt.middleware)
			}
			router.GET(tt.route, makeHandle(tt))
			if tt.errorHandle != nil {
				router.Use(stringOr(tt.errorHandlerRoute, tt.route), tt.errorHandle)
			}
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request("GET", server.URL+tt.path, nil)
			assert.Nil(err)
			assert.Equal(tt.statusCode, statusCode)
			assert.Equal(body200, body)
		})
	}
}

func TestRouter_Params(t *testing.T) {
	tests := []struct {
		route1            string
		route2            string
		route3            string
		path1             string
		path2             string
		path3             string
		params1           Params
		params2           Params
		params3           Params
		middlewareParams1 Params
		middlewareParams2 Params
		middlewareParams3 Params
	}{
		{
			"/:foo",
			"/:bar",
			"/(.*)",
			"/foo",
			"/foo/bar",
			"/foo/bar/test",
			Params{"foo": "foo"},
			Params{"bar": "bar"},
			Params{0: "test"},
			Params{"foo": "foo", 0: ""},
			Params{"bar": "bar", 0: ""},
			Params{0: "test", 1: ""},
		},
		{
			"/:foo",
			"/([^/]*)",
			"/(.*)",
			"/foo",
			"/foo/bar",
			"/foo/bar/test",
			Params{"foo": "foo"},
			Params{0: "bar"},
			Params{0: "test"},
			Params{"foo": "foo", 0: ""},
			Params{0: "bar", 1: ""},
			Params{0: "test", 1: ""},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router1 := NewRouter()
			router1.Use(tt.route1, func(c *Context) {
				if c.Request.URL.Path == tt.path1 {
					assert.Equal(tt.middlewareParams1, c.Params())
				}
				c.Next()
			})
			router1.GET(tt.route1, func(c *Context) {
				assert.Equal(tt.params1, c.Params())
				c.String(body200)
			})
			router2 := NewRouter()
			router2.Use(tt.route2, func(c *Context) {
				if c.Request.URL.Path == tt.path2 {
					assert.Equal(tt.middlewareParams2, c.Params())
				}
				c.Next()
			})
			router2.GET(tt.route2, func(c *Context) {
				assert.Equal(tt.params2, c.Params())
				c.String(body200)
			})
			router3 := NewRouter()
			router3.Use(tt.route3, func(c *Context) {
				if c.Request.URL.Path == tt.path3 {
					assert.Equal(tt.middlewareParams3, c.Params())
				}
				c.Next()
			})
			router3.Use(func(c *Context) {
				assert.Equal(tt.params3, c.Params())
				c.Next()
			})
			router3.GET(tt.route3, func(c *Context) {
				assert.Equal(tt.params3, c.Params())
				c.String(body200)
			})
			router2.Use(tt.route2, router3)
			router1.Use(tt.route1, router2)
			router1.Use(func(v interface{}, c *Context) {
				assert.Nil(v)
			})
			server := httptest.NewServer(router1)
			defer server.Close()

			statusCode, _, _, err := request("GET", server.URL+tt.path1, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)

			statusCode, _, _, err = request("GET", server.URL+tt.path2, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)

			statusCode, _, _, err = request("GET", server.URL+tt.path3, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)
		})
	}

	t.Run("panic", func(t *testing.T) {
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				assert := assert.New(t)
				router1 := NewRouter()
				router1.GET(tt.route1, func(c *Context) {
					panic("error")
				})
				router2 := NewRouter()
				router2.GET(tt.route2, func(c *Context) {
					panic("error")
				})
				router2.Use(tt.route2, func(v interface{}, c *Context) {
					assert.NotNil(v)
					assert.Equal(tt.middlewareParams2, c.Params())
					c.SendStatus(500)
				})
				router3 := NewRouter()
				router3.GET(tt.route3, func(c *Context) {
					panic("error")
				})
				router3.Use(tt.route3, func(v interface{}, c *Context) {
					assert.NotNil(v)
					assert.Equal(tt.middlewareParams3, c.Params())
					c.SendStatus(500)
				})
				router2.Use(tt.route2, router3)
				router1.Use(tt.route1, router2)
				router1.Use(tt.route1, func(v interface{}, c *Context) {
					assert.NotNil(v)
					assert.Equal(tt.middlewareParams1, c.Params())
					c.SendStatus(500)
				})

				server := httptest.NewServer(router1)
				defer server.Close()

				statusCode, _, _, err := request("GET", server.URL+tt.path1, nil)
				assert.Nil(err)
				assert.Equal(500, statusCode)

				statusCode, _, _, err = request("GET", server.URL+tt.path2, nil)
				assert.Nil(err)
				assert.Equal(500, statusCode)

				statusCode, _, _, err = request("GET", server.URL+tt.path3, nil)
				assert.Nil(err)
				assert.Equal(500, statusCode)
			})
		}
	})
}

func TestRouterMethods(t *testing.T) {
	tt := test{route: "/", path: "/", body: body200}
	check := func(method, url string) {
		statusCode, _, body, err := request(method, url, nil)
		assert.Nil(t, err)
		assert.Equal(t, 200, statusCode)

		if method != "HEAD" {
			assert.Equal(t, tt.body, body)
		}

		assert := assert.New(t)
		for _, m := range methods {
			if m != method {
				statusCode, _, body, err = request(m, url, nil)
				assert.Nil(err)
				assert.Equal(404, statusCode)
				if m != "HEAD" {
					assert.Equal(body404, body)
				}
			}
		}
	}

	t.Run("GET", func(t *testing.T) {
		router := NewRouter()
		router.GET(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("GET", server.URL+tt.path)
	})

	t.Run("HEAD", func(t *testing.T) {
		router := NewRouter()
		router.HEAD(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("HEAD", server.URL+tt.path)
	})

	t.Run("POST", func(t *testing.T) {
		router := NewRouter()
		router.POST(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("POST", server.URL+tt.path)
	})

	t.Run("PUT", func(t *testing.T) {
		router := NewRouter()
		router.PUT(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("PUT", server.URL+tt.path)
	})

	t.Run("PATCH", func(t *testing.T) {
		router := NewRouter()
		router.PATCH(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("PATCH", server.URL+tt.path)
	})

	t.Run("DELETE", func(t *testing.T) {
		router := NewRouter()
		router.DELETE(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("DELETE", server.URL+tt.path)
	})

	t.Run("OPTIONS", func(t *testing.T) {
		router := NewRouter()
		router.OPTIONS(tt.route, makeHandle(tt))
		server := httptest.NewServer(router)
		defer server.Close()
		check("OPTIONS", server.URL+tt.path)
	})
}

func TestRouter_Handle(t *testing.T) {
	for i, method := range methods {
		t.Run(method, func(t *testing.T) {
			assert := assert.New(t)
			tt := test{route: "/", path: "/", body: body200}
			router := NewRouter()
			router.Handle(method, tt.route, makeHandle(tt))
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request(method, server.URL+tt.path, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)
			if method != "HEAD" {
				assert.Equal(tt.body, body)
			}

			method = methods[(i+1)%len(methods)]
			statusCode, _, body, err = request(method, server.URL+tt.path, nil)
			assert.Nil(err)
			assert.Equal(404, statusCode)
			if method != "HEAD" {
				assert.Equal(body404, body)
			}
		})
	}
}

func TestRouter_ALL(t *testing.T) {
	tt := test{route: "/", path: "/", body: body200}
	router := NewRouter()
	router.ALL(tt.route, makeHandle(tt))
	server := httptest.NewServer(router)
	defer server.Close()

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			assert := assert.New(t)
			statusCode, _, body, err := request(method, server.URL+tt.path, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)
			if method != "HEAD" {
				assert.Equal(tt.body, body)
			}
		})
	}
}

func TestRouter_Param(t *testing.T) {
	tests := []struct {
		method     string
		route      string
		path       string
		paramName  string
		paramValue string
		err        error
	}{
		{"GET", "/:foo", "/bar", "foo", "bar", nil},
		{"POST", "/name/:foo", "/name/bar", "foo", "bar", nil},
		{"PATCH", "/name/:foo", "/name/bar", "foo", "bar", errors.New("err")},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router := NewRouter()
			router.Handle(tt.method, tt.route, func(c *Context) {
				assert.Equal(tt.paramValue, c.Locals().Get(tt.paramName))
			})
			if tt.err != nil {
				router.Use(func(v interface{}, c *Context) {
					assert.Equal(nil, c.Locals().Get(tt.paramName))
				})
			}
			router.Param(tt.paramName, func(r *Request, s string) {
				if tt.err != nil {
					panic(tt.err)
				}
				r.Locals.Set(tt.paramName, s)
			})
			server := httptest.NewServer(router)
			defer server.Close()
			_, _, _, err := request(tt.method, server.URL+tt.path, nil)
			assert.Nil(err)
		})
	}

	t.Run("once", func(t *testing.T) {
		assert := assert.New(t)
		router, calledCount := NewRouter(), 0
		router.GET("/name/:foo", func(c *Context) {
			c.Next()
		})
		router.GET("/name/:foo", func(c *Context) {
			c.String("foo")
		})
		router.Param("foo", func(r *Request, s string) {
			r.Locals.Set("foo", s)
			calledCount++
		})
		server := httptest.NewServer(router)
		defer server.Close()
		_, _, _, err := request("GET", server.URL+"/name/bar", nil)
		assert.Nil(err)
		assert.Equal(1, calledCount)
		n := 2
		for i := 0; i < n; i++ {
			_, _, _, err = request("GET", server.URL+"/name/bar", nil)
			assert.Nil(err)
		}
		assert.Equal(n+1, calledCount)
	})

	t.Run("multiple", func(t *testing.T) {
		assert := assert.New(t)
		router, calledCount := NewRouter(), 0
		router.Use("/:foo", func(c *Context) {
			assert.Equal("name", c.Locals().Get("foo"))
			c.Next()
		})
		router.GET("/name/:foo/:bar", func(c *Context) {
			assert.Equal("id", c.Locals().Get("foo"))
			c.Next()
		})
		router.GET("/name/:foo/:bar", func(c *Context) {
			assert.Equal("id", c.Locals().Get("foo"))
			c.Next()
		})
		router.GET("/name/id/:foo", func(c *Context) {
			assert.Equal("bar", c.Locals().Get("foo"))
			c.String("foo")
		})
		router.Param("foo", func(r *Request, s string) {
			r.Locals.Set("foo", s)
			calledCount++
		})
		server := httptest.NewServer(router)
		defer server.Close()
		_, _, _, err := request("GET", server.URL+"/name/id/bar", nil)
		assert.Nil(err)
		assert.Equal(3, calledCount)
		n := 2
		for i := 0; i < n; i++ {
			_, _, _, err = request("GET", server.URL+"/name/id/bar", nil)
			assert.Nil(err)
		}
		assert.Equal(n*3+3, calledCount)
	})

	t.Run("sub-router", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router, subRouter, calledCount := NewRouter(), NewRouter(), 0
			router.GET("/name/:foo", func(c *Context) {
				c.String("body")
			})
			subRouter.Param("foo", func(r *Request, s string) {
				calledCount++
			})
			router.Use(subRouter)
			server := httptest.NewServer(router)
			defer server.Close()
			_, _, _, err := request("GET", server.URL+"/name/bar", nil)
			assert.Nil(err)
			assert.Equal(0, calledCount)
		})

		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router, subRouter, calledCount := NewRouter(), NewRouter(), 0
			subRouter.GET("/name/:foo", func(c *Context) {
				c.String("body")
			})
			subRouter.Param("foo", func(r *Request, s string) {
				calledCount++
			})
			router.Use(subRouter)
			server := httptest.NewServer(router)
			defer server.Close()
			_, _, _, err := request("GET", server.URL+"/name/bar", nil)
			assert.Nil(err)
			assert.Equal(1, calledCount)
		})

		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router, subRouter, calledCount := NewRouter(), NewRouter(), 0
			router.Use("/:foo", func(c *Context) {
				c.Next()
			})
			router.GET("/name/:foo", func(c *Context) {
				c.Next()
			})
			subRouter.Use("/:foo", func(c *Context) {
				c.Next()
			})
			subRouter.GET("/name/:foo", func(c *Context) {
				c.String("body")
			})
			subRouter.Param("foo", func(r *Request, s string) {
				calledCount++
			})
			router.Use(subRouter)
			server := httptest.NewServer(router)
			defer server.Close()
			_, _, _, err := request("GET", server.URL+"/name/bar", nil)
			assert.Nil(err)
			assert.Equal(2, calledCount)
		})

		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router, subRouter, calledCount := NewRouter(), NewRouter(), 0
			router.Use("/:foo", func(c *Context) {
				c.Next()
			})
			router.GET("/name/:foo", func(c *Context) {
				c.Next()
			})
			router.Param("foo", func(r *Request, s string) {
				calledCount++
			})
			subRouter.Use("/:foo", func(c *Context) {
				c.Next()
			})
			subRouter.GET("/name/:foo", func(c *Context) {
				c.String("body")
			})
			subRouter.Param("foo", func(r *Request, s string) {
				calledCount++
			})
			router.Use(subRouter)
			server := httptest.NewServer(router)
			defer server.Close()
			_, _, _, err := request("GET", server.URL+"/name/bar", nil)
			assert.Nil(err)
			assert.Equal(4, calledCount)
		})
	})
}

func TestRouter_Use(t *testing.T) {
	deferFn := func() {
		assert.NotNil(t, recover())
	}

	childRouter := NewRouter()
	childRouter.Use(func(c *Context) {})

	tests := []struct {
		params  []interface{}
		deferFn func()
	}{
		{[]interface{}{}, deferFn},
		{[]interface{}{1}, deferFn},
		{[]interface{}{1, 1}, deferFn},
		{[]interface{}{1, func() {}}, deferFn},
		{[]interface{}{1, func(c *Context) {}}, deferFn},
		{[]interface{}{"/foo", func() {}}, deferFn},
		{[]interface{}{"/foo", func(c *Context) {}}, nil},
		{[]interface{}{func(c *Context) {}}, nil},
		{[]interface{}{func(v interface{}, c *Context) {}}, nil},
		{[]interface{}{"/foo", func(v interface{}, c *Context) {}}, nil},
		{[]interface{}{childRouter}, nil},
		{[]interface{}{"/foo", childRouter}, nil},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			router := NewRouter()
			if tt.deferFn != nil {
				defer tt.deferFn()
			}

			router.Use(tt.params...)
			assert.Greater(t, len(router.routes), 0)
		})
	}
}

func TestRouterProxy(t *testing.T) {
	expectedBody, route := "OK", "/foo"
	handle := func(c *Context) {
		c.String(expectedBody)
	}

	t.Run("one-by-one", func(t *testing.T) {
		for _, method := range append(methods, HTTPMethodAll) {
			t.Run(method, func(t *testing.T) {
				assert := assert.New(t)
				router := NewRouter()
				router.Route(route).Handle(method, handle)
				server := httptest.NewServer(router)
				defer server.Close()
				statusCode, _, body, err := request(method, server.URL+route, nil)
				assert.Nil(err)
				assert.Equal(200, statusCode)
				if method != http.MethodHead {
					assert.Equal(expectedBody, body)
				}
			})
		}

		t.Run("with-next", func(t *testing.T) {
			assert := assert.New(t)
			router := NewRouter()
			router.Route(route).ALL(func(c *Context) {
				c.Next()
			}).GET(handle)
			server := httptest.NewServer(router)
			defer server.Close()
			statusCode, _, body, err := request(http.MethodGet, server.URL+route, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)
			assert.Equal(expectedBody, body)
		})

		t.Run("without-next", func(t *testing.T) {
			assert := assert.New(t)
			router := NewRouter()
			router.Route(route).ALL(func(c *Context) {}).GET(handle)
			server := httptest.NewServer(router)
			defer server.Close()
			statusCode, _, body, err := request(http.MethodGet, server.URL+route, nil)
			assert.Nil(err)
			assert.Equal(200, statusCode)
			assert.Equal("", body)
		})
	})

	t.Run("all", func(t *testing.T) {
		h, router := handle, NewRouter()
		router.Route(route).GET(h).HEAD(h).POST(h).PUT(h).PATCH(h).DELETE(h).OPTIONS(h)
		server := httptest.NewServer(router)
		defer server.Close()
		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				assert := assert.New(t)
				statusCode, _, body, err := request(method, server.URL+route, nil)
				assert.Nil(err)
				assert.Equal(200, statusCode)
				if method != http.MethodHead {
					assert.Equal(expectedBody, body)
				}
			})
		}
	})
}

func stringOr(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}
