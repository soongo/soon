// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/soongo/negotiator"
)

func TestParams_Get(t *testing.T) {
	tests := []struct {
		p        Params
		k        interface{}
		expected string
	}{
		{Params{"name": "foo", 0: "bar"}, "name", "foo"},
		{Params{"name": "foo", 0: "bar"}, 0, "bar"},
		{Params{"name": "foo", 0: "bar"}, 1, ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.p.Get(tt.k))
	}
}

func TestParams_Set(t *testing.T) {
	tests := []struct {
		p        Params
		k        interface{}
		v        string
		expected string
	}{
		{Params{}, "name", "foo", "foo"},
		{Params{0: "bar"}, "name", "foo", "foo"},
		{Params{"name": "foo", 0: "bar"}, 0, "foo", "foo"},
	}

	for _, tt := range tests {
		tt.p.Set(tt.k, tt.v)
		assert.Equal(t, tt.expected, tt.p.Get(tt.k))
	}
}

func TestParams_MarshalJSON(t *testing.T) {
	tests := []struct {
		p        Params
		expected string
	}{
		{Params{}, "{}"},
		{Params{"name": "foo", 0: "bar"}, `{"0":"bar","name":"foo"}`},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		bts, err := tt.p.MarshalJSON()
		assert.Nil(err)
		assert.Equal(tt.expected, string(bts))
	}
}

func TestRequest_GetHeader(t *testing.T) {
	tests := []struct {
		header string
		value  string
	}{
		{"Accept", "text/html"},
		{"Content-Type", " text/html "},
		{"Accept", "text/html; "},
		{"Content-Type", " text/html ; "},
		{"Accept", " text/html ; text/xml"},
		{"Content-Type", " "},
		{"Accept", "     "},
		{"Content-Type", "     "},
		{"Accept-Charset", "utf-8, iso-8859-1;q=0.8, utf-7;q=0.2"},
		{"Accept-Charset", " utf-8, iso-8859-1;q=0.8, utf-7;q=0.2 "},
		{"Accept-Encoding", "gzip, compress;q=0.8, identity;q=0.2"},
		{"Accept-Encoding", " gzip, compress;q=0.8, identity;q=0.2 "},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(tt.header, tt.value)
		r := NewRequest(req)
		assert.Equal(t, tt.value, r.GetHeader(tt.header))
	}
}

func TestRequest_ContentType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    string
	}{
		{"text/html", "text/html"},
		{" text/html ", "text/html"},
		{"text/html; ", "text/html"},
		{" text/html ; ", "text/html"},
		{" text/html ; text/xml", "text/html"},
		{" ", ""},
		{"     ", ""},
	}
	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Content-Type", tt.contentType)
		r := NewRequest(req)
		assert.Equal(t, tt.expected, r.ContentType())
	}
}

func TestRequest_Accepts(t *testing.T) {
	tests := []struct {
		accept   []string
		types    []string
		expected []string
	}{
		{[]string{"text/html"}, []string{"html"}, []string{"html"}},
		{nil, []string{"html", "image/png"}, []string{"html"}},
		{[]string{}, []string{"html", "image/png"}, []string{"html"}},
		{[]string{"text/*", "image/png"}, []string{"html"}, []string{"html"}},
		{[]string{"text/*", "image/png"}, []string{"text/html"}, []string{"text/html"}},
		{[]string{"text/*", "image/png"}, []string{"png", "text"}, []string{"png"}},
		{[]string{"text/*", "image/png"}, []string{"image/png"}, []string{"image/png"}},
		{[]string{"text/*", "image/png"}, []string{"image/jpg"}, nil},
		{[]string{"text/*", "image/png"}, []string{"jpg"}, nil},
		{[]string{"text/*;q=.5", "image/png"}, []string{"html", "png"}, []string{"png"}},
		{[]string{"text/*", "image/png"}, nil, []string{"text/*", "image/png"}},
		{[]string{"text/*", "image/png"}, []string{}, []string{"text/*", "image/png"}},
		{[]string{"text/*;q=.5", "image/png"}, nil, []string{"image/png", "text/*"}},
	}

	req := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	for _, tt := range tests {
		req.Header = http.Header{negotiator.HeaderAccept: tt.accept}
		assert.Equal(t, tt.expected, req.Accepts(tt.types...))
	}
}

func TestRequest_AcceptsEncodings(t *testing.T) {
	tests := []struct {
		accept    string
		encodings []string
		expected  []string
	}{
		{"gzip", []string{"gzip"}, []string{"gzip"}},
		{"gzip, compress", []string{"gzip"}, []string{"gzip"}},
		{"gzip, compress", []string{"compress"}, []string{"compress"}},
		{"gzip, compress", []string{"gzip", "compress"}, []string{"gzip"}},
		{"gzip, compress", []string{"compress", "gzip"}, []string{"gzip"}},
		{"gzip, compress", []string{"identity"}, []string{"identity"}},
		{"gzip, compress", nil, []string{"gzip", "compress", "identity"}},
		{"gzip, compress", []string{}, []string{"gzip", "compress", "identity"}},
		{"gzip, compress", []string{"deflate"}, nil},
		{"gzip, compress;q=0.8", []string{"compress"}, []string{"compress"}},
		{"gzip;q=0.5, compress;q=0.8", nil, []string{"compress", "gzip", "identity"}},
		{"gzip;q=0.5, compress;q=0.8", []string{"gzip", "compress"}, []string{"compress"}},
	}

	req := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	for _, tt := range tests {
		req.Header = http.Header{negotiator.HeaderAcceptEncoding: dotRegexp.Split(tt.accept, -1)}
		assert.Equal(t, tt.expected, req.AcceptsEncodings(tt.encodings...))
	}
}

func TestRequest_AcceptsCharsets(t *testing.T) {
	tests := []struct {
		accept   string
		charsets []string
		expected []string
	}{
		{"utf-8", []string{"utf-8"}, []string{"utf-8"}},
		{"utf-8, iso-8859-1", []string{"utf-8"}, []string{"utf-8"}},
		{"utf-8, iso-8859-1", []string{"iso-8859-1"}, []string{"iso-8859-1"}},
		{"utf-8, iso-8859-1", []string{"utf-8", "iso-8859-1"}, []string{"utf-8"}},
		{"utf-8, iso-8859-1", []string{"iso-8859-1", "utf-8"}, []string{"utf-8"}},
		{"utf-8, iso-8859-1", []string{"utf-7"}, nil},
		{"utf-8, iso-8859-1", nil, []string{"utf-8", "iso-8859-1"}},
		{"utf-8, iso-8859-1", []string{}, []string{"utf-8", "iso-8859-1"}},
		{"utf-8, iso-8859-1;q=0.8", []string{"iso-8859-1"}, []string{"iso-8859-1"}},
		{"utf-8;q=0.5, iso-8859-1;q=0.8", nil, []string{"iso-8859-1", "utf-8"}},
		{"utf-8;q=0.5, iso-8859-1;q=0.8", []string{"utf-8", "iso-8859-1"}, []string{"iso-8859-1"}},
	}

	req := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	for _, tt := range tests {
		req.Header = http.Header{negotiator.HeaderAcceptCharset: dotRegexp.Split(tt.accept, -1)}
		assert.Equal(t, tt.expected, req.AcceptsCharsets(tt.charsets...))
	}
}

func TestRequest_AcceptsLanguages(t *testing.T) {
	tests := []struct {
		accept    string
		languages []string
		expected  []string
	}{
		{"en", []string{"en"}, []string{"en"}},
		{"en, zh", []string{"en"}, []string{"en"}},
		{"en, zh", []string{"zh"}, []string{"zh"}},
		{"en, zh", []string{"en", "zh"}, []string{"en"}},
		{"en, zh", []string{"zh", "en"}, []string{"en"}},
		{"en, zh", []string{"fr"}, nil},
		{"en, zh", nil, []string{"en", "zh"}},
		{"en, zh", []string{}, []string{"en", "zh"}},
		{"en, zh;q=0.8", []string{"zh"}, []string{"zh"}},
		{"en;q=0.5, zh;q=0.8", nil, []string{"zh", "en"}},
		{"en;q=0.5, zh;q=0.8", []string{"en", "zh"}, []string{"zh"}},
	}

	req := NewRequest(httptest.NewRequest(http.MethodGet, "/", nil))
	for _, tt := range tests {
		req.Header = http.Header{negotiator.HeaderAcceptLanguage: dotRegexp.Split(tt.accept, -1)}
		assert.Equal(t, tt.expected, req.AcceptsLanguages(tt.languages...))
	}
}

func TestRequest_ResetParams(t *testing.T) {
	tests := []struct {
		p        Params
		expected Params
	}{
		{nil, Params{}},
		{Params{}, Params{}},
		{Params{"name": "foo"}, Params{}},
	}

	for _, tt := range tests {
		req := NewRequest(httptest.NewRequest("GET", "/", nil))
		req.Params = tt.p
		req.resetParams()
		assert.Equal(t, tt.expected, req.Params)
	}
}

func TestRequest_BaseUrl(t *testing.T) {
	tests := []struct {
		route0          string
		route1          string
		route2          string
		expectedBaseUrl string
		path            string
	}{
		{"/", "", "", "", "/"},
		{"/foo", "/bar", "", "/foo", "/foo/bar"},
		{"/foo", "/bar", "/test", "/foo/bar", "/foo/bar/test"},
		{"/:foo", "/:bar", "", "/foo", "/foo/bar"},
		{"/:foo", "/:bar", "/:test", "/foo/bar", "/foo/bar/test"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert := assert.New(t)
			router0 := NewRouter()
			if tt.route1 != "" {
				router1 := NewRouter()
				if tt.route2 != "" {
					router2 := NewRouter()
					router2.Use(tt.route2, func(c *Context) {
						assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
						c.Next()
					})
					router2.GET(tt.route2, func(c *Context) {
						assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
					})
					router1.Use(tt.route1, router2)
				} else {
					router1.Use(tt.route1, func(c *Context) {
						assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
						c.Next()
					})
					router1.GET(tt.route1, func(c *Context) {
						assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
					})
				}
				router0.Use(tt.route0, router1)
			} else {
				router0.Use(tt.route0, func(c *Context) {
					assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
					c.Next()
				})
				router0.GET(tt.route0, func(c *Context) {
					assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
				})
			}
			server := httptest.NewServer(router0)
			defer server.Close()
			_, _, _, err := request("GET", server.URL+tt.path, nil)
			assert.Nil(err)
		})
	}
}
