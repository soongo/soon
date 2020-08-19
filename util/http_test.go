// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, tt.expected, w.Header()[tt.k])
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
		assert.Equal(t, tt.expected, w.Header())
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
		assert.Equal(t, tt.expectedContentType, w.Header().Get("Content-Type"))
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
		assert.Equal(t, tt.expected, result)
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
		assert.Equal(t, tt.expected, AppendToVaryHeader(tt.vary, tt.fields))
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
		assert.Equal(t, tt.expected, ParseHeader(tt.header))
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
		assert.Equal(t, tt.expected, GetHeaderValues(tt.header, k))
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
		assert.Equal(t, tt.expected, NormalizeType(tt.t))
	}
}
