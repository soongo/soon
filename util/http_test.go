// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSetHeader(t *testing.T) {
	k, expected := "Content-Type", "application/json; charset=UTF-8"

	t.Run("normal", func(t *testing.T) {
		res := httptest.NewRecorder()
		SetHeader(res, k, expected)
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("replace", func(t *testing.T) {
		res := httptest.NewRecorder()
		res.Header().Set(k, "text/plain")
		SetHeader(res, k, expected)
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
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
		if !reflect.DeepEqual(result, expected) {
			t.Errorf(testErrorFormat, result, expected)
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
			t.Errorf(testErrorFormat, result, expected)
		}

		SetContentType(res, "index.html")
		result = res.Header().Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("slash", func(t *testing.T) {
		expected := "image/png"
		res := httptest.NewRecorder()
		SetContentType(res, expected)
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}

		expected = "/"
		SetContentType(res, expected)
		result = res.Header().Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})

	t.Run("empty", func(t *testing.T) {
		expected := "application/octet-stream"
		res := httptest.NewRecorder()
		SetContentType(res, "")
		result := res.Header().Get(k)
		if result != expected {
			t.Errorf(testErrorFormat, result, expected)
		}
	})
}

func TestVary(t *testing.T) {
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
		w := httptest.NewRecorder()
		w.Header().Set(key, tt.vary)
		Vary(w, tt.fields)
		result := w.Header().Get(key)
		if result != tt.expected {
			t.Errorf(testErrorFormat, result, tt.expected)
		}
	}
}

func TestAppendToVaryHeader(t *testing.T) {
	tests := []struct {
		vary     string
		fields   []string
		expected string
	}{
		{"", []string{"foo", "bar"}, "foo, bar"},
		{"foo", []string{"bar"}, "foo, bar"},
		{"foo", []string{"foo", "bar"}, "foo, bar"},
		{"foo,bar", []string{"foo", "bar"}, "foo,bar"},
		{"foo,bar", []string{"foo", "host"}, "foo,bar, host"},
		{"foo,bar", []string{"你好", "bar"}, "foo,bar, 你好"},
	}
	for _, tt := range tests {
		if got := AppendToVaryHeader(tt.vary, tt.fields); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		header   string
		expected []string
	}{
		{"foo,bar", []string{"foo", "bar"}},
		{" foo, bar ", []string{"foo", "bar"}},
		{" foo, 你好,bar ", []string{"foo", "你好", "bar"}},
		{" foo,你好 ,bar ", []string{"foo", "你好", "bar"}},
	}
	for _, tt := range tests {
		if got := ParseHeader(tt.header); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func mapToHeader(m map[string]string) http.Header {
	h := map[string][]string{}
	for k, v := range m {
		h[k] = []string{v}
	}
	return h
}
