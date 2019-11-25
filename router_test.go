// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type header map[string]string

type test struct {
	route      string
	path       string
	statusCode int
	header     header
	body       string
}

const system404Body = "404 page not found\n"

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
	return func(req *http.Request, res *Response, next func()) {
		for k, v := range t.header {
			res.Header().Set(k, v)
		}
		if t.statusCode != 0 {
			res.WriteHeader(t.statusCode)
		}
		if t.body != "" {
			if err := res.Send(t.body); err != nil {
				panic(err)
			}
		}
	}
}

func request(method, url string) (statusCode int, header http.Header, body string, err error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	statusCode = resp.StatusCode
	header = resp.Header

	defer resp.Body.Close()

	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	body = string(byteArray)

	return
}

func getWantStatusCode(statusCode int) int {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	return statusCode
}

func notFound(req *http.Request, res *Response, next func()) {
	http.Error(res, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func TestRouterWithDefaultOptions(t *testing.T) {
	tests := []test{
		{route: "/", path: "/", body: "root page"},
		{route: "/", path: "", statusCode: http.StatusOK, body: "root page"},
		{route: "//", path: "/", statusCode: http.StatusNotFound, body: system404Body},
		{route: "//", path: "//", body: "root page"},
		{route: "//", path: "///", statusCode: http.StatusNotFound, body: system404Body},

		{route: "", path: "/", body: "root page"},
		{route: "", path: "", body: "root page"},

		{route: "/health-check/", path: "/health-check/", statusCode: http.StatusOK,
			header: header{"Content-Type": "application/json"}, body: `{"alive": true}`},
		{route: "/health-check/", path: "/health-check", statusCode: http.StatusNotFound,
			body: system404Body},
		{route: "/health-check//", path: "/health-check/", statusCode: http.StatusNotFound,
			body: system404Body},
		{route: "/health-check//", path: "/health-check//", body: `{"alive": true}`},
		{route: "/health-check//", path: "/health-check///", statusCode: http.StatusNotFound,
			body: system404Body},

		{route: "/health-check", path: "/health-check/", body: `{"alive": true}`},
		{route: "/health-check", path: "/health-check", body: `{"alive": true}`},
	}

	t.Run("one-by-one", func(t *testing.T) {
		for _, tt := range tests {
			tt := tt
			t.Run("", func(t *testing.T) {
				router := NewRouter()
				router.GET(tt.route, makeHandle(tt))
				server := httptest.NewServer(router)
				defer server.Close()

				statusCode, header, body, err := request(http.MethodGet, server.URL+tt.path)
				if err != nil {
					t.Error(err)
				}
				wantStatusCode := getWantStatusCode(tt.statusCode)
				if statusCode != wantStatusCode {
					t.Errorf("handler returned wrong status code: got %v want %v",
						statusCode, wantStatusCode)
				}
				for k, v := range tt.header {
					if header.Get(k) != v {
						t.Errorf("handler returned unexpected header of [%v]: got '%v' want '%v'",
							k, header.Get(k), v)
					}
				}
				if body != tt.body {
					t.Errorf("handler returned unexpected body: got '%v' want '%v'",
						body, tt.body)
				}

				statusCode, _, _, err = request(http.MethodHead, server.URL+tt.path)
				if err != nil {
					t.Error(err)
				}
				if statusCode != http.StatusNotFound {
					t.Errorf("handler returned wrong status code: got %v want %v",
						statusCode, http.StatusNotFound)
				}
			})
		}
	})

	tests = []test{
		{route: "/", path: "/", body: "root page"},
		{route: "/", path: "", statusCode: http.StatusOK, body: "root page"},
		{route: "//", path: "/", statusCode: http.StatusOK, body: "root page"},
		{route: "//", path: "//", body: "root page"},
		{route: "//", path: "///", statusCode: http.StatusOK, body: "root page"},

		{route: "", path: "/", body: "root page"},
		{route: "", path: "", body: "root page"},

		{route: "/health-check/", path: "/health-check/", statusCode: http.StatusOK,
			header: header{"Content-Type": "application/json"}, body: `{"alive": true}`},
		{route: "/health-check/", path: "/health-check", statusCode: http.StatusOK,
			body: `{"alive": true}`},
		{route: "/health-check//", path: "/health-check/", statusCode: http.StatusOK,
			body: `{"alive": true}`},
		{route: "/health-check//", path: "/health-check//", body: `{"alive": true}`},
		{route: "/health-check//", path: "/health-check///", statusCode: http.StatusOK,
			body: `{"alive": true}`},

		{route: "/health-check", path: "/health-check/", body: `{"alive": true}`},
		{route: "/health-check", path: "/health-check", body: `{"alive": true}`},
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
				statusCode, header, body, err := request(http.MethodGet, server.URL+tt.path)
				if err != nil {
					t.Error(err)
				}
				wantStatusCode := getWantStatusCode(tt.statusCode)
				if statusCode != wantStatusCode {
					t.Errorf("handler returned wrong status code: got %v want %v",
						statusCode, wantStatusCode)
				}
				for k, v := range tt.header {
					if header.Get(k) != v {
						t.Errorf("handler returned unexpected header of [%v]: got '%v' want '%v'",
							k, header.Get(k), v)
					}
				}
				if body != tt.body {
					t.Errorf("handler returned unexpected body: got '%v' want '%v'",
						body, tt.body)
				}
			})
		}
	})
}

func TestRouterWithTrailingSlashPolicyNone(t *testing.T) {
	tests := []test{
		{route: "/", path: "/", body: "root page"},
		{route: "/", path: "", body: "root page"},
		{route: "//", path: "/", statusCode: http.StatusNotFound, body: system404Body},
		{route: "//", path: "", statusCode: http.StatusNotFound, body: system404Body},

		{route: "", path: "/", body: "root page"},
		{route: "", path: "", body: "root page"},
		{route: "//", path: "/", statusCode: http.StatusNotFound, body: system404Body},
		{route: "//", path: "", statusCode: http.StatusNotFound, body: system404Body},

		{route: "/health-check/", path: "/health-check/", body: `{"alive": true}`},
		{route: "/health-check/", path: "/health-check", statusCode: http.StatusNotFound,
			body: system404Body},
		{route: "/health-check//", path: "/health-check/", statusCode: http.StatusNotFound,
			body: system404Body},
		{route: "/health-check//", path: "/health-check", statusCode: http.StatusNotFound,
			body: system404Body},

		{route: "/health-check", path: "/health-check/", statusCode: http.StatusNotFound,
			body: system404Body},
		{route: "/health-check", path: "/health-check", body: `{"alive": true}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			router := NewRouter()
			router.GET(tt.route, makeHandle(tt))
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request(http.MethodGet, server.URL+tt.path)
			if err != nil {
				t.Error(err)
			}
			wantStatusCode := getWantStatusCode(tt.statusCode)
			if statusCode != wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					statusCode, wantStatusCode)
			}
			if body != tt.body {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					body, tt.body)
			}
		})
	}
}

func TestRouterWithCustomNotFound(t *testing.T) {
	tests := []test{
		{route: "/", path: "/", body: "root page"},
		{route: "/", path: "/404", statusCode: http.StatusNotFound,
			body: fmt.Sprintln(http.StatusText(http.StatusNotFound))},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			router := NewRouter()
			router.GET(tt.route, makeHandle(tt))
			router.Use(notFound)
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request(http.MethodGet, server.URL+tt.path)
			if err != nil {
				t.Error(err)
			}
			wantStatusCode := getWantStatusCode(tt.statusCode)
			if statusCode != wantStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					statusCode, wantStatusCode)
			}
			if body != tt.body {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					body, tt.body)
			}
		})
	}
}

func TestRouterMiddleware(t *testing.T) {
	test := test{route: "/", path: "/", body: "root page"}
	middlewareBody := "middleware"
	router := NewRouter()
	router.Use(func(req *http.Request, res *Response, next func()) {
		if err := res.Send(middlewareBody); err != nil {
			panic(err)
		}
	})
	router.GET(test.route, makeHandle(test))
	server := httptest.NewServer(router)
	defer server.Close()

	statusCode, _, body, err := request(http.MethodGet, server.URL+test.path)
	if err != nil {
		t.Error(err)
	}
	if statusCode != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			statusCode, http.StatusOK)
	}
	if body != middlewareBody {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'",
			body, middlewareBody)
	}
}

func TestRouterMethods(t *testing.T) {
	for i, method := range methods {
		t.Run(method, func(t *testing.T) {
			test := test{route: "/", path: "/", body: "root page"}
			router := NewRouter()
			router.Handle(method, test.route, makeHandle(test))
			server := httptest.NewServer(router)
			defer server.Close()

			statusCode, _, body, err := request(method, server.URL+test.path)
			if err != nil {
				t.Error(err)
			}
			if statusCode != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					statusCode, http.StatusOK)
			}
			if method != http.MethodHead && body != test.body {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					body, test.body)
			}

			method = methods[(i+1)%len(methods)]
			statusCode, _, body, err = request(method, server.URL+test.path)
			if err != nil {
				t.Error(err)
			}
			if statusCode != http.StatusNotFound {
				t.Errorf("handler returned wrong status code: got %v want %v",
					statusCode, http.StatusNotFound)
			}
			if method != http.MethodHead && body != system404Body {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					body, system404Body)
			}
		})
	}
}

func TestRouterAll(t *testing.T) {
	test := test{route: "/", path: "/", body: "root page"}
	router := NewRouter()
	router.ALL(test.route, makeHandle(test))
	server := httptest.NewServer(router)
	defer server.Close()

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			statusCode, _, body, err := request(method, server.URL+test.path)
			if err != nil {
				t.Error(err)
			}
			if statusCode != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					statusCode, http.StatusOK)
			}
			if method != http.MethodHead && body != test.body {
				t.Errorf("handler returned unexpected body: got '%v' want '%v'",
					body, test.body)
			}
		})
	}
}
