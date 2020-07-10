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
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/soongo/soon/util"

	"github.com/soongo/soon/renderer"
)

var (
	testErrorFormat = "got `%v`, expect `%v`"
	timeFormat      = http.TimeFormat
	dotRegexp       = regexp.MustCompile("\\s*,\\s*")
	plainType       = "text/plain; charset=UTF-8"
	htmlType        = "text/html; charset=UTF-8"
	jsonType        = "application/json; charset=UTF-8"
	jsonpType       = "text/javascript; charset=UTF-8"
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
			},
			true,
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		tt.handle(c)
		if got := c.HeadersSent(); got != tt.expected {
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
		c.Locals().Set(tt.k, tt.v)
		if got := c.Locals().Get(tt.k); !reflect.DeepEqual(got, tt.v) {
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
		{"Content-Type", "text/html", []string{htmlType}},
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
				htmlType,
				"application/octet-stream",
				jsonType,
				"text/*; charset=UTF-7",
			},
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Append(tt.k, tt.v)
		if got := c.Writer.Header()[tt.k]; !reflect.DeepEqual(got, tt.expected) {
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
		{k, "text/html", http.Header{k: []string{htmlType}}},
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
				htmlType,
				"application/octet-stream",
				jsonType,
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
				k:          []string{htmlType},
				"X-Custom": []string{"custom"},
			},
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Writer.Header().Set(k, "text/*")
		if tt.k == "" {
			c.Set(tt.v)
		} else {
			c.Set(tt.k, tt.v)
		}
		if got := c.Writer.Header(); !reflect.DeepEqual(got, tt.expected) {
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
		{"Content-Type", "application/json", jsonType},
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

func TestContext_Vary(t *testing.T) {
	tests := []struct {
		vary     string
		fields   []string
		expected string
	}{
		{"", []string{"Accept-Encoding"}, "Accept-Encoding"},
		{"Accept-Encoding", []string{"Accept-Encoding"}, "Accept-Encoding"},
		{"Accept-Encoding", []string{"Host"}, "Accept-Encoding, Host"},
		{"Accept-Encoding, Host", []string{"Accept-Encoding", "Host"}, "Accept-Encoding, Host"},
		{"Accept-Encoding, Host", []string{"Host", "User-Agent"}, "Accept-Encoding, Host, User-Agent"},
	}
	key := "Vary"
	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Set(key, tt.vary)
		c.Vary(tt.fields...)
		result := c.Get(key)
		if result != tt.expected {
			t.Errorf(testErrorFormat, result, tt.expected)
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

		if got := c.Writer.Status(); got != tt.code {
			t.Errorf(testErrorFormat, got, tt.code)
		}

		c.Writer.Flush()
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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
		{"html", htmlType},
		{"index.html", htmlType},
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

func TestContext_Links(t *testing.T) {
	tests := []struct {
		origin   string
		links    map[string]string
		expected string
	}{
		{
			"",
			map[string]string{
				"next": "http://api.example.com/users?page=2",
				"last": "http://api.example.com/users?page=5",
			},
			"<http://api.example.com/users?page=2>; rel=\"next\", " +
				"<http://api.example.com/users?page=5>; rel=\"last\"",
		},
		{
			"<http://api.example.com/users?page=1>; rel=\"pre\"",
			map[string]string{
				"next": "http://api.example.com/users?page=2",
				"last": "http://api.example.com/users?page=5",
			},
			"<http://api.example.com/users?page=1>; rel=\"pre\", " +
				"<http://api.example.com/users?page=2>; rel=\"next\", " +
				"<http://api.example.com/users?page=5>; rel=\"last\"",
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.Set("Link", tt.origin)
		c.Links(tt.links)
		got := c.Get("Link")
		arr1, arr2 := strings.Split(got, ", "), strings.Split(tt.expected, ", ")
		if !stringSliceEquals(arr1, arr2) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestContext_Location(t *testing.T) {
	tests := []struct {
		location string
		referrer string
		expected string
	}{
		{location: "http://example.com"},
		{location: "http://example.com/café"},
		{location: "/foo/bar"},
		{location: " /foo/bar "},
		{location: "back", expected: "/"},
		{location: "back", referrer: "/previous", expected: "/previous"},
	}
	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		if tt.expected == "" {
			tt.expected = util.EncodeURI(strings.Trim(tt.location, " "))
		}
		if tt.referrer != "" {
			c.Set("Referrer", tt.referrer)
		}
		c.Location(tt.location)
		if got := c.Get("location"); got != tt.expected {
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
		if got := c.Writer.Header()["Set-Cookie"]; !reflect.DeepEqual(got, tt.expected) {
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
			t.Errorf(testErrorFormat, nil, "error")
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
		{"empty-filepath", "", nil, 200, "", nil, deferFn},
		{
			"with-root-path",
			"README.md",
			&renderer.FileOptions{Root: pwd, LastModifiedDisabled: true},
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{"not-root-filepath", "README.md", nil, 200, "", nil, deferFn},
		{"directory", pwd, nil, 200, "", renderer.ErrIsDir, deferFn},
		{
			"hidden-default",
			path.Join(pwd, ".travis.yml"),
			nil,
			200,
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
			200,
			"",
			renderer.ErrForbidden,
			deferFn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContext(nil, httptest.NewRecorder())
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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
		c := NewContext(nil, httptest.NewRecorder())
		c.Download(tt.filePath, tt.options)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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
			c := NewContext(nil, httptest.NewRecorder())
			tt.handle(c)
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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
		{"foo", 200, plainType},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		c.String(tt.s)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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
			jsonType,
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
			jsonType,
			`{"name":"foo","pageTotal":50}`,
		},
		{
			func(c *Context) {
				c.Json([]string{"foo", "bar"})
			},
			200,
			jsonType,
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
			jsonType,
			`[{"name":"foo","pageTotal":50},{"name":"bar","pageTotal":20}]`,
		},
	}

	for _, tt := range tests {
		c := NewContext(nil, httptest.NewRecorder())
		tt.handle(c)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
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

func TestContext_Jsonp(t *testing.T) {
	tests := []struct {
		request             *http.Request
		handle              Handle
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			nil,
			func(c *Context) {
				c.Jsonp(nil)
			},
			200,
			jsonpType,
			`/**/ typeof _jsonp_callback_ === 'function' && _jsonp_callback_(null);`,
		},
		{
			nil,
			func(c *Context) {
				c.Jsonp(struct {
					id   uint32
					Name string
				}{1, "x"})
			},
			200,
			jsonpType,
			`/**/ typeof _jsonp_callback_ === 'function' && _jsonp_callback_({"Name":"x"});`,
		},
		{
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			func(c *Context) {
				c.Jsonp(struct {
					ID   uint32 `json:"id"`
					Name string `json:"name"`
				}{1, "x"})
			},
			200,
			jsonpType,
			`/**/ typeof jsonp12345 === 'function' && jsonp12345({"id":1,"name":"x"});`,
		},
		{
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			func(c *Context) {
				c.Jsonp([]string{"foo", "bar"})
			},
			200,
			jsonpType,
			`/**/ typeof jsonp12345 === 'function' && jsonp12345(["foo","bar"]);`,
		},
		{
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			func(c *Context) {
				c.Jsonp([]struct {
					ID   uint32 `json:"id"`
					Name string `json:"name"`
				}{{1, "x"}, {2, "y"}})
			},
			200,
			jsonpType,
			`/**/ typeof jsonp12345 === 'function' && jsonp12345([{"id":1,"name":"x"},{"id":2,"name":"y"}]);`,
		},
	}

	for _, tt := range tests {
		c := NewContext(tt.request, httptest.NewRecorder())
		tt.handle(c)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		if got := w.Code; got != tt.expectedStatus {
			t.Errorf(testErrorFormat, got, tt.expectedStatus)
		}
		if got := c.Get("Content-Type"); got != tt.expectedContentType {
			t.Errorf(testErrorFormat, got, tt.expectedContentType)
		}
		if got := w.Body.String(); got != tt.expectedBody {
			t.Errorf(testErrorFormat, got, tt.expectedBody)
		}
	}
}

func TestContext_Redirect(t *testing.T) {
	deferFn1 := func(w *httptest.ResponseRecorder, code int, cType, loc, body string) {
		if err := recover(); err == nil {
			t.Errorf(testErrorFormat, "nil", "error")
		}
		if got := w.Code; got != code {
			t.Errorf(testErrorFormat, got, code)
		}
		if got := w.Header().Get("Content-Type"); got != cType {
			t.Errorf(testErrorFormat, got, cType)
		}
		if got := w.Header().Get("Location"); got != loc {
			t.Errorf(testErrorFormat, got, loc)
		}
		if got := w.Body.String(); got != body {
			t.Errorf(testErrorFormat, got, body)
		}
	}
	deferFn2 := func(w *httptest.ResponseRecorder, code int, contentType, location, body string) {
		if err := recover(); err != nil {
			t.Errorf(testErrorFormat, err, "nil")
		}
	}
	tests := []struct {
		status         int
		location       string
		method         string
		expectedStatus int
		expectedType   string
		expectedLoc    string
		expectedBody   string
		deferFn        func(w *httptest.ResponseRecorder, code int, cType, loc, body string)
	}{
		{200, "/foo/bar", "GET", 200, "", "", "", deferFn1},
		{309, "/foo/bar", "POST", 200, "", "", "", deferFn1},
		{
			201,
			"/foo/bar",
			"GET",
			201,
			"text/html; charset=utf-8",
			"/foo/bar",
			"<a href=\"/foo/bar\">Created</a>.\n\n",
			deferFn2,
		},
		{301, "/foo/bar", "HEAD", 301, "text/html; charset=utf-8", "/foo/bar", "", deferFn2},
		{
			302,
			"http://google.com",
			"GET",
			302,
			"text/html; charset=utf-8",
			"http://google.com",
			"<a href=\"http://google.com\">Found</a>.\n\n",
			deferFn2,
		},
		{302, "http://google.com", "POST", 302, "", "http://google.com", "", deferFn2},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			c := NewContext(httptest.NewRequest(tt.method, "/", nil), httptest.NewRecorder())
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
			defer tt.deferFn(w, tt.expectedStatus, tt.expectedType, tt.expectedLoc, tt.expectedBody)
			c.Redirect(tt.status, tt.location)
			if got := c.response.Status(); got != tt.expectedStatus {
				t.Errorf(testErrorFormat, got, tt.expectedStatus)
			}
			if got := w.Header().Get("Content-Type"); got != tt.expectedType {
				t.Errorf(testErrorFormat, got, tt.expectedType)
			}
			if got := c.Get("Location"); got != tt.expectedLoc {
				t.Errorf(testErrorFormat, got, tt.expectedLoc)
			}
			if got := w.Body.String(); got != tt.expectedBody {
				t.Errorf(testErrorFormat, got, tt.expectedBody)
			}
		})
	}
}

func TestContext_Render(t *testing.T) {
	str := "foo"
	strRenderer := renderer.String{Data: str}
	tests := []struct {
		name                string
		request             *http.Request
		renderer            renderer.Renderer
		expectedStatus      int
		expectedContentType string
		expectedBody        string
		deferFn             func(w *httptest.ResponseRecorder, code int, contentType, body string)
	}{
		{"string", nil, strRenderer, 200, plainType, str, nil},
		{"string", nil, strRenderer, 100, plainType, "", nil},
		{"string", nil, strRenderer, 204, plainType, "", nil},
		{"string", nil, strRenderer, 304, plainType, "", nil},
		{
			"json",
			nil,
			renderer.JSON{Data: struct {
				Name      string
				PageTotal uint16
			}{"foo", 50}},
			200,
			jsonType,
			`{"Name":"foo","PageTotal":50}`,
			nil,
		},
		{
			"json-custom",
			nil,
			renderer.JSON{Data: struct {
				Name      string `json:"name"`
				PageTotal uint16 `json:"pageTotal"`
			}{"foo", 50}},
			200,
			jsonType,
			`{"name":"foo","pageTotal":50}`,
			nil,
		},
		{
			"json-error",
			nil,
			renderer.JSON{Data: func() {}},
			200,
			jsonType,
			"",
			func(w *httptest.ResponseRecorder, code int, contentType, body string) {
				if got := recover(); got == nil {
					t.Errorf(testErrorFormat, got, "error")
				}
				if got := w.Code; got != code {
					t.Errorf(testErrorFormat, got, code)
				}
				if got := w.Header().Get("Content-Type"); got != contentType {
					t.Errorf(testErrorFormat, got, contentType)
				}
				if got := w.Body.String(); got != body {
					t.Errorf(testErrorFormat, got, body)
				}
			},
		},
		{
			"jsonp",
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			renderer.JSONP{Data: struct {
				ID   uint32 `json:"id"`
				Name string `json:"name"`
			}{1, "x"}},
			200,
			jsonpType,
			`/**/ typeof jsonp12345 === 'function' && jsonp12345({"id":1,"name":"x"});`,
			nil,
		},
		{
			"jsonp-error",
			nil,
			renderer.JSONP{Data: func() {}},
			200,
			jsonpType,
			"",
			func(w *httptest.ResponseRecorder, code int, contentType, body string) {
				if got := recover(); got == nil {
					t.Errorf(testErrorFormat, got, "error")
				}
				if got := w.Code; got != code {
					t.Errorf(testErrorFormat, got, code)
				}
				if got := w.Header().Get("Content-Type"); got != contentType {
					t.Errorf(testErrorFormat, got, contentType)
				}
				if got := w.Body.String(); got != body {
					t.Errorf(testErrorFormat, got, body)
				}
			},
		},
		{
			"redirect",
			httptest.NewRequest("GET", "http://example.com", nil),
			renderer.Redirect{Code: 301, Location: "http://example.com"},
			301,
			"text/html; charset=utf-8",
			"<a href=\"http://example.com\">Moved Permanently</a>.\n\n",
			nil,
		},
		{
			"redirect-error",
			httptest.NewRequest("GET", "http://example.com", nil),
			renderer.Redirect{Code: 200, Location: "http://example.com"},
			200,
			"",
			"",
			func(w *httptest.ResponseRecorder, code int, contentType, body string) {
				if got := recover(); got == nil {
					t.Errorf(testErrorFormat, got, "error")
				}
				if got := w.Code; got != code {
					t.Errorf(testErrorFormat, got, code)
				}
				if got := w.Header().Get("Content-Type"); got != contentType {
					t.Errorf(testErrorFormat, got, contentType)
				}
				if got := w.Header().Get("Location"); got != "" {
					t.Errorf(testErrorFormat, got, "")
				}
				if got := w.Body.String(); got != body {
					t.Errorf(testErrorFormat, got, body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContext(tt.request, httptest.NewRecorder())
			if tt.expectedStatus != 200 {
				c.Status(tt.expectedStatus)
			}
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
			if tt.deferFn != nil {
				defer tt.deferFn(w, tt.expectedStatus, tt.expectedContentType, tt.expectedBody)
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
				if rr, ok := tt.renderer.(renderer.Redirect); ok {
					if got := w.Header().Get("Location"); got != rr.Location {
						t.Errorf(testErrorFormat, got, rr.Location)
					}
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

func stringSliceEquals(a, b []string) bool {
	sort.StringSlice(a).Sort()
	sort.StringSlice(b).Sort()
	return strings.Join(a, ",") == strings.Join(b, ",")
}
