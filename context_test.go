// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/soongo/soon/renderer"
)

var (
	testErrorFormat = "got `%v`, expect `%v`"
	timeFormat      = http.TimeFormat
	dotRegexp       = regexp.MustCompile("\\s*,\\s*")
)

func TestContext_HeadersSent(t *testing.T) {
	tests := []struct {
		handle   Handle
		expected bool
	}{
		{
			func(c *Context) {
				// do nothing
			},
			false,
		},
		{
			func(c *Context) {
				c.Send("hi")
			},
			true,
		},
		{
			func(c *Context) {
				c.End()
				c.Send("hi")
			},
			false,
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		tt.handle(c)
		if got := c.HeadersSent; got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Locals(t *testing.T) {
	tests := []struct {
		k string
		v interface{}
	}{
		{"string", "hi"},
		{"int", 10},
		{"bool", true},
		{"slice", []string{"foo", "bar"}},
		{"map", map[string]interface{}{"name": "foo", "age": 10}},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Locals.Set(tt.k, tt.v)
		if got := c.Locals.Get(tt.k); !reflect.DeepEqual(got, tt.v) {
			t.Errorf(testErrorFormat, got, tt.v)
		}
	}
}

func TestContext_Params(t *testing.T) {
	tests := []struct {
		params Params
	}{
		{Params{}},
		{Params{"name": "foo", 0: "12"}},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		c := NewContext(req, nil)
		for k, v := range tt.params {
			c.Request.Params.Set(k, v)
		}
		if got := c.Params(); !reflect.DeepEqual(got, tt.params) {
			t.Errorf(testErrorFormat, got, tt.params)
		}
	}
}

func TestContext_Query(t *testing.T) {
	tests := []struct {
		q        string
		expected url.Values
	}{
		{
			"name=foo&age=18",
			url.Values{"name": {"foo"}, "age": {"18"}},
		},
		{
			"name=foo&age=18&name=bar",
			url.Values{"name": {"foo", "bar"}, "age": {"18"}},
		},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/?"+tt.q, nil)
		c := NewContext(req, nil)
		if got := c.Query(); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Append(t *testing.T) {
	tests := []struct {
		k        string
		v        interface{}
		expected interface{}
	}{
		{"Content-Type", "text/html", []string{"text/html; charset=UTF-8"}},
		{"Content-Type", "text/html; charset=UTF-7", []string{"text/html; charset=UTF-7"}},
		{"Content-Type", "application/octet-stream", []string{"application/octet-stream"}},
		{
			"Content-Type",
			[]string{
				"text/html",
				"application/octet-stream",
				"application/json",
				"text/*; charset=UTF-7",
			},
			[]string{
				"text/html; charset=UTF-8",
				"application/octet-stream",
				"application/json; charset=UTF-8",
				"text/*; charset=UTF-7",
			},
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Append(tt.k, tt.v)
		if got := c.Response.Header()[tt.k]; !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.v)
		}
	}
}

func TestContext_Set(t *testing.T) {
	k := "Content-Type"
	tests := []struct {
		k        string
		v        interface{}
		expected interface{}
	}{
		{k, "text/html", http.Header{k: []string{"text/html; charset=UTF-8"}}},
		{k, "text/html; charset=UTF-7", http.Header{k: []string{"text/html; charset=UTF-7"}}},
		{k, "application/octet-stream", http.Header{k: []string{"application/octet-stream"}}},
		{
			k,
			[]string{
				"text/html",
				"application/octet-stream",
				"application/json",
				"text/*; charset=UTF-7",
			},
			http.Header{k: []string{
				"text/html; charset=UTF-8",
				"application/octet-stream",
				"application/json; charset=UTF-8",
				"text/*; charset=UTF-7",
			}},
		},
		{
			"",
			map[string]string{
				k:          "text/html",
				"X-Custom": "custom",
			},
			http.Header{
				k:          []string{"text/html; charset=UTF-8"},
				"X-Custom": []string{"custom"},
			},
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Response.Header().Set(k, "text/*")
		if tt.k == "" {
			c.Set(tt.v)
		} else {
			c.Set(tt.k, tt.v)
		}
		if got := c.Response.Header(); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Get(t *testing.T) {
	tests := []struct {
		k        string
		v        interface{}
		expected string
	}{
		{"Content-Type", "application/json", "application/json; charset=UTF-8"},
		{"Content-Type", "text/*", "text/*; charset=UTF-8"},
		{"Content-Type", "application/octet-stream", "application/octet-stream"},
		{"Content-Type", []string{"text/*", "application/json"}, "text/*; charset=UTF-8"},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		if got := c.Get(tt.k); got != "" {
			t.Errorf(testErrorFormat, got, "")
		}
		c.Set(tt.k, tt.v)
		if got := c.Get(tt.k); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Status(t *testing.T) {
	tests := []struct {
		code int
	}{
		{200},
		{302},
		{404},
		{500},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Status(tt.code)
		w := c.Response.(*httptest.ResponseRecorder)
		if got := w.Code; got != tt.code {
			t.Errorf(testErrorFormat, got, tt.code)
		}
	}
}

func TestContext_SendStatus(t *testing.T) {
	tests := []struct {
		code int
	}{
		{200},
		{302},
		{404},
		{500},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.SendStatus(tt.code)
		w := c.Response.(*httptest.ResponseRecorder)
		if got := w.Code; got != tt.code {
			t.Errorf(testErrorFormat, got, tt.code)
		}
		if got := w.Body.String(); got != http.StatusText(w.Code) {
			t.Errorf(testErrorFormat, got, http.StatusText(w.Code))
		}
	}
}

func TestContext_Type(t *testing.T) {
	tests := []struct {
		t        string
		expected string
	}{
		{"html", "text/html; charset=UTF-8"},
		{"index.html", "text/html; charset=UTF-8"},
		{"image/png", "image/png"},
		{"/", "/"},
		{"", "application/octet-stream"},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Type(tt.t)
		if got := c.Get("Content-Type"); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Attachment(t *testing.T) {
	tests := []struct {
		s        string
		none     bool
		expected string
	}{
		{"foo.png", false, "attachment; filename=\"foo.png\""},
		{"", false, "attachment; filename=\"\""},
		{"", true, "attachment"},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		if tt.none {
			c.Attachment()
		} else {
			c.Attachment(tt.s)
		}
		if got := c.Get("Content-Disposition"); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Cookie(t *testing.T) {
	now := time.Now()
	nowStr := now.UTC().Format(timeFormat)
	tests := []struct {
		cookie   *http.Cookie
		expected string
	}{
		{
			&http.Cookie{Name: "foo", Value: "bar", Path: "/", HttpOnly: true},
			"foo=bar; Path=/; HttpOnly",
		},
		{
			&http.Cookie{
				Name:     "foo",
				Value:    "bar",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				Expires:  now,
			},
			fmt.Sprintf("foo=bar; Path=/; Expires=%s; HttpOnly; Secure", nowStr),
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Cookie(tt.cookie)
		if got := c.Get("Set-Cookie"); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_ClearCookie(t *testing.T) {
	tests := []struct {
		cookie   *http.Cookie
		expected []string
	}{
		{
			&http.Cookie{Name: "foo", Value: "bar", Path: "/", HttpOnly: true},
			[]string{
				"foo=bar; Path=/; HttpOnly",
				fmt.Sprintf("foo=; Path=/; Expires=%s", time.Unix(0, 0).UTC().Format(timeFormat)),
			},
		},
		{
			&http.Cookie{Name: "foo", Value: "bar", Path: ""},
			[]string{
				"foo=bar",
				fmt.Sprintf("foo=; Path=/; Expires=%s", time.Unix(0, 0).UTC().Format(timeFormat)),
			},
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Cookie(tt.cookie)
		if got := c.Get("Set-Cookie"); got != tt.expected[0] {
			t.Errorf(testErrorFormat, got, tt.expected[0])
		}
		c.ClearCookie(tt.cookie)
		if got := c.Response.Header()["Set-Cookie"]; !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_SendFile(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	deferFn := func(w *httptest.ResponseRecorder, code int, e error) {
		err := recover()
		if err == nil {
			t.Errorf(testErrorFormat, nil, "none nil error")
		}
		if e != nil && err != e {
			t.Errorf(testErrorFormat, err, e)
		}
		if w.Code != code {
			t.Errorf(testErrorFormat, w.Code, code)
		}
		if got := w.Header()["Content-Type"]; len(got) != 0 {
			t.Errorf(testErrorFormat, got, nil)
		}
	}

	maxAge := time.Hour
	tests := []struct {
		name                string
		filePath            string
		options             *renderer.FileOptions
		expectedStatus      int
		expectedContentType string
		expectedError       error
		deferFn             func(w *httptest.ResponseRecorder, code int, e error)
	}{
		{
			"normal-1",
			path.Join(pwd, "README.md"),
			nil,
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{
			"normal-2",
			path.Join(pwd, "README.md"),
			&renderer.FileOptions{
				MaxAge: &maxAge,
				Header: map[string]string{
					"Accept-Charset":  "utf-8",
					"Accept-Language": "en;q=0.5, zh;q=0.8",
				},
			},
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{"empty-filepath", "", nil, 500, "", nil, deferFn},
		{
			"with-root-path",
			"README.md",
			&renderer.FileOptions{Root: pwd, LastModifiedDisabled: true},
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{"not-root-filepath", "README.md", nil, 500, "", nil, deferFn},
		{"directory", pwd, nil, 500, "", renderer.ErrIsDir, deferFn},
		{
			"hidden-default",
			path.Join(pwd, ".travis.yml"),
			nil,
			500,
			"",
			renderer.ErrNotFound,
			deferFn,
		},
		{
			"hidden-allow",
			path.Join(pwd, ".travis.yml"),
			&renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyAllow},
			200,
			"text/yaml; charset=UTF-8",
			nil,
			nil,
		},
		{
			"hidden-deny",
			path.Join(pwd, ".travis.yml"),
			&renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyDeny},
			500,
			"",
			renderer.ErrForbidden,
			deferFn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContext(nil, httptest.NewRecorder())
			w := c.Response.(*httptest.ResponseRecorder)
			if tt.deferFn != nil {
				func() {
					defer tt.deferFn(w, tt.expectedStatus, tt.expectedError)
					c.SendFile(tt.filePath, tt.options)
				}()
			} else {
				c.SendFile(tt.filePath, tt.options)
				fileInfo, fileContent := getFileContent(tt.filePath)
				lastModified := fileInfo.ModTime().UTC().Format(timeFormat)
				if got := w.Code; got != tt.expectedStatus {
					t.Errorf(testErrorFormat, got, tt.expectedContentType)
				}
				if got := w.Body.String(); got != fileContent {
					t.Errorf(testErrorFormat, got, fileContent)
				}
				if got := c.Get("Content-Type"); got != tt.expectedContentType {
					t.Errorf(testErrorFormat, got, tt.expectedContentType)
				}
				if tt.options != nil {
					if tt.options.MaxAge != nil {
						cc := fmt.Sprintf("max-age=%.0f", maxAge.Seconds())
						if got := c.Get("Cache-Control"); got != cc {
							t.Errorf(testErrorFormat, got, cc)
						}
					}
					if tt.options.Header != nil {
						for k, v := range tt.options.Header {
							if got := c.Get(k); got != v {
								t.Errorf(testErrorFormat, got, v)
							}
						}
					}
					expectedLastModified := lastModified
					if tt.options.LastModifiedDisabled {
						expectedLastModified = ""
					}
					if got := c.Get("Last-Modified"); got != expectedLastModified {
						t.Errorf(testErrorFormat, got, expectedLastModified)
					}
				} else {
					if got := c.Get("Last-Modified"); got != lastModified {
						t.Errorf(testErrorFormat, got, lastModified)
					}
				}
			}
		})
	}
}

func TestContext_Download(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	tests := []struct {
		filePath       string
		options        *renderer.FileOptions
		expectedStatus int
	}{
		{path.Join(pwd, "README.md"), nil, 200},
		{path.Join(pwd, "README.md"), &renderer.FileOptions{Name: "custom-name"}, 200},
	}

	for _, tt := range tests {
		c := &Context{Response: httptest.NewRecorder()}
		c.Download(tt.filePath, tt.options)
		w := c.Response.(*httptest.ResponseRecorder)
		if got := w.Code; got != tt.expectedStatus {
			t.Errorf(testErrorFormat, got, tt.expectedStatus)
		}
		fileInfo, fileContent := getFileContent(tt.filePath)
		if got := w.Body.String(); got != fileContent {
			t.Errorf(testErrorFormat, got, fileContent)
		}
		name := fileInfo.Name()
		if tt.options != nil && tt.options.Name != "" {
			name = tt.options.Name
		}
		contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", name)
		if got := c.Get("Content-Disposition"); got != contentDisposition {
			t.Errorf(testErrorFormat, got, contentDisposition)
		}
	}
}

func TestContext_End(t *testing.T) {
	tests := []struct {
		name     string
		handle   Handle
		expected string
	}{
		{
			"without-end",
			func(c *Context) {
				c.Send("foo")
				c.Send("bar")
			},
			"foobar",
		},
		{
			"with-end",
			func(c *Context) {
				c.Send("foo")
				c.End()
				c.Send("bar")
			},
			"foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{Response: httptest.NewRecorder()}
			tt.handle(c)
			w := c.Response.(*httptest.ResponseRecorder)
			if got := w.Body.String(); got != tt.expected {
				t.Errorf(testErrorFormat, got, tt.expected)
			}
		})
	}
}

func TestContext_Format(t *testing.T) {
	handles := map[string]Handle{
		"text/plain": func(c *Context) {
			c.Send("hey")
		},
		"text/html": func(c *Context) {
			c.Send("<p>hey</p>")
		},
		"application/json": func(c *Context) {
			c.Json(struct {
				Message string `json:"message"`
			}{"hey"})
		},
	}
	handlesWithDefault := map[string]Handle{
		"default": func(c *Context) {
			c.Send("hey")
		},
	}
	for k, v := range handles {
		handlesWithDefault[k] = v
	}
	extensionHandles := map[string]Handle{
		"text": func(c *Context) {
			c.Send("hey")
		},
		"html": func(c *Context) {
			c.Send("<p>hey</p>")
		},
		"json": func(c *Context) {
			c.Json(struct {
				Message string `json:"message"`
			}{"hey"})
		},
	}
	extensionHandlesWithDefault := map[string]Handle{
		"default": func(c *Context) {
			c.Send("hey")
		},
	}
	for k, v := range extensionHandles {
		extensionHandlesWithDefault[k] = v
	}
	type testObj struct {
		accept         string
		format         map[string]Handle
		expectedStatus int
		expectedBody   string
	}
	tests := []testObj{
		{"text/html", handles, 200, "<p>hey</p>"},
		{"application/xml", handles, 406, http.StatusText(406)},
		{"*/*", handles, 200, `{"message":"hey"}`},
		{"application/*", handles, 200, `{"message":"hey"}`},
		{"application/json", handlesWithDefault, 200, `{"message":"hey"}`},
		{"application/xml", handlesWithDefault, 200, "hey"},
		{"*/*", handlesWithDefault, 200, `{"message":"hey"}`},
		{"application/*", handlesWithDefault, 200, `{"message":"hey"}`},
		{"text/html", extensionHandles, 200, "<p>hey</p>"},
		{"application/xml", extensionHandles, 406, http.StatusText(406)},
		{"*/*", extensionHandles, 200, "<p>hey</p>"},
		{"application/*", extensionHandles, 200, `{"message":"hey"}`},
		{"application/json", extensionHandlesWithDefault, 200, `{"message":"hey"}`},
		{"application/xml", extensionHandlesWithDefault, 200, "hey"},
		{"*/*", extensionHandlesWithDefault, 200, "<p>hey</p>"},
		{"application/*", extensionHandlesWithDefault, 200, `{"message":"hey"}`},
	}

	app := New()
	server := httptest.NewServer(app)
	defer server.Close()

	for i, tt := range tests {
		path := "/" + strconv.Itoa(i)
		app.GET(path, func(tt testObj) func(c *Context) {
			return func(c *Context) {
				c.Format(tt.format)
			}
		}(tt))
	}

	for i, tt := range tests {
		path := "/" + strconv.Itoa(i)
		header := http.Header{"Accept": dotRegexp.Split(tt.accept, -1)}
		statusCode, _, body, err := request(http.MethodGet, server.URL+path, header)
		body = strings.Trim(body, "\n")
		if err != nil {
			t.Error(err)
		}
		if got := statusCode; got != tt.expectedStatus {
			t.Errorf(testErrorFormat, got, tt.expectedStatus)
		}
		if got := body; got != tt.expectedBody {
			t.Errorf(testErrorFormat, got, tt.expectedBody)
		}
	}
}

func TestContext_String(t *testing.T) {
	tests := []struct {
		s                   string
		expectedStatus      int
		expectedContentType string
	}{
		{"foo", 200, "text/plain; charset=utf-8"},
	}

	for _, tt := range tests {
		c := &Context{Response: httptest.NewRecorder()}
		c.String(tt.s)
		w := c.Response.(*httptest.ResponseRecorder)
		if got := w.Code; got != tt.expectedStatus {
			t.Errorf(testErrorFormat, got, tt.expectedStatus)
		}
		if got := c.Get("Content-Type"); got != tt.expectedContentType {
			t.Errorf(testErrorFormat, got, tt.expectedContentType)
		}
		if got := w.Body.String(); got != tt.s {
			t.Errorf(testErrorFormat, got, tt.s)
		}
	}
}

func TestContext_Json(t *testing.T) {
	contentType := "application/json; charset=utf-8"
	tests := []struct {
		handle              Handle
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			func(c *Context) {
				book := struct {
					Name      string
					PageTotal uint16
				}{"foo", 50}
				c.Json(book)
			},
			200,
			contentType,
			`{"Name":"foo","PageTotal":50}`,
		},
		{
			func(c *Context) {
				book := struct {
					Name      string `json:"name"`
					PageTotal uint16 `json:"pageTotal"`
				}{"foo", 50}
				c.Json(book)
			},
			200,
			contentType,
			`{"name":"foo","pageTotal":50}`,
		},
		{
			func(c *Context) {
				c.Json([]string{"foo", "bar"})
			},
			200,
			contentType,
			`["foo","bar"]`,
		},
		{
			func(c *Context) {
				books := []struct {
					Name      string `json:"name"`
					PageTotal uint16 `json:"pageTotal"`
				}{{"foo", 50}, {"bar", 20}}
				c.Json(books)
			},
			200,
			contentType,
			`[{"name":"foo","pageTotal":50},{"name":"bar","pageTotal":20}]`,
		},
	}

	for _, tt := range tests {
		c := &Context{Response: httptest.NewRecorder()}
		tt.handle(c)
		w := c.Response.(*httptest.ResponseRecorder)
		if got := w.Code; got != tt.expectedStatus {
			t.Errorf(testErrorFormat, got, tt.expectedStatus)
		}
		if got := c.Get("Content-Type"); got != tt.expectedContentType {
			t.Errorf(testErrorFormat, got, tt.expectedContentType)
		}
		buf := bytes.NewBuffer([]byte(tt.expectedBody))
		buf.WriteByte('\n')
		expectedBody := buf.String()
		if got := w.Body.String(); got != expectedBody {
			t.Errorf(testErrorFormat, got, expectedBody)
		}
	}
}

func TestContext_Render(t *testing.T) {
	tests := []struct {
		name                string
		renderer            renderer.Renderer
		expectedStatus      int
		expectedContentType string
		expectedBody        string
		deferFn             func(w *httptest.ResponseRecorder, code int, contentType string)
	}{
		{
			"string",
			renderer.String{Data: "foo"},
			200,
			"text/plain; charset=utf-8",
			"foo",
			nil,
		},
		{
			"json",
			renderer.JSON{Data: struct {
				Name      string
				PageTotal uint16
			}{"foo", 50}},
			200,
			"application/json; charset=utf-8",
			`{"Name":"foo","PageTotal":50}`,
			nil,
		},
		{
			"json-custom",
			renderer.JSON{Data: struct {
				Name      string `json:"name"`
				PageTotal uint16 `json:"pageTotal"`
			}{"foo", 50}},
			200,
			"application/json; charset=utf-8",
			`{"name":"foo","pageTotal":50}`,
			nil,
		},
		{
			"json-error",
			renderer.JSON{Data: func() {}},
			500,
			"application/json; charset=utf-8",
			"",
			func(w *httptest.ResponseRecorder, code int, contentType string) {
				if got := recover(); got == nil {
					t.Errorf(testErrorFormat, got, "none nil error")
				}
				if got := w.Code; got != code {
					t.Errorf(testErrorFormat, got, code)
				}
				if got := w.Header().Get("Content-Type"); got != contentType {
					t.Errorf(testErrorFormat, got, contentType)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Context{Response: httptest.NewRecorder()}
			w := c.Response.(*httptest.ResponseRecorder)
			if tt.deferFn != nil {
				defer tt.deferFn(w, tt.expectedStatus, tt.expectedContentType)
				c.Render(tt.renderer)
			} else {
				c.Render(tt.renderer)
				if got := w.Code; got != tt.expectedStatus {
					t.Errorf(testErrorFormat, got, tt.expectedStatus)
				}
				if got := c.Get("Content-Type"); got != tt.expectedContentType {
					t.Errorf(testErrorFormat, got, tt.expectedContentType)
				}
				expectedBody := tt.expectedBody
				if _, ok := tt.renderer.(renderer.JSON); ok {
					buf := bytes.NewBuffer([]byte(tt.expectedBody))
					buf.WriteByte('\n')
					expectedBody = buf.String()
				}
				if got := w.Body.String(); got != expectedBody {
					t.Errorf(testErrorFormat, got, expectedBody)
				}
			}
		})
	}
}

func getFileContent(p string) (os.FileInfo, string) {
	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	return fileInfo, string(bts)
}
