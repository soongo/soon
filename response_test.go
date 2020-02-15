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
	"testing"
	"time"

	"github.com/soongo/soon/renderer"
)

func TestAppend(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		k, expected := "Content-Type", "application/json"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Append(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("custom", func(t *testing.T) {
		k, expected := "x-custom", "custom"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Append(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		k, expected := "Content-Type", "application/json"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Append(k, []string{expected, expected})
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}

		m := map[string][]string{k: {expected, expected}}
		r := map[string][]string(res.Header())
		if !headerEquals(m, r) {
			t.Errorf("got `%v`, expect `%v`", r, m)
		}
	})
}

func TestSet(t *testing.T) {
	k, expected := "Content-Type", "application/json; charset=UTF-8"

	t.Run("normal", func(t *testing.T) {
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Set(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("replace", func(t *testing.T) {
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Header().Set(k, "text/plain")
		res.Set(k, expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]string{
			"Content-Type": "application/json; charset=UTF-8",
			"X-Custom":     "custom",
		}
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Set(m)
		result, expected := res.Header(), mapToHeader(m)
		if !headerEquals(result, expected) {
			t.Errorf("got `%v`, expect `%v`", result, expected)
		}
	})
}

func TestGet(t *testing.T) {
	k, expected := "Content-Type", "application/json"
	res := &Response{ResponseWriter: httptest.NewRecorder()}
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

func TestStatus(t *testing.T) {
	expected := 404
	res := &Response{ResponseWriter: httptest.NewRecorder()}
	res.Status(expected)
	recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
	result := recorder.Code
	if result != expected {
		t.Errorf("got `%d`, expect `%d`", result, expected)
	}
}

func TestSendStatus(t *testing.T) {
	expectedCode := 302
	expectedBody := http.StatusText(expectedCode)
	res := &Response{ResponseWriter: httptest.NewRecorder()}
	res.SendStatus(expectedCode)
	recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
	resultCode, resultBody := recorder.Code, recorder.Body.String()
	if resultCode != expectedCode {
		t.Errorf("got `%d`, expect `%d`", resultCode, expectedCode)
	}
	if resultBody != expectedBody {
		t.Errorf("got `%s`, expect `%s`", resultBody, expectedBody)
	}
}

func TestType(t *testing.T) {
	k := "Content-Type"
	t.Run("normal", func(t *testing.T) {
		expected := "text/html; charset=UTF-8"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Type("html")
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}

		res.Type("index.html")
		result = res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("slash", func(t *testing.T) {
		expected := "image/png"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Type(expected)
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}

		expected = "/"
		res.Type(expected)
		result = res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("empty", func(t *testing.T) {
		expected := "application/octet-stream"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Type("")
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})
}

func TestAttachment(t *testing.T) {
	k := "Content-Disposition"
	t.Run("name", func(t *testing.T) {
		name := "foo.png"
		expected := fmt.Sprintf("attachment; filename=\"%s\"", name)
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Attachment(name)
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("empty", func(t *testing.T) {
		expected := "attachment"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Attachment()
		result := res.Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})
}

func TestCookie(t *testing.T) {
	k, expected := "Set-Cookie", "foo=bar; Path=/; HttpOnly"
	res := &Response{ResponseWriter: httptest.NewRecorder()}
	c := &http.Cookie{Name: "foo", Value: "bar", Path: "/", HttpOnly: true}
	res.Cookie(c)
	result := res.Get(k)
	if result != expected {
		t.Errorf("got `%s`, expect `%s`", result, expected)
	}
}

func TestClearCookie(t *testing.T) {
	k, expected := "Set-Cookie", "foo=bar; Path=/; HttpOnly"
	res := &Response{ResponseWriter: httptest.NewRecorder()}
	c := &http.Cookie{Name: "foo", Value: "bar", Path: "/", HttpOnly: true}
	res.Cookie(c)
	result := res.Get(k)
	if result != expected {
		t.Errorf("got `%s`, expect `%s`", result, expected)
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
		t.Errorf("got `%v`, expect `%v`", results, expects)
	}
}

func TestSendFile(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	t.Run("normal", func(t *testing.T) {
		filePath := pwd + "/README.md"
		fileInfo, fileContent := getFileContent(filePath)
		lastModified := fileInfo.ModTime().UTC().Format(http.TimeFormat)
		k, expectedHeader := "Content-Type", "text/markdown; charset=UTF-8"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.SendFile(filePath, nil)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		body := recorder.Body.String()
		if body != fileContent {
			t.Errorf("got `%s`, expect `%s`", body, fileContent)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
		}
		if res.Get("Last-Modified") != lastModified {
			t.Errorf("got `%s`, expect `%s`", res.Get("Last-Modified"), lastModified)
		}
	})

	t.Run("hidden", func(t *testing.T) {
		filePath := pwd + "/.travis.yml"
		t.Run("default", func(t *testing.T) {
			res := &Response{ResponseWriter: httptest.NewRecorder()}
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
			res := &Response{ResponseWriter: httptest.NewRecorder()}
			options := &renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyAllow}
			res.SendFile(filePath, options)
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			body := recorder.Body.String()
			_, fileContent := getFileContent(filePath)
			if recorder.Code != http.StatusOK {
				t.Errorf("got `%d`, expect `%d`", recorder.Code, http.StatusOK)
			}
			if body != fileContent {
				t.Errorf("got `%s`, expect `%s`", body, fileContent)
			}
		})
		t.Run("deny", func(t *testing.T) {
			res := &Response{ResponseWriter: httptest.NewRecorder()}
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

func TestDownload(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filePath := pwd + "/README.md"

	t.Run("normal", func(t *testing.T) {
		fileInfo, fileContent := getFileContent(filePath)
		lastModified := fileInfo.ModTime().UTC().Format(http.TimeFormat)
		k, expectedHeader := "Content-Type", "text/markdown; charset=UTF-8"
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Download(filePath, nil)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		body := recorder.Body.String()
		if body != fileContent {
			t.Errorf("got `%s`, expect `%s`", body, fileContent)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
		}
		if res.Get("Last-Modified") != lastModified {
			t.Errorf("got `%s`, expect `%s`", res.Get("Last-Modified"), lastModified)
		}
		contentDisposition := `attachment; filename="README.md"`
		if res.Get("Content-Disposition") != contentDisposition {
			t.Errorf("got `%s`, expect `%s`", res.Get("Content-Disposition"), contentDisposition)
		}
	})

	t.Run("custom name", func(t *testing.T) {
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		options := &renderer.FileOptions{Name: "custom-name"}
		res.Download(filePath, options)
		contentDisposition := `attachment; filename="` + options.Name + `"`
		if res.Get("Content-Disposition") != contentDisposition {
			t.Errorf("got `%s`, expect `%s`", res.Get("Content-Disposition"), contentDisposition)
		}
	})
}

func TestEnd(t *testing.T) {
	t.Run("without end", func(t *testing.T) {
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Send("foo")
		res.Send("bar")
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		expected, body := "foobar", recorder.Body.String()
		if body != expected {
			t.Errorf("got `%s`, expect `%s`", body, expected)
		}
	})

	t.Run("without end", func(t *testing.T) {
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Send("foo")
		res.End()
		res.Send("bar")
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		expected, body := "foo", recorder.Body.String()
		if body != expected {
			t.Errorf("got `%s`, expect `%s`", body, expected)
		}
	})
}

func TestString(t *testing.T) {
	k, expected, expectedHeader := "Content-Type", "foo", "text/plain; charset=utf-8"
	res := &Response{ResponseWriter: httptest.NewRecorder()}
	res.String(expected)
	recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
	result := recorder.Body.String()
	if result != expected {
		t.Errorf("got `%s`, expect `%s`", result, expected)
	}
	resultHeader := res.Get(k)
	if resultHeader != expectedHeader {
		t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
	}
}

func TestJson(t *testing.T) {
	k, expectedHeader := "Content-Type", "application/json; charset=utf-8"

	t.Run("normal", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(`{"Name":"foo","PageTotal":500}`))
		buf.WriteByte('\n')
		expected := buf.String()
		book := struct {
			Name      string
			PageTotal uint16
		}{"foo", 500}
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Json(book)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		result := recorder.Body.String()
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
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
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Json(book)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		result := recorder.Body.String()
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
		}
	})
}

func TestRender(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		k, expected, expectedHeader := "Content-Type", "foo", "text/plain; charset=utf-8"
		r := renderer.String{Data: "foo"}
		res := &Response{ResponseWriter: httptest.NewRecorder()}
		res.Render(r)
		recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
		result := recorder.Body.String()
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
		resultHeader := res.Get(k)
		if resultHeader != expectedHeader {
			t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
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
			res := &Response{ResponseWriter: httptest.NewRecorder()}
			res.Render(r)
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			result := recorder.Body.String()
			if result != expected {
				t.Errorf("got `%s`, expect `%s`", result, expected)
			}
			resultHeader := res.Get(k)
			if resultHeader != expectedHeader {
				t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
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
			res := &Response{ResponseWriter: httptest.NewRecorder()}
			res.Render(r)
			recorder := res.ResponseWriter.(*httptest.ResponseRecorder)
			result := recorder.Body.String()
			if result != expected {
				t.Errorf("got `%s`, expect `%s`", result, expected)
			}
			resultHeader := res.Get(k)
			if resultHeader != expectedHeader {
				t.Errorf("got `%s`, expect `%s`", resultHeader, expectedHeader)
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
