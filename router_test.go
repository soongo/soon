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
)

type header map[string]string

type test struct {
	route       string
	path        string
	statusCode  int
	header      header
	body        string
	middleware  func(c *Context)
	err         interface{}
	errorHandle func(v interface{}, c *Context)
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

func TestRouterWithDefaultOptions(t *testing.T) {
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
		{route: "/health-check/", path: "/health-check", statusCode: 404, body: body404},
		{route: "/health-check//", path: "/health-check/", statusCode: 404, body: body404},
		{route: "/health-check//", path: "/health-check//", statusCode: 200, body: body200},
		{route: "/health-check//", path: "/health-check///", statusCode: 404, body: body404},
		{route: "/health-check", path: "/health-check/", statusCode: 200, body: body200},
		{route: "/health-check", path: "/health-check", statusCode: 200, body: body200},
	}

	t.Run("one-by-one", func(t *testing.T) {
		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				router := NewRouter()
				router.GET(tt.route, makeHandle(tt))
				server := httptest.NewServer(router)
				defer server.Close()

				statusCode, header, body, err := request("GET", server.URL+tt.path, nil)
				if err != nil {
					t.Error(err)
				}
				if statusCode != tt.statusCode {
					t.Errorf(testErrorFormat, statusCode, tt.statusCode)
				}
				for k, v := range tt.header {
					if header.Get(k) != v {
						t.Errorf(testErrorFormat, header.Get(k), v)
					}
				}
				if body != tt.body {
					t.Errorf(testErrorFormat, body, tt.body)
				}

				statusCode, _, _, err = request("HEAD", server.URL+tt.path, nil)
				if err != nil {
					t.Error(err)
				}
				if statusCode != 404 {
					t.Errorf(testErrorFormat, statusCode, 404)
				}
			})
		}
	})

	tests = []test{
		{route: "/", path: "/", statusCode: 200, body: body200},
		{route: "/", path: "", statusCode: 200, body: body200},
		{route: "//", path: "/", statusCode: 200, body: body200},
		{route: "//", path: "//", statusCode: 200, body: body200},
		{route: "//", path: "///", statusCode: 200, body: body200},
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
		{route: "/health-check//", path: "/health-check///", statusCode: 200, body: body200},
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
				statusCode, header, body, err := request("GET", server.URL+tt.path, nil)
				if err != nil {
					t.Error(err)
				}
				if statusCode != tt.statusCode {
					t.Errorf(testErrorFormat, statusCode, tt.statusCode)
				}
				for k, v := range tt.header {
					if header.Get(k) != v {
						t.Errorf(testErrorFormat, header.Get(k), v)
					}
				}
				if body != tt.body {
					t.Errorf(testErrorFormat, body, tt.body)
				}
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
			router := NewRouter()
			if tt.middleware != nil {
				router.Use(tt.middleware)
			}
			router.GET(tt.route, makeHandle(tt))
			if tt.errorHandle != nil {
				router.Use(tt.errorHandle)
			}
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request("GET", server.URL+tt.path, nil)
			if err != nil {
				t.Error(err)
			}

			if statusCode != tt.statusCode {
				t.Errorf(testErrorFormat, statusCode, tt.statusCode)
			}

			if body != body200 {
				t.Errorf(testErrorFormat, body, body200)
			}
		})
	}
}

func TestRouterMethods(t *testing.T) {
	tt := test{route: "/", path: "/", body: body200}
	check := func(method, url string) {
		statusCode, _, body, err := request(method, url, nil)
		if err != nil {
			t.Error(err)
		}
		if statusCode != 200 {
			t.Errorf(testErrorFormat, statusCode, 200)
		}
		if method != "HEAD" && body != tt.body {
			t.Errorf(testErrorFormat, body, tt.body)
		}

		for _, m := range methods {
			if m != method {
				statusCode, _, body, err = request(m, url, nil)
				if err != nil {
					t.Error(err)
				}
				if statusCode != 404 {
					t.Errorf(testErrorFormat, statusCode, 404)
				}
				if m != "HEAD" && body != body404 {
					t.Errorf(testErrorFormat, body, body404)
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
			tt := test{route: "/", path: "/", body: body200}
			router := NewRouter()
			router.Handle(method, tt.route, makeHandle(tt))
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request(method, server.URL+tt.path, nil)
			if err != nil {
				t.Error(err)
			}
			if statusCode != 200 {
				t.Errorf(testErrorFormat, statusCode, 200)
			}
			if method != "HEAD" && body != tt.body {
				t.Errorf(testErrorFormat, body, tt.body)
			}

			method = methods[(i+1)%len(methods)]
			statusCode, _, body, err = request(method, server.URL+tt.path, nil)
			if err != nil {
				t.Error(err)
			}
			if statusCode != 404 {
				t.Errorf(testErrorFormat, statusCode, 404)
			}
			if method != "HEAD" && body != body404 {
				t.Errorf(testErrorFormat, body, body404)
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
			statusCode, _, body, err := request(method, server.URL+tt.path, nil)
			if err != nil {
				t.Error(err)
			}
			if statusCode != 200 {
				t.Errorf(testErrorFormat, statusCode, 200)
			}
			if method != "HEAD" && body != tt.body {
				t.Errorf(testErrorFormat, body, tt.body)
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
			router := NewRouter()
			router.Handle(tt.method, tt.route, func(c *Context) {
				if got := c.Locals().Get(tt.paramName); got != tt.paramValue {
					t.Errorf(testErrorFormat, got, tt.paramValue)
				}
			})
			if tt.err != nil {
				router.Use(func(v interface{}, c *Context) {
					if got := c.Locals().Get(tt.paramName); got != nil {
						t.Errorf(testErrorFormat, got, nil)
					}
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
			if err != nil {
				t.Error(err)
			}
		})
	}

	t.Run("once", func(t *testing.T) {
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
		if err != nil {
			t.Error(err)
		}
		if calledCount != 1 {
			t.Errorf(testErrorFormat, calledCount, 1)
		}
		n := 2
		for i := 0; i < n; i++ {
			_, _, _, err = request("GET", server.URL+"/name/bar", nil)
			if err != nil {
				t.Error(err)
			}
		}
		if calledCount != n+1 {
			t.Errorf(testErrorFormat, calledCount, n+1)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		router, calledCount := NewRouter(), 0
		router.Use("/:foo", func(c *Context) {
			if c.Locals().Get("foo") != "name" {
				t.Errorf(testErrorFormat, c.Locals().Get("foo"), "name")
			}
			c.Next()
		})
		router.GET("/name/:foo/:bar", func(c *Context) {
			if c.Locals().Get("foo") != "id" {
				t.Errorf(testErrorFormat, c.Locals().Get("foo"), "id")
			}
			c.Next()
		})
		router.GET("/name/:foo/:bar", func(c *Context) {
			if c.Locals().Get("foo") != "id" {
				t.Errorf(testErrorFormat, c.Locals().Get("foo"), "id")
			}
			c.Next()
		})
		router.GET("/name/id/:foo", func(c *Context) {
			if c.Locals().Get("foo") != "bar" {
				t.Errorf(testErrorFormat, c.Locals().Get("foo"), "bar")
			}
			c.String("foo")
		})
		router.Param("foo", func(r *Request, s string) {
			r.Locals.Set("foo", s)
			calledCount++
		})
		server := httptest.NewServer(router)
		defer server.Close()
		_, _, _, err := request("GET", server.URL+"/name/id/bar", nil)
		if err != nil {
			t.Error(err)
		}
		if calledCount != 3 {
			t.Errorf(testErrorFormat, calledCount, 3)
		}
		n := 2
		for i := 0; i < n; i++ {
			_, _, _, err = request("GET", server.URL+"/name/id/bar", nil)
			if err != nil {
				t.Error(err)
			}
		}
		if calledCount != n*3+3 {
			t.Errorf(testErrorFormat, calledCount, n*3+3)
		}
	})

	t.Run("sub-router", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
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
			if err != nil {
				t.Error(err)
			}
			if calledCount != 0 {
				t.Errorf(testErrorFormat, calledCount, 0)
			}
		})

		t.Run("", func(t *testing.T) {
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
			if err != nil {
				t.Error(err)
			}
			if calledCount != 1 {
				t.Errorf(testErrorFormat, calledCount, 1)
			}
		})

		t.Run("", func(t *testing.T) {
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
			if err != nil {
				t.Error(err)
			}
			if calledCount != 2 {
				t.Errorf(testErrorFormat, calledCount, 2)
			}
		})

		t.Run("", func(t *testing.T) {
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
			if err != nil {
				t.Error(err)
			}
			if calledCount != 4 {
				t.Errorf(testErrorFormat, calledCount, 4)
			}
		})
	})
}

func TestRouter_Use(t *testing.T) {
	deferFn := func() {
		if err := recover(); err == nil {
			t.Errorf(testErrorFormat, err, "none nil error")
		}
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
			if got := len(router.routes); got == 0 {
				t.Errorf(testErrorFormat, got, ">0")
			}
		})
	}
}
