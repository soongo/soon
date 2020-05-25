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
	"os"
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
	dotRegexp       = regexp.MustCompile("\\s*,\\s*")
)

func TestContext_HeadersSent(t *testing.T) {
	c := &Context{ResponseWriter: httptest.NewRecorder()}
	if c.HeadersSent != false {
		t.Errorf(testErrorFormat, c.HeadersSent, false)
	}

	c.Send("foo")
	if c.HeadersSent != true {
		t.Errorf(testErrorFormat, c.HeadersSent, true)
	}
}

func TestContext_Locals(t *testing.T) {
	key, expected := "foo", "bar"
	c := &Context{ResponseWriter: httptest.NewRecorder()}
	c.init()
	c.Locals.Set(key, expected)
	result := c.Locals.Get(key)
	if result != expected {
		t.Errorf(testErrorFormat, result, expected)
	}
}

func TestContext_Append(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		k, expected := "Content-Type", "application/json"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Append(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("custom", func(t *testing.T) {
		k, expected := "x-custom", "custom"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Append(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		k, expected := "Content-Type", "application/json"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Append(k, []string{expected, expected})
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}

		m := map[string][]string{k: {expected, expected}}
		r := map[string][]string(res.Header())
		if !headerEquals(m, r) {
			t.Errorf(testErrorFormat, r, m)
		}
	})
}

func TestContext_Set(t *testing.T) {
	k, expected := "Content-Type", "application/json; charset=UTF-8"

	t.Run("normal", func(t *testing.T) {
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Set(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("replace", func(t *testing.T) {
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Header().Set(k, "text/plain")
		res.Set(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]string{
			"Content-Type": "application/json; charset=UTF-8",
			"X-Custom":     "custom",
		}
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Set(m)
		result, expected := res.Header(), mapToHeader(m)
		if !headerEquals(result, expected) {
			t.Errorf(testErrorFormat, result, expected)
		}
	})
}

func TestContext_Get(t *testing.T) {
	k, expected := "Content-Type", "application/json"
	res := &Context{ResponseWriter: httptest.NewRecorder()}
	result := res.Get(k)
	if result != "" {
		t.Errorf("got `%s`, expect `%v`", result, expected)
	}

	res.Header().Set(k, expected)
	result = res.Get(k)
	if result != expected {
		t.Errorf("got `%s`, expect `%v`", result, expected)
	}
}

func TestContext_Status(t *testing.T) {
	expected := 404
	res := &Context{ResponseWriter: httptest.NewRecorder()}
	res.Status(expected)
	recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
	result := recorder.Code
	if result != expected {
		t.Errorf("got `%d`, expect `%d`", result, expected)
	}
}

func TestContext_SendStatus(t *testing.T) {
	expectedCode := 302
	expectedBody := http.StatusText(expectedCode)
	res := &Context{ResponseWriter: httptest.NewRecorder()}
	res.SendStatus(expectedCode)
	recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
	resultCode, resultBody := recorder.Code, recorder.Body.String()
	if resultCode != expectedCode {
		t.Errorf("got `%d`, expect `%d`", resultCode, expectedCode)
	}
	if resultBody != expectedBody {
		t.Errorf(testErrorFormat, resultBody, expectedBody)
	}
}

func TestContext_Type(t *testing.T) {
	k := "Content-Type"
	t.Run("normal", func(t *testing.T) {
		expected := "text/html; charset=UTF-8"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Type("html")
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}

		res.Type("index.html")
		result = res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("slash", func(t *testing.T) {
		expected := "image/png"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Type(expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}

		expected = "/"
		res.Type(expected)
		result = res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("empty", func(t *testing.T) {
		expected := "application/octet-stream"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Type("")
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})
}

func TestContext_Attachment(t *testing.T) {
	k := "Content-Disposition"
	t.Run("name", func(t *testing.T) {
		name := "foo.png"
		expected := fmt.Sprintf("attachment; filename=\"%s\"", name)
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Attachment(name)
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("empty", func(t *testing.T) {
		expected := "attachment"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Attachment()
		result := res.Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})
}

func TestContext_Cookie(t *testing.T) {
	k, expected := "Set-Cookie", "foo=bar; Path=/; HttpOnly"
	res := &Context{ResponseWriter: httptest.NewRecorder()}
	c := &http.Cookie{Name: "foo", Value: "bar", Path: "/", HttpOnly: true}
	res.Cookie(c)
	result := res.Get(k)
	if result != expected {
		t.Errorf(testErrorFormat, result, expected)
	}
}

func TestContext_ClearCookie(t *testing.T) {
	k, expected := "Set-Cookie", "foo=bar; Path=/; HttpOnly"
	res := &Context{ResponseWriter: httptest.NewRecorder()}
	c := &http.Cookie{Name: "foo", Value: "bar", Path: "/", HttpOnly: true}
	res.Cookie(c)
	result := res.Get(k)
	if result != expected {
		t.Errorf(testErrorFormat, result, expected)
	}

	res.ClearCookie(c)
	c1 := &http.Cookie{
		Name:    c.Name,
		Value:   "",
		Path:    c.Path,
		Expires: time.Unix(0, 0),
	}
	expects := []string{c.String(), c1.String()}

	results := (map[string][]string(res.Header()))[k]
	if !reflect.DeepEqual(results, expects) {
		t.Errorf(testErrorFormat, results, expects)
	}
}

func TestContext_SendFile(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	t.Run("normal", func(t *testing.T) {
		filePath := pwd + "/README.md"
		fileInfo, fileContent := getFileContent(filePath)
		lastModified := fileInfo.ModTime().UTC().Format(http.TimeFormat)
		k, expectedHeader := "Content-Type", "text/markdown; charset=UTF-8"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.SendFile(filePath, nil)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		body := recorder.Body.String()
		if body != fileContent {
			t.Errorf(testErrorFormat, body, fileContent)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf(testErrorFormat, resultHeader, expectedHeader)
		}
		if res.Get("Last-Modified") != lastModified {
			t.Errorf(testErrorFormat, res.Get("Last-Modified"), lastModified)
		}
	})

	t.Run("hidden", func(t *testing.T) {
		filePath := pwd + "/.travis.yml"
		t.Run("default", func(t *testing.T) {
			res := &Context{ResponseWriter: httptest.NewRecorder()}
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			defer func() {
				if err := recover(); err != nil {
					if err != renderer.ErrNotFound {
						t.Error("should got renderer.ErrNotFound error")
					}
					if recorder.Code != http.StatusInternalServerError {
						t.Errorf("got `%d`, expect `%d`", recorder.Code, http.StatusInternalServerError)
					}
				}
			}()
			res.SendFile(filePath, nil)
		})
		t.Run("allow", func(t *testing.T) {
			res := &Context{ResponseWriter: httptest.NewRecorder()}
			options := &renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyAllow}
			res.SendFile(filePath, options)
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			body := recorder.Body.String()
			_, fileContent := getFileContent(filePath)
			if recorder.Code != http.StatusOK {
				t.Errorf("got `%d`, expect `%d`", recorder.Code, http.StatusOK)
			}
			if body != fileContent {
				t.Errorf(testErrorFormat, body, fileContent)
			}
		})
		t.Run("deny", func(t *testing.T) {
			res := &Context{ResponseWriter: httptest.NewRecorder()}
			options := &renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyDeny}
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			defer func() {
				if err := recover(); err != nil {
					if err != renderer.ErrForbidden {
						t.Error("should got renderer.ErrNotFound error")
					}
					if recorder.Code != http.StatusInternalServerError {
						t.Errorf("got `%d`, expect `%d`", recorder.Code, http.StatusInternalServerError)
					}
				}
			}()
			res.SendFile(filePath, options)
		})
	})
}

func TestContext_Download(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filePath := pwd + "/README.md"

	t.Run("normal", func(t *testing.T) {
		fileInfo, fileContent := getFileContent(filePath)
		lastModified := fileInfo.ModTime().UTC().Format(http.TimeFormat)
		k, expectedHeader := "Content-Type", "text/markdown; charset=UTF-8"
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Download(filePath, nil)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		body := recorder.Body.String()
		if body != fileContent {
			t.Errorf(testErrorFormat, body, fileContent)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf(testErrorFormat, resultHeader, expectedHeader)
		}
		if res.Get("Last-Modified") != lastModified {
			t.Errorf(testErrorFormat, res.Get("Last-Modified"), lastModified)
		}
		contentDisposition := `attachment; filename="README.md"`
		if res.Get("Content-Disposition") != contentDisposition {
			t.Errorf(testErrorFormat, res.Get("Content-Disposition"), contentDisposition)
		}
	})

	t.Run("custom name", func(t *testing.T) {
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		options := &renderer.FileOptions{Name: "custom-name"}
		res.Download(filePath, options)
		contentDisposition := `attachment; filename="` + options.Name + `"`
		if res.Get("Content-Disposition") != contentDisposition {
			t.Errorf(testErrorFormat, res.Get("Content-Disposition"), contentDisposition)
		}
	})
}

func TestContext_End(t *testing.T) {
	t.Run("without end", func(t *testing.T) {
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Send("foo")
		res.Send("bar")
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		expected, body := "foobar", recorder.Body.String()
		if body != expected {
			t.Errorf(testErrorFormat, body, expected)
		}
	})

	t.Run("without end", func(t *testing.T) {
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Send("foo")
		res.End()
		res.Send("bar")
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		expected, body := "foo", recorder.Body.String()
		if body != expected {
			t.Errorf(testErrorFormat, body, expected)
		}
	})
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
	k, expected, expectedHeader := "Content-Type", "foo", "text/plain; charset=utf-8"
	res := &Context{ResponseWriter: httptest.NewRecorder()}
	res.String(expected)
	recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
	result := recorder.Body.String()
	if result != expected {
		t.Errorf(testErrorFormat, result, expected)
	}
	resultHeader := res.Get(k)
	if resultHeader != expectedHeader {
		t.Errorf(testErrorFormat, resultHeader, expectedHeader)
	}
}

func TestContext_Json(t *testing.T) {
	k, expectedHeader := "Content-Type", "application/json; charset=utf-8"

	t.Run("normal", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(`{"Name":"foo","PageTotal":500}`))
		buf.WriteByte('\n')
		expected := buf.String()
		book := struct {
			Name      string
			PageTotal uint16
		}{"foo", 500}
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Json(book)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		result := recorder.Body.String()
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf(testErrorFormat, resultHeader, expectedHeader)
		}
	})

	t.Run("custom", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(`{"name":"foo","pageTotal":500}`))
		buf.WriteByte('\n')
		expected := buf.String()
		book := struct {
			Name      string `json:"name"`
			PageTotal uint16 `json:"pageTotal"`
		}{"foo", 500}
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Json(book)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		result := recorder.Body.String()
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf(testErrorFormat, resultHeader, expectedHeader)
		}
	})
}

func TestContext_Render(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		k, expected, expectedHeader := "Content-Type", "foo", "text/plain; charset=utf-8"
		r := renderer.String{Data: "foo"}
		res := &Context{ResponseWriter: httptest.NewRecorder()}
		res.Render(r)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		result := recorder.Body.String()
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf(testErrorFormat, resultHeader, expectedHeader)
		}
	})

	t.Run("json", func(t *testing.T) {
		k, expectedHeader := "Content-Type", "application/json; charset=utf-8"
		t.Run("normal", func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(`{"Name":"foo","PageTotal":500}`))
			buf.WriteByte('\n')
			expected := buf.String()
			book := struct {
				Name      string
				PageTotal uint16
			}{"foo", 500}
			r := renderer.JSON{Data: book}
			res := &Context{ResponseWriter: httptest.NewRecorder()}
			res.Render(r)
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			result := recorder.Body.String()
			if result != expected {
				t.Errorf(testErrorFormat, result, expected)
			}
			resultHeader := res.Get(k)
			if resultHeader != expectedHeader {
				t.Errorf(testErrorFormat, resultHeader, expectedHeader)
			}
		})

		t.Run("custom", func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(`{"name":"foo","pageTotal":500}`))
			buf.WriteByte('\n')
			expected := buf.String()
			book := struct {
				Name      string `json:"name"`
				PageTotal uint16 `json:"pageTotal"`
			}{"foo", 500}
			r := renderer.JSON{Data: book}
			res := &Context{ResponseWriter: httptest.NewRecorder()}
			res.Render(r)
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			result := recorder.Body.String()
			if result != expected {
				t.Errorf(testErrorFormat, result, expected)
			}
			resultHeader := res.Get(k)
			if resultHeader != expectedHeader {
				t.Errorf(testErrorFormat, resultHeader, expectedHeader)
			}
		})
	})
}

func headerEquals(h1, h2 http.Header) bool {
	if len(h1) != len(h2) {
		return false
	}

	for k, v := range h1 {
		if !stringsEqual(v, h2[k]) {
			return false
		}
	}

	return true
}

func stringsEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, v := range s1 {
		if v != s2[i] {
			return false
		}
	}

	return true
}

func mapToHeader(m map[string]string) http.Header {
	h := map[string][]string{}
	for k, v := range m {
		h[k] = []string{v}
	}
	return h
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
