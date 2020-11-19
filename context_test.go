// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/soongo/soon/binding"
	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	timeFormat   = http.TimeFormat
	dotRegexp    = regexp.MustCompile("\\s*,\\s*")
	plainType    = "text/plain; charset=UTF-8"
	htmlType     = "text/html; charset=UTF-8"
	jsonType     = "application/json; charset=UTF-8"
	jsonpType    = "text/javascript; charset=UTF-8"
	emptyRequest = httptest.NewRequest("GET", "/", nil)
)

type ErrTooLargeReader struct{}

func (r *ErrTooLargeReader) Read(p []byte) (n int, err error) {
	panic(bytes.ErrTooLarge)
}

type jsonChild struct {
	Name   string `json:"name" validate:"required,min=3"`
	Age    int    `json:"age" validate:"gte=0,max=150"`
	Gender string `json:"gender" validate:"oneof=male female"`
}

type jsonRoot struct {
	Foo   string    `json:"foo" validate:"required"`
	Child jsonChild `json:"child"`
}

var jsonBindTests = []struct {
	s        jsonRoot
	json     string
	errs     []string
	expected jsonRoot
}{
	{
		json: `{"foo": "FOO", "child": {"name": "matt", "age": 39, "gender": "male"}}`,
		expected: jsonRoot{
			Foo:   "FOO",
			Child: jsonChild{Name: "matt", Age: 39, Gender: "male"},
		},
	},
	{
		json: `{"foo": "FOO", "child": {"name": "hi", "age": -1, "gender": "x"}}`,
		expected: jsonRoot{
			Foo:   "FOO",
			Child: jsonChild{Name: "hi", Age: -1, Gender: "x"},
		},
		errs: []string{
			"Key: 'jsonRoot.Child.Name' Error:Field validation for 'Name' failed on the 'min' tag",
			"Key: 'jsonRoot.Child.Age' Error:Field validation for 'Age' failed on the 'gte' tag",
			"Key: 'jsonRoot.Child.Gender' Error:Field validation for 'Gender' failed on the 'oneof' tag",
		},
	},
	{
		json: `{"foo": ""}`,
		expected: jsonRoot{
			Foo: "",
		},
		errs: []string{
			"Key: 'jsonRoot.Foo' Error:Field validation for 'Foo' failed on the 'required' tag",
			"Key: 'jsonRoot.Child.Name' Error:Field validation for 'Name' failed on the 'required' tag",
			"Key: 'jsonRoot.Child.Gender' Error:Field validation for 'Gender' failed on the 'oneof' tag",
		},
	},
}

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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		tt.handle(c)
		assert.Equal(t, tt.expected, c.HeadersSent())
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.SetLocal(tt.k, tt.v)
		v, ok := c.GetLocal(tt.k)
		assert.True(t, ok)
		assert.Equal(t, tt.v, v)
		v, ok = c.GetLocal("not_exists_key")
		assert.False(t, ok)
		assert.Nil(t, v)
	}

	t.Run("SetLocals", func(t *testing.T) {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		locals := map[string]interface{}{"name": "foo", "age": 10}
		c.SetLocals(locals)
		assert.Equal(t, locals, c.Locals)
		assert.Equal(t, locals["name"], c.MustGetLocal("name"))
		assert.Equal(t, locals["age"], c.MustGetLocal("age"))
		locals["name"] = "bar"
		assert.NotEqual(t, locals["name"], c.MustGetLocal("name"))
	})

	t.Run("MustGetLocal", func(t *testing.T) {
		defer func() {
			assert.NotNil(t, recover())
		}()
		c := NewContext(emptyRequest, httptest.NewRecorder())
		locals := map[string]interface{}{"name": "foo", "age": 10}
		c.SetLocals(locals)
		c.MustGetLocal("not_exists_key")
	})

	t.Run("ResetLocals", func(t *testing.T) {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		locals := map[string]interface{}{"name": "foo", "age": 10}
		c.SetLocals(locals)
		assert.Equal(t, locals, c.Locals)
		assert.Equal(t, locals["name"], c.MustGetLocal("name"))
		assert.Equal(t, locals["age"], c.MustGetLocal("age"))
		c.ResetLocals(nil)
		assert.Equal(t, map[string]interface{}{}, c.Locals)
		assert.Equal(t, 0, len(c.Locals))

		c.ResetLocals(map[string]interface{}{"name": "bar", "gender": "male"})
		assert.Equal(t, 2, len(c.Locals))
		assert.Equal(t, "bar", c.MustGetLocal("name"))
		assert.Equal(t, "male", c.MustGetLocal("gender"))
		age, ok := c.GetLocal("age")
		assert.Nil(t, age)
		assert.False(t, ok)
	})
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
		assert.Equal(t, tt.params, c.Request.Params)
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
		assert.Equal(t, tt.expected, c.Request.Query)
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Append(tt.k, tt.v)
		assert.Equal(t, tt.expected, c.Writer.Header()[tt.k])
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Writer.Header().Set(k, "text/*")
		if tt.k == "" {
			c.Set(tt.v)
		} else {
			c.Set(tt.k, tt.v)
		}
		assert.Equal(t, tt.expected, c.Writer.Header())
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

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		assert.Equal("", c.Get(tt.k))
		c.Set(tt.k, tt.v)
		assert.Equal(tt.expected, c.Get(tt.k))
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Set(key, tt.vary)
		c.Vary(tt.fields...)
		assert.Equal(t, tt.expected, c.Get(key))
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

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Status(tt.code)
		assert.Equal(tt.code, c.Writer.Status())

		c.Writer.Flush()
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		assert.Equal(tt.code, w.Code)
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

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.SendStatus(tt.code)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		assert.Equal(tt.code, w.Code)
		assert.Equal(http.StatusText(w.Code), w.Body.String())
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Type(tt.t)
		assert.Equal(t, tt.expected, c.Get("Content-Type"))
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Set("Link", tt.origin)
		c.Links(tt.links)
		got := c.Get("Link")
		assert.ElementsMatch(t, strings.Split(tt.expected, ", "), strings.Split(got, ", "))
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		if tt.expected == "" {
			tt.expected = util.EncodeURI(strings.Trim(tt.location, " "))
		}
		if tt.referrer != "" {
			c.Set("Referrer", tt.referrer)
		}
		c.Location(tt.location)
		assert.Equal(t, tt.expected, c.Get("location"))
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		if tt.none {
			c.Attachment()
		} else {
			c.Attachment(tt.s)
		}
		assert.Equal(t, tt.expected, c.Get("Content-Disposition"))
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
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Cookie(tt.cookie)
		assert.Equal(t, tt.expected, c.Get("Set-Cookie"))
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

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Cookie(tt.cookie)
		assert.Equal(tt.expected[0], c.Get("Set-Cookie"))
		c.ClearCookie(tt.cookie)
		assert.Equal(tt.expected, c.Writer.Header()["Set-Cookie"])
	}
}

func TestContext_SendFile(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	deferFn := func(w *httptest.ResponseRecorder, code int, e error) {
		err := recover()
		assert.NotNil(t, err)
		if e != nil {
			assert.Equal(t, e, err)
			if httpErr, ok := e.(HttpError); ok {
				assert.Equal(t, code, httpErr.Status())
			} else {
				assert.Equal(t, code, w.Code)
			}
		} else {
			assert.Equal(t, code, w.Code)
		}
		assert.Empty(t, w.Header()["Content-Type"])
	}

	maxAge := time.Hour
	tests := []struct {
		name                string
		filePath            string
		options             renderer.FileOptions
		expectedStatus      int
		expectedContentType string
		expectedError       error
		deferFn             func(w *httptest.ResponseRecorder, code int, e error)
	}{
		{
			"normal-1",
			path.Join(pwd, "README.md"),
			renderer.FileOptions{},
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{
			"normal-2",
			path.Join(pwd, "README.md"),
			renderer.FileOptions{
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
		{"empty-filepath", "", renderer.FileOptions{}, 200, "", nil, deferFn},
		{
			"with-root-path",
			"README.md",
			renderer.FileOptions{Root: pwd, LastModifiedDisabled: true},
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{"not-root-filepath", "README.md", renderer.FileOptions{}, 200, "", nil, deferFn},
		{
			"directory",
			pwd,
			renderer.FileOptions{Index: renderer.IndexDisabled},
			400,
			"",
			renderer.ErrIsDir,
			deferFn,
		},
		{
			"custom directory index",
			pwd,
			renderer.FileOptions{Index: "README.md"},
			200,
			"text/markdown; charset=UTF-8",
			nil,
			nil,
		},
		{
			"hidden-default",
			path.Join(pwd, ".travis.yml"),
			renderer.FileOptions{},
			404,
			"",
			renderer.ErrNotFound,
			deferFn,
		},
		{
			"hidden-allow",
			path.Join(pwd, ".travis.yml"),
			renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyAllow},
			200,
			"text/yaml; charset=UTF-8",
			nil,
			nil,
		},
		{
			"hidden-deny",
			path.Join(pwd, ".travis.yml"),
			renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyDeny},
			403,
			"",
			renderer.ErrForbidden,
			deferFn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			c := NewContext(emptyRequest, httptest.NewRecorder())
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
			if tt.deferFn != nil {
				func() {
					defer tt.deferFn(w, tt.expectedStatus, tt.expectedError)
					c.SendFile(tt.filePath, tt.options)
				}()
			} else {
				c.SendFile(tt.filePath, tt.options)
				if tt.options.Index != "" {
					tt.filePath = path.Join(tt.filePath, tt.options.Index)
				}
				fileInfo, fileContent := getFileContent(tt.filePath, nil)
				lastModified := fileInfo.ModTime().UTC().Format(timeFormat)
				assert.Equal(tt.expectedStatus, w.Code)
				assert.Equal(fileContent, w.Body.String())
				assert.Equal(tt.expectedContentType, c.Get("Content-Type"))
				if tt.options.MaxAge != nil {
					cc := fmt.Sprintf("max-age=%.0f", maxAge.Seconds())
					assert.Equal(cc, c.Get("Cache-Control"))
				}
				if tt.options.Header != nil {
					for k, v := range tt.options.Header {
						assert.Equal(v, c.Get(k))
					}
				}
				expectedLastModified := lastModified
				if tt.options.LastModifiedDisabled {
					expectedLastModified = ""
				}
				assert.Equal(expectedLastModified, c.Get("Last-Modified"))
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
		options        renderer.FileOptions
		expectedStatus int
	}{
		{path.Join(pwd, "README.md"), renderer.FileOptions{}, 200},
		{path.Join(pwd, "README.md"), renderer.FileOptions{Name: "custom-name"}, 200},
	}

	for _, tt := range tests {
		assert := assert.New(t)
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.Download(tt.filePath, tt.options)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		assert.Equal(tt.expectedStatus, w.Code)
		fileInfo, fileContent := getFileContent(tt.filePath, nil)
		assert.Equal(fileContent, w.Body.String())
		name := fileInfo.Name()
		if tt.options.Name != "" {
			name = tt.options.Name
		}
		contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", name)
		assert.Equal(contentDisposition, c.Get("Content-Disposition"))
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
			"foo",
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
		{
			"with-end-2",
			func(c *Context) {
				c.End()
				c.Send("foo")
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContext(emptyRequest, httptest.NewRecorder())
			tt.handle(c)
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
			assert.Equal(t, tt.expected, w.Body.String())
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

	assert := assert.New(t)
	for i, tt := range tests {
		path := "/" + strconv.Itoa(i)
		h := http.Header{"Accept": dotRegexp.Split(tt.accept, -1)}
		statusCode, _, body, err := request(http.MethodGet, server.URL+path, h)
		body = strings.Trim(body, "\n")
		assert.Nil(err)
		assert.Equal(tt.expectedStatus, statusCode)
		assert.Equal(tt.expectedBody, body)
	}
}

func TestContext_BindJSON(t *testing.T) {
	for _, tt := range jsonBindTests {
		req := httptest.NewRequest("GET", "/", strings.NewReader(tt.json))
		c := NewContext(req, httptest.NewRecorder())
		err := c.BindJSON(&tt.s)
		if tt.errs == nil {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.s)
			var s jsonRoot
			err = c.BindJSON(&s)
			require.EqualError(t, err, "EOF")
		} else {
			require.Error(t, err)
			require.EqualError(t, err, strings.Join(tt.errs, "\n"))
			assert.Equal(t, tt.expected, tt.s)
		}
	}
}

func TestContext_BindQuery(t *testing.T) {
	req := httptest.NewRequest("POST", "/?foo=bar&bar=foo", bytes.NewBufferString("foo=unused"))
	w := httptest.NewRecorder()
	c := NewContext(req, w)

	var obj struct {
		Foo string `form:"foo"`
		Bar string `form:"bar"`
	}
	assert.NoError(t, c.BindQuery(&obj))
	assert.Equal(t, "foo", obj.Bar)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, 0, w.Body.Len())
}

func TestContext_BindHeader(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(req, w)

	c.Request.Header.Add("rate", "8000")
	c.Request.Header.Add("domain", "music")
	c.Request.Header.Add("limit", "1000")

	var testHeader struct {
		Rate   int    `header:"Rate"`
		Domain string `header:"Domain"`
		Limit  int    `header:"limit"`
	}

	assert.NoError(t, c.BindHeader(&testHeader))
	assert.Equal(t, 8000, testHeader.Rate)
	assert.Equal(t, "music", testHeader.Domain)
	assert.Equal(t, 1000, testHeader.Limit)
	assert.Equal(t, 0, w.Body.Len())
}

func TestContext_BindUri(t *testing.T) {
	router := NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	type Person struct {
		Name string `uri:"name" validate:"required"`
		ID   string `uri:"id" validate:"required"`
	}

	router.GET("/rest/:name/:id/(.*)", func(c *Context) {
		var person Person
		assert.NoError(t, c.BindUri(&person))
		assert.Equal(t, "foo", person.Name)
		assert.Equal(t, "001", person.ID)
		c.String(body200)
	})

	statusCode, _, body, err := request(http.MethodGet, server.URL+"/rest/foo/001/", nil)
	require.NoError(t, err)
	assert.Equal(t, 200, statusCode)
	assert.Equal(t, body200, body)
}

func TestContext_BindWith(t *testing.T) {
	for _, tt := range jsonBindTests {
		req := httptest.NewRequest("GET", "/", strings.NewReader(tt.json))
		c := NewContext(req, httptest.NewRecorder())
		err := c.BindWith(&tt.s, binding.JSON)
		if tt.errs == nil {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.s)
			var s jsonRoot
			err = c.BindWith(&s, binding.JSON)
			require.EqualError(t, err, "EOF")
		} else {
			require.Error(t, err)
			require.EqualError(t, err, strings.Join(tt.errs, "\n"))
			assert.Equal(t, tt.expected, tt.s)
		}
	}
}

func TestContext_MustBindJSON(t *testing.T) {
	for _, tt := range jsonBindTests {
		t.Run("", func(t *testing.T) {
			if tt.errs != nil {
				defer func() {
					err := recover()
					require.NotNil(t, err)
					httpErr := err.(HttpError)
					assert.Equal(t, http.StatusBadRequest, httpErr.Status())
					assert.Equal(t, strings.Join(tt.errs, "\n"), httpErr.Error())
				}()
			}
			req := httptest.NewRequest("GET", "/", strings.NewReader(tt.json))
			c := NewContext(req, httptest.NewRecorder())
			c.MustBindJSON(&tt.s)
		})
	}
}

func TestContext_MustBindQuery(t *testing.T) {
	req := httptest.NewRequest("POST", "/?foo=bar&age=-1", bytes.NewBufferString("foo=unused"))
	w := httptest.NewRecorder()
	c := NewContext(req, w)

	defer func() {
		err := recover()
		require.NotNil(t, err)
		httpErr := err.(HttpError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Status())

		errText := strings.Join([]string{
			"Key: 'Bar' Error:Field validation for 'Bar' failed on the 'required' tag",
			"Key: 'Age' Error:Field validation for 'Age' failed on the 'gte' tag",
		}, "\n")
		assert.Equal(t, errText, httpErr.Error())
	}()

	var obj struct {
		Foo string `form:"foo" validate:"required"`
		Bar string `form:"bar" validate:"required"`
		Age int    `form:"age" validate:"required,gte=0"`
	}
	c.MustBindQuery(&obj)
}

func TestContext_MustBindHeader(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	c := NewContext(req, w)

	c.Request.Header.Add("rate", "-1")
	c.Request.Header.Add("limit", "2")

	defer func() {
		err := recover()
		require.NotNil(t, err)
		httpErr := err.(HttpError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Status())

		errText := strings.Join([]string{
			"Key: 'Rate' Error:Field validation for 'Rate' failed on the 'gte' tag",
			"Key: 'Domain' Error:Field validation for 'Domain' failed on the 'required' tag",
			"Key: 'Limit' Error:Field validation for 'Limit' failed on the 'min' tag",
		}, "\n")
		assert.Equal(t, errText, httpErr.Error())
	}()

	var testHeader struct {
		Rate   int    `header:"Rate" validate:"gte=0,max=10"`
		Domain string `header:"Domain" validate:"required"`
		Limit  int    `header:"limit" validate:"required,min=3"`
	}
	c.MustBindHeader(&testHeader)
}

func TestContext_MustBindUri(t *testing.T) {
	router := NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	type Person struct {
		Name string `uri:"name" validate:"required"`
		ID   string `uri:"id" validate:"required,min=6"`
	}

	router.GET("/rest/:name/:id", func(c *Context) {
		var person Person
		c.MustBindUri(&person)
		c.String(body200)
	})

	router.Use(func(err interface{}, c *Context) {
		require.NotNil(t, err)
		httpErr := err.(HttpError)
		assert.Equal(t, http.StatusBadRequest, httpErr.Status())
		c.Next(err)
	})

	statusCode, _, body, err := request(http.MethodGet, server.URL+"/rest/foo/100", nil)
	require.NoError(t, err)
	assert.Equal(t, 400, statusCode)
	assert.Equal(t, "Key: 'Person.ID' Error:Field validation for 'ID' failed on the 'min' tag", body)
}

func TestContext_MustBindWith(t *testing.T) {
	for _, tt := range jsonBindTests {
		t.Run("", func(t *testing.T) {
			if tt.errs != nil {
				defer func() {
					err := recover()
					require.NotNil(t, err)
					httpErr := err.(HttpError)
					assert.Equal(t, http.StatusBadRequest, httpErr.Status())
					assert.Equal(t, strings.Join(tt.errs, "\n"), httpErr.Error())
				}()
			}
			req := httptest.NewRequest("GET", "/", strings.NewReader(tt.json))
			c := NewContext(req, httptest.NewRecorder())
			c.MustBindWith(&tt.s, binding.JSON)
		})
	}
}

func TestContext_BindBodyWith(t *testing.T) {
	for _, tt := range jsonBindTests {
		req := httptest.NewRequest("GET", "/", strings.NewReader(tt.json))
		c := NewContext(req, httptest.NewRecorder())
		err := c.BindBodyWith(&tt.s, binding.JSON)
		if tt.errs == nil {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.s)
			var s jsonRoot
			err = c.BindWith(&s, binding.JSON)
			require.EqualError(t, err, "EOF")
			err = c.BindBodyWith(&s, binding.JSON)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, s)
		} else {
			require.Error(t, err)
			require.EqualError(t, err, strings.Join(tt.errs, "\n"))
			assert.Equal(t, tt.expected, tt.s)
		}
	}

	req := httptest.NewRequest("GET", "/", &ErrTooLargeReader{})
	c := NewContext(req, httptest.NewRecorder())
	err := c.BindBodyWith(&jsonBindTests[0].s, binding.JSON)
	require.Equal(t, bytes.ErrTooLarge, err)
}

func TestContext_String(t *testing.T) {
	tests := []struct {
		s                   string
		expectedStatus      int
		expectedContentType string
	}{
		{"foo", 200, plainType},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		c.String(tt.s)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		assert.Equal(tt.expectedStatus, w.Code)
		assert.Equal(tt.expectedContentType, c.Get("Content-Type"))
		assert.Equal(tt.s, w.Body.String())
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

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(emptyRequest, httptest.NewRecorder())
		tt.handle(c)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		assert.Equal(tt.expectedStatus, w.Code)
		assert.Equal(tt.expectedContentType, c.Get("Content-Type"))
		assert.Equal(tt.expectedBody+"\n", w.Body.String())
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
			emptyRequest,
			func(c *Context) {
				c.Jsonp(nil)
			},
			200,
			jsonpType,
			`/**/ typeof _jsonp_callback_ === 'function' && _jsonp_callback_(null);`,
		},
		{
			emptyRequest,
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

	assert := assert.New(t)
	for _, tt := range tests {
		c := NewContext(tt.request, httptest.NewRecorder())
		tt.handle(c)
		w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
		assert.Equal(tt.expectedStatus, w.Code)
		assert.Equal(tt.expectedContentType, c.Get("Content-Type"))
		assert.Equal(tt.expectedBody, w.Body.String())
	}
}

func TestContext_Redirect(t *testing.T) {
	deferFn1 := func(w *httptest.ResponseRecorder, code int, cType, loc, body string) {
		err := recover()
		assert.NotNil(t, err)
		assert.Equal(t, code, w.Code)
		assert.Equal(t, cType, w.Header().Get("Content-Type"))
		assert.Equal(t, loc, w.Header().Get("Location"))
		assert.Equal(t, body, w.Body.String())
	}
	deferFn2 := func(w *httptest.ResponseRecorder, code int, contentType, location, body string) {
		err := recover()
		assert.Nil(t, err)
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
		{200, "/foo/bar", "GET", 200, "text/html; charset=utf-8", "/foo/bar", "", deferFn1},
		{309, "/foo/bar", "POST", 200, "", "/foo/bar", "", deferFn1},
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
		{
			301,
			"/foo/bar",
			"GET",
			301,
			"text/html; charset=utf-8",
			"/foo/bar",
			"<a href=\"/foo/bar\">Moved Permanently</a>.\n\n",
			deferFn2,
		},
		{
			301,
			"/foo/bar",
			"HEAD",
			301,
			"text/html; charset=utf-8",
			"/foo/bar",
			"",
			deferFn2,
		},
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
			assert := assert.New(t)
			c := NewContext(httptest.NewRequest(tt.method, "/", nil), httptest.NewRecorder())
			w := c.response.ResponseWriter.(*httptest.ResponseRecorder)
			defer tt.deferFn(w, tt.expectedStatus, tt.expectedType, tt.expectedLoc, tt.expectedBody)
			c.Redirect(tt.status, tt.location)
			assert.Equal(tt.expectedStatus, c.response.Status())
			assert.Equal(tt.expectedType, w.Header().Get("Content-Type"))
			assert.Equal(tt.expectedLoc, c.Get("Location"))
			assert.Equal(tt.expectedBody, w.Body.String())
		})
	}
}

func TestContext_Render(t *testing.T) {
	str := "foo"
	strRenderer := &renderer.String{Data: str}
	tests := []struct {
		name                string
		request             *http.Request
		renderer            renderer.Renderer
		expectedStatus      int
		expectedContentType string
		expectedBody        string
		deferFn             func(w *httptest.ResponseRecorder, code int, contentType, body string)
	}{
		{"string", emptyRequest, strRenderer, 200, plainType, str, nil},
		{"string", emptyRequest, strRenderer, 100, plainType, "", nil},
		{"string", emptyRequest, strRenderer, 204, "", "", nil},
		{"string", emptyRequest, strRenderer, 304, "", "", nil},
		{
			"json",
			emptyRequest,
			&renderer.JSON{Data: struct {
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
			emptyRequest,
			&renderer.JSON{Data: struct {
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
			emptyRequest,
			&renderer.JSON{Data: func() {}},
			200,
			jsonType,
			"",
			func(w *httptest.ResponseRecorder, code int, contentType, body string) {
				err := recover()
				assert.NotNil(t, err)
				assert.Equal(t, code, w.Code)
				assert.Equal(t, contentType, w.Header().Get("Content-Type"))
				assert.Equal(t, body, w.Body.String())
			},
		},
		{
			"jsonp",
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			&renderer.JSONP{Data: struct {
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
			emptyRequest,
			&renderer.JSONP{Data: func() {}},
			200,
			jsonpType,
			"",
			func(w *httptest.ResponseRecorder, code int, contentType, body string) {
				err := recover()
				assert.NotNil(t, err)
				assert.Equal(t, code, w.Code)
				assert.Equal(t, contentType, w.Header().Get("Content-Type"))
				assert.Equal(t, body, w.Body.String())
			},
		},
		{
			"redirect",
			httptest.NewRequest("GET", "http://example.com", nil),
			&renderer.Redirect{Code: 301, Location: "http://example.com"},
			301,
			"text/html; charset=utf-8",
			"<a href=\"http://example.com\">Moved Permanently</a>.\n\n",
			nil,
		},
		{
			"redirect-error",
			httptest.NewRequest("GET", "http://example.com", nil),
			&renderer.Redirect{Code: 200, Location: "http://example.com"},
			200,
			"text/html; charset=utf-8",
			"",
			func(w *httptest.ResponseRecorder, code int, contentType, body string) {
				err := recover()
				assert.NotNil(t, err)
				assert.Equal(t, code, w.Code)
				assert.Equal(t, contentType, w.Header().Get("Content-Type"))
				assert.Equal(t, "http://example.com", w.Header().Get("Location"))
				assert.Equal(t, body, w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
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
				assert.Equal(tt.expectedStatus, w.Code)
				assert.Equal(tt.expectedContentType, c.Get("Content-Type"))
				expectedBody := tt.expectedBody
				if _, ok := tt.renderer.(*renderer.JSON); ok {
					expectedBody += "\n"
				}
				assert.Equal(expectedBody, w.Body.String())
				if rr, ok := tt.renderer.(*renderer.Redirect); ok {
					assert.Equal(rr.Location, w.Header().Get("Location"))
				}
			}
		})
	}
}

func getFileContent(p string, r *util.Range) (os.FileInfo, string) {
	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	fileInfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	start, size := int64(0), fileInfo.Size()
	if r != nil {
		start = r.Start
		size = r.End - r.Start + 1
	}

	bts := make([]byte, size)
	_, err = f.ReadAt(bts, start)
	if err != nil {
		panic(err)
	}

	return fileInfo, string(bts)
}
