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

func TestAddHeader(t *testing.T) {
	tests := []struct {
		k        string
		v        interface{}
		expected []string
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
		w := httptest.NewRecorder()
		AddHeader(w, tt.k, tt.v)
		if got := w.Header()[tt.k]; !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestSetHeader(t *testing.T) {
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
		w := httptest.NewRecorder()
		w.Header().Set(k, "text/*")
		if tt.k == "" {
			SetHeader(w, tt.v)
		} else {
			SetHeader(w, tt.k, tt.v)
		}
		if got := w.Header(); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestSetContentType(t *testing.T) {
	tests := []struct {
		name                string
		contentType         string
		expectedContentType string
	}{
		{"normal-0", "html", "text/html; charset=UTF-8"},
		{"normal-1", "index.html", "text/html; charset=UTF-8"},
		{"slash-0", "image/png", "image/png"},
		{"slash-1", "/", "/"},
		{"empty", "", "application/octet-stream"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		SetContentType(w, tt.contentType)
		if got := w.Header().Get("Content-Type"); got != tt.expectedContentType {
			t.Errorf(testErrorFormat, got, tt.expectedContentType)
		}
	}
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
		{"foo", nil, "foo"},
		{"*", []string{"foo"}, "*"},
		{"foo", []string{"foo", "*"}, "*"},
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

func TestGetHeaderValues(t *testing.T) {
	k := "Content-Type"
	tests := []struct {
		header   http.Header
		expected []string
	}{
		{nil, nil},
		{http.Header{k: []string{"text/*", "image/png"}}, []string{"text/*", "image/png"}},
	}

	for _, tt := range tests {
		if got := GetHeaderValues(tt.header, k); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		t        string
		expected AcceptParams
	}{
		{"html", AcceptParams{"text/html", 0, nil}},
		{"text/html", AcceptParams{"text/html", 1, nil}},
		{"text/html;q=0.8", AcceptParams{"text/html", .8, nil}},
		{"text/html;p=0.8", AcceptParams{"text/html", 1, map[string]string{"p": "0.8"}}},
		{"***", AcceptParams{"application/octet-stream", 0, nil}},
	}

	for _, tt := range tests {
		if got := NormalizeType(tt.t); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}
