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

func TestFresh(t *testing.T) {
	tests := []struct {
		desc      string
		reqHeader http.Header
		resHeader http.Header
		expected  bool
	}{
		{
			"when a non-conditional GET is performed, it should be stale",
			http.Header{},
			http.Header{},
			false,
		},
		{
			"when requested with If-None-Match, and ETags match, it should be fresh",
			http.Header{H("if-none-match"): []string{`"foo"`}},
			http.Header{H("etag"): []string{`"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and ETags mismatch, it should be stale",
			http.Header{H("if-none-match"): []string{`"foo"`}},
			http.Header{H("etag"): []string{`"bar"`}},
			false,
		},
		{
			"when requested with If-None-Match, and at least one matches, it should be fresh",
			http.Header{H("if-none-match"): []string{` "bar" , "foo"`}},
			http.Header{H("etag"): []string{`"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and ETag is missing, it should be stale",
			http.Header{H("if-none-match"): []string{`"foo"`}},
			http.Header{},
			false,
		},
		{
			"when requested with If-None-Match, and ETag is weak, it should be fresh on exact match",
			http.Header{H("if-none-match"): []string{`W/"foo"`}},
			http.Header{H("etag"): []string{`W/"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and ETag is weak, it should be fresh on strong match",
			http.Header{H("if-none-match"): []string{`W/"foo"`}},
			http.Header{H("etag"): []string{`"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and ETag is strong, it should be fresh on exact match",
			http.Header{H("if-none-match"): []string{`"foo"`}},
			http.Header{H("etag"): []string{`"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and ETag is strong, it should be fresh on weak match",
			http.Header{H("if-none-match"): []string{`"foo"`}},
			http.Header{H("etag"): []string{`W/"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and * is given, it should be fresh",
			http.Header{H("if-none-match"): []string{`*`}},
			http.Header{H("etag"): []string{`"foo"`}},
			true,
		},
		{
			"when requested with If-None-Match, and * is given, it should get ignored if not only value",
			http.Header{H("if-none-match"): []string{`*, "bar"`}},
			http.Header{H("etag"): []string{`"foo"`}},
			false,
		},
		{
			"when requested with If-Modified-Since, and modified since the date, it should be stale",
			http.Header{H("if-modified-since"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
			http.Header{H("last-modified"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"}},
			false,
		},
		{
			"when requested with If-Modified-Since, and unmodified since the date, it should be fresh",
			http.Header{H("if-modified-since"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"}},
			http.Header{H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
			true,
		},
		{
			"when requested with If-Modified-Since, and Last-Modified is missing, it should be stale",
			http.Header{H("if-modified-since"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
			http.Header{},
			false,
		},
		{
			"when requested with If-Modified-Since, and with invalid If-Modified-Since date, it should be stale",
			http.Header{H("if-modified-since"): []string{"foo"}},
			http.Header{H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
			false,
		},
		{
			"when requested with If-Modified-Since, and with invalid Last-Modified date, it should be stale",
			http.Header{H("if-modified-since"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
			http.Header{H("last-modified"): []string{"foo"}},
			false,
		},
		{
			"when requested with If-Modified-Since and If-None-Match, and both match, it should be fresh",
			http.Header{
				H("if-none-match"):     []string{`"foo"`},
				H("if-modified-since"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"},
			},
			http.Header{
				H("etag"):          []string{`"foo"`},
				H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"},
			},
			true,
		},
		{
			"when requested with If-Modified-Since and If-None-Match, and only ETag matches, it should be stale",
			http.Header{
				H("if-none-match"):     []string{`"foo"`},
				H("if-modified-since"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"},
			},
			http.Header{
				H("etag"):          []string{`"foo"`},
				H("last-modified"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"},
			},
			false,
		},
		{
			"when requested with If-Modified-Since and If-None-Match, and only Last-Modified matches, it should be stale",
			http.Header{
				H("if-none-match"):     []string{`"foo"`},
				H("if-modified-since"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"},
			},
			http.Header{
				H("etag"):          []string{`"bar"`},
				H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"},
			},
			false,
		},
		{
			"when requested with If-Modified-Since and If-None-Match, and none match, it should be stale",
			http.Header{
				H("if-none-match"):     []string{`"foo"`},
				H("if-modified-since"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"},
			},
			http.Header{
				H("etag"):          []string{`"bar"`},
				H("last-modified"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"},
			},
			false,
		},
		{
			"when requested with Cache-Control: no-cache, it should be stale",
			http.Header{H("cache-control"): []string{" no-cache"}},
			http.Header{},
			false,
		},
		{
			"when requested with Cache-Control: no-cache, and ETags match, it should be stale",
			http.Header{
				H("cache-control"): []string{" no-cache"},
				H("if-none-match"): []string{`"foo"`},
			},
			http.Header{H("etag"): []string{`"foo"`}},
			false,
		},
		{
			"when requested with Cache-Control: no-cache, and unmodified since the date, it should be stale",
			http.Header{
				H("cache-control"):     []string{" no-cache"},
				H("if-modified-since"): []string{"Sat, 01 Jan 2000 01:00:00 GMT"},
			},
			http.Header{H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
			false,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, Fresh(tt.reqHeader, tt.resHeader))
	}
}

func BenchmarkFresh(b *testing.B) {
	b.Run("etag", func(b *testing.B) {
		b.Run("star", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Fresh(
					http.Header{H("if-none-match"): []string{`"*"`}},
					http.Header{H("etag"): []string{`"foo"`}},
				)
			}
		})
		b.Run("single etag", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Fresh(
					http.Header{H("if-none-match"): []string{`"foo"`}},
					http.Header{H("etag"): []string{`"foo"`}},
				)
			}
		})
		b.Run("several etags", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Fresh(
					http.Header{H("if-none-match"): []string{`"foo", "bar", "fizz", "buzz"`}},
					http.Header{H("etag"): []string{`"buzz"`}},
				)
			}
		})
	})
	b.Run("modified", func(b *testing.B) {
		b.Run("not modified", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Fresh(
					http.Header{H("if-modified-since"): []string{"Fri, 01 Jan 2010 00:00:00 GMT"}},
					http.Header{H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
				)
			}
		})
		b.Run("modified", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Fresh(
					http.Header{H("if-modified-since"): []string{"Mon, 01 Jan 1990 00:00:00 GMT"}},
					http.Header{H("last-modified"): []string{"Sat, 01 Jan 2000 00:00:00 GMT"}},
				)
			}
		})
	})
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

func TestRequestTypeIs(t *testing.T) {
	createRequest := func(contentType string, noBody bool) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		if contentType != "" {
			req.Header.Set("content-type", contentType)
		}
		if !noBody {
			req.Header.Set("transfer-encoding", "chunked")
		}
		return req
	}

	tests := []struct {
		desc        string
		contentType string
		noBody      bool
		types       []string
		expected    string
	}{
		{
			desc:        "should ignore params",
			contentType: "text/html; charset=utf-8",
			types:       []string{"text/*"},
			expected:    "text/html",
		},
		{
			desc:        "should ignore params LWS",
			contentType: "text/html ; charset=utf-8",
			types:       []string{"text/*"},
			expected:    "text/html",
		},
		{
			desc:        "should ignore casing",
			contentType: "text/HTML",
			types:       []string{"text/*"},
			expected:    "text/html",
		},
		{
			desc:        "should fail invalid type",
			contentType: "text/html**",
			types:       []string{"text/*"},
			expected:    "",
		},
		{
			desc:        "should not match invalid type",
			contentType: "text/html",
			types:       []string{"text/html/"},
			expected:    "",
		},
		{
			desc:        "should return empty string when no body is given",
			contentType: "text/html",
			noBody:      true,
			expected:    "",
		},
		{
			desc:        "should return empty string when no body is given",
			contentType: "text/html",
			noBody:      true,
			types:       []string{"image/*"},
			expected:    "",
		},
		{
			desc:        "should return empty string when no body is given",
			contentType: "text/html",
			noBody:      true,
			types:       []string{"image/*", "text/*"},
			expected:    "",
		},
		{
			desc:        "should return empty string when no content type is given",
			contentType: "",
			expected:    "",
		},
		{
			desc:        "should return empty string when no content type is given",
			contentType: "",
			types:       []string{"text/*", "image/*"},
			expected:    "",
		},
		{
			desc:        "should return the mime type when given no types",
			contentType: "image/png",
			expected:    "image/png",
		},
		{
			desc:        "should return the type when given one type",
			contentType: "image/png",
			types:       []string{"png"},
			expected:    "png",
		},
		{
			desc:        "should return the type when given one type",
			contentType: "image/png",
			types:       []string{".png"},
			expected:    ".png",
		},
		{
			desc:        "should return the type when given one type",
			contentType: "image/png",
			types:       []string{"image/png"},
			expected:    "image/png",
		},
		{
			desc:        "should return the type when given one type",
			contentType: "image/png",
			types:       []string{"image/*"},
			expected:    "image/png",
		},
		{
			desc:        "should return the type when given one type",
			contentType: "image/png",
			types:       []string{"*/png"},
			expected:    "image/png",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{"jpeg"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{".jpeg"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{"image/jpeg"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{"text/*"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{"*/jpeg"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{"bogus"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given one type",
			contentType: "image/png",
			types:       []string{"something/bogus*"},
			expected:    "",
		},
		{
			desc:        "should return the first match when given multiple types",
			contentType: "image/png",
			types:       []string{"text/*", "image/*"},
			expected:    "image/png",
		},
		{
			desc:        "should return the first match when given multiple types",
			contentType: "image/png",
			types:       []string{"image/*", "text/*"},
			expected:    "image/png",
		},
		{
			desc:        "should return the first match when given multiple types",
			contentType: "image/png",
			types:       []string{"image/*", "image/png"},
			expected:    "image/png",
		},
		{
			desc:        "should return the first match when given multiple types",
			contentType: "image/png",
			types:       []string{"image/png", "image/*"},
			expected:    "image/png",
		},
		{
			desc:        "should return empty string when given multiple types",
			contentType: "image/png",
			types:       []string{"text/*", "application/*"},
			expected:    "",
		},
		{
			desc:        "should return empty string when given multiple types",
			contentType: "image/png",
			types:       []string{"text/html", "text/plain", "application/json"},
			expected:    "",
		},
		{
			desc:        "should match suffix types when given +suffix",
			contentType: "application/vnd+json",
			types:       []string{"+json"},
			expected:    "application/vnd+json",
		},
		{
			desc:        "should match suffix types when given +suffix",
			contentType: "application/vnd+json",
			types:       []string{"application/vnd+json"},
			expected:    "application/vnd+json",
		},
		{
			desc:        "should match suffix types when given +suffix",
			contentType: "application/vnd+json",
			types:       []string{"application/*+json"},
			expected:    "application/vnd+json",
		},
		{
			desc:        "should match suffix types when given +suffix",
			contentType: "application/vnd+json",
			types:       []string{"*/vnd+json"},
			expected:    "application/vnd+json",
		},
		{
			desc:        "should match suffix types when given +suffix",
			contentType: "application/vnd+json",
			types:       []string{"application/json"},
			expected:    "",
		},
		{
			desc:        "should match suffix types when given +suffix",
			contentType: "application/vnd+json",
			types:       []string{"text/*+json"},
			expected:    "",
		},
		{
			desc:        "should match any content-type when given '*/*'",
			contentType: "text/html",
			types:       []string{"*/*"},
			expected:    "text/html",
		},
		{
			desc:        "should match any content-type when given '*/*'",
			contentType: "text/xml",
			types:       []string{"*/*"},
			expected:    "text/xml",
		},
		{
			desc:        "should match any content-type when given '*/*'",
			contentType: "application/json",
			types:       []string{"*/*"},
			expected:    "application/json",
		},
		{
			desc:        "should match any content-type when given '*/*'",
			contentType: "application/vnd+json",
			types:       []string{"*/*"},
			expected:    "application/vnd+json",
		},
		{
			desc:        "should not match invalid content-type when given '*/*'",
			contentType: "bogus",
			types:       []string{"*/*"},
			expected:    "",
		},
		{
			desc:        "should not match body-less request when given '*/*'",
			contentType: "text/html",
			noBody:      true,
			types:       []string{"*/*"},
			expected:    "",
		},
		{
			desc:        "should match 'urlencoded' when Content-Type: application/x-www-form-urlencoded",
			contentType: "application/x-www-form-urlencoded",
			types:       []string{"urlencoded"},
			expected:    "urlencoded",
		},
		{
			desc:        "should match 'urlencoded' when Content-Type: application/x-www-form-urlencoded",
			contentType: "application/x-www-form-urlencoded",
			types:       []string{"json", "urlencoded"},
			expected:    "urlencoded",
		},
		{
			desc:        "should match 'urlencoded' when Content-Type: application/x-www-form-urlencoded",
			contentType: "application/x-www-form-urlencoded",
			types:       []string{"urlencoded", "json"},
			expected:    "urlencoded",
		},
		{
			desc:        "should match 'multipart/*' when Content-Type: multipart/form-data",
			contentType: "multipart/form-data",
			types:       []string{"multipart/*"},
			expected:    "multipart/form-data",
		},
		{
			desc:        "should match 'multipart' when Content-Type: multipart/form-data",
			contentType: "multipart/form-data",
			types:       []string{"multipart"},
			expected:    "multipart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := RequestTypeIs(createRequest(tt.contentType, tt.noBody), tt.types...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasBody(t *testing.T) {
	createRequest := func(contentLength, transferEncoding string) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		if contentLength != "" {
			req.Header.Set("content-length", contentLength)
		}
		if transferEncoding != "" {
			req.Header.Set("transfer-encoding", transferEncoding)
		}
		return req
	}

	tests := []struct {
		desc             string
		contentLength    string
		transferEncoding string
		expected         bool
	}{
		{desc: "should indicate body", contentLength: "1", expected: true},
		{desc: "should be true when 0", contentLength: "0", expected: true},
		{desc: "should be false when bogus", contentLength: "bogus", expected: false},
		{desc: "should indicate body", transferEncoding: "chunked", expected: true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, HasBody(createRequest(tt.contentLength, tt.transferEncoding)))
	}
}
