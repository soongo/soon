// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetHeader(t *testing.T) {
	k, expected := "Content-Type", "application/json; charset=UTF-8"

	t.Run("normal", func(t *testing.T) {
		res := httptest.NewRecorder()
		SetHeader(res, k, expected)
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("replace", func(t *testing.T) {
		res := httptest.NewRecorder()
		res.Header().Set(k, "text/plain")
		SetHeader(res, k, expected)
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]string{
			"Content-Type": "application/json; charset=UTF-8",
			"X-Custom":     "custom",
		}
		res := httptest.NewRecorder()
		SetHeader(res, m)
		result, expected := res.Header(), mapToHeader(m)
		if !headerEquals(result, expected) {
			t.Errorf("got `%v`, expect `%v`", result, expected)
		}
	})
}

func TestSetContentType(t *testing.T) {
	k := "Content-Type"
	t.Run("normal", func(t *testing.T) {
		expected := "text/html; charset=UTF-8"
		res := httptest.NewRecorder()
		SetContentType(res, "html")
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}

		SetContentType(res, "index.html")
		result = res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("slash", func(t *testing.T) {
		expected := "image/png"
		res := httptest.NewRecorder()
		SetContentType(res, expected)
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}

		expected = "/"
		SetContentType(res, expected)
		result = res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})

	t.Run("empty", func(t *testing.T) {
		expected := "application/octet-stream"
		res := httptest.NewRecorder()
		SetContentType(res, "")
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf("got `%s`, expect `%s`", result, expected)
		}
	})
}

func mapToHeader(m map[string]string) http.Header {
	h := map[string][]string{}
	for k, v := range m {
		h[k] = []string{v}
	}
	return h
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
