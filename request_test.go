// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/soongo/negotiator"
	"github.com/soongo/soon/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRequest_Get(t *testing.T) {
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
		assert.Equal(t, tt.value, r.Get(tt.header))
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
	t.Run("nested", func(t *testing.T) {
		tests := []struct {
			route0                    string
			route1                    string
			route2                    string
			expectedMiddlewareBaseUrl string
			expectedBaseUrl           string
			path                      string
		}{
			{"/", "", "", "", "", "/"},
			{"/", "", "", "", "", "/foo"},
			{"/", "/foo", "", "/foo", "", "/foo"},
			{"/:foo", "", "", "/foo", "", "/foo"},
			{"/foo", "/bar", "", "/foo/bar", "/foo", "/foo/bar"},
			{"/foo", "/bar", "/test", "/foo/bar/test", "/foo/bar", "/foo/bar/test"},
			{"/:foo", "/:bar", "", "/foo/bar", "/foo", "/foo/bar"},
			{"/:foo", "/:bar", "/:test", "/foo/bar/test", "/foo/bar", "/foo/bar/test"},
			{"/:foo", "/:bar", "/:test", "/foo/bar/test", "/foo/bar", "/foo/bar/test/123"},
			{"/:foo", "/:bar", "/:test/(.*)", "/foo/bar/test/123", "/foo/bar", "/foo/bar/test/123"},
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
							assert.Equal(tt.expectedMiddlewareBaseUrl, c.Request.BaseUrl)
							c.Next()
						})
						router2.GET(tt.route2, func(c *Context) {
							assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
						})
						router1.Use(tt.route1, router2)
					} else {
						router1.Use(tt.route1, func(c *Context) {
							assert.Equal(tt.expectedMiddlewareBaseUrl, c.Request.BaseUrl)
							c.Next()
						})
						router1.GET(tt.route1, func(c *Context) {
							assert.Equal(tt.expectedBaseUrl, c.Request.BaseUrl)
						})
					}
					router0.Use(tt.route0, router1)
				} else {
					router0.Use(tt.route0, func(c *Context) {
						assert.Equal(tt.expectedMiddlewareBaseUrl, c.Request.BaseUrl)
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
	})

	t.Run("parallel", func(t *testing.T) {
		tests := []struct {
			middlewareRoute           string
			route                     string
			expectedMiddlewareBaseUrl string
			expectedBaseUrl           string
			path                      string
		}{
			{"/", "/", "", "", "/"},
			{"/", "/", "", "", "/foo"},
			{"/", "/foo", "", "", "/foo"},
			{"/:foo", "/:foo", "/foo", "", "/foo"},
			{"/foo", "/foo/bar", "/foo", "", "/foo/bar"},
			{"/foo/bar", "/foo/bar/test", "/foo/bar", "", "/foo/bar/test"},
			{"/:foo/:bar", "/:foo/:bar", "/foo/bar", "", "/foo/bar"},
			{"/:foo/:bar", "/:foo/:bar/(.*)", "/foo/bar", "", "/foo/bar/test"},
			{"/:foo/:bar/:test", "/:foo/:bar/(.*)", "/foo/bar/test", "", "/foo/bar/test"},
			{"/:foo/:bar/:test/", "/:foo/:bar/(.*)/", "/foo/bar/test", "", "/foo/bar/test"},
			{"/:foo/:bar/:test/", "/:foo/:bar/(.*)/", "/foo/bar/test", "", "/foo/bar/test/"},
			{"/:foo/:bar/:test", "/:foo/:bar/(.*)", "/foo/bar/test", "", "/foo/bar/test/123"},
			{"/(.*)", "/(.*)/", "/foo/bar/test/123", "", "/foo/bar/test/123"},
			{"/:foo/(.*)", "/:foo/(.*)/", "/foo/bar/test/123", "", "/foo/bar/test/123"},
			{"/:foo/:bar/(.*)", "/:foo/:bar/(.*)/", "/foo/bar/test/123", "", "/foo/bar/test/123"},
			{"/:foo/:bar/(.*)", "/:foo/:bar/(.*)/", "/foo/bar/test/123/", "", "/foo/bar/test/123/"},
			{"/:foo/:bar/(.*)/", "/:foo/:bar/(.*)/", "/foo/bar/test/123/", "", "/foo/bar/test/123/"},
			{"/:foo/:bar/(.*)", "/:foo/:bar/(.*)/", "/foo/bar/test/123", "", "/foo/bar/test/123"},
			{"/:foo/:bar/(.*)/", "/:foo/:bar/(.*)/", "/foo/bar/test/123/", "", "/foo/bar/test/123"},
		}

		for _, tt := range tests {
			t.Run("", func(t *testing.T) {
				router := NewRouter()
				router.Use(tt.middlewareRoute, func(c *Context) {
					assert.Equal(t, tt.expectedMiddlewareBaseUrl, c.Request.BaseUrl)
					c.Next()
				})
				router.GET(tt.route, func(c *Context) {
					assert.Equal(t, tt.expectedBaseUrl, c.Request.BaseUrl)
				})
				server := httptest.NewServer(router)
				defer server.Close()
				_, _, _, err := request("GET", server.URL+tt.path, nil)
				require.NoError(t, err)
			})
		}
	})
}

func TestRequest_Fresh(t *testing.T) {
	t.Run("should return true when the resource is not modified", func(t *testing.T) {
		router, etag := NewRouter(), `"12345"`
		router.Use(func(c *Context) {
			c.Set("ETag", etag)
			c.Send(strconv.FormatBool(c.Request.Fresh()))
		})
		server := httptest.NewServer(router)
		defer server.Close()
		status, _, _, err := request("GET", server.URL+"/", http.Header{util.H("if-none-match"): []string{etag}})
		require.Nil(t, err)
		assert.Equal(t, http.StatusNotModified, status)
	})
}

func TestRequest_Is(t *testing.T) {
	createRequest := func(contentType string) *Request {
		req := httptest.NewRequest("GET", "/", nil)
		if contentType != "" {
			req.Header.Set("content-type", contentType)
		}
		req.Header.Set("transfer-encoding", "chunked")
		return NewRequest(req)
	}

	tests := []struct {
		contentType string
		types       []string
		expected    string
	}{
		{"text/html; charset=utf-8", []string{"html"}, "html"},
		{"text/html; charset=utf-8", []string{"text/html"}, "text/html"},
		{"text/html; charset=utf-8", []string{"text/*"}, "text/html"},
		{"application/json", []string{"json"}, "json"},
		{"application/json", []string{"application/json"}, "application/json"},
		{"application/json", []string{"application/*"}, "application/json"},
		{"application/json", []string{"html"}, ""},
		{"", []string{"html"}, ""},
		{"", []string{"*"}, ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, createRequest(tt.contentType).Is(tt.types...))
	}
}

func TestRequest_Range(t *testing.T) {
	createRequest := func(rangeHeader string) *Request {
		req := httptest.NewRequest("GET", "/", nil)
		if rangeHeader != "" {
			req.Header.Set("range", rangeHeader)
		}
		return NewRequest(req)
	}

	tests := []struct {
		rangeHeader    string
		size           int64
		combine        bool
		expectedRanges util.Ranges
	}{
		{
			rangeHeader: "bytes=0-499",
			size:        1000,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 0, End: 499}},
			},
		},
		{
			rangeHeader: "bytes=0-499",
			size:        200,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 0, End: 199}},
			},
		},
		{
			rangeHeader: "bytes=-400",
			size:        1000,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 600, End: 999}},
			},
		},
		{
			rangeHeader: "bytes=400-",
			size:        1000,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 400, End: 999}},
			},
		},
		{
			rangeHeader: "bytes=0-",
			size:        1000,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 0, End: 999}},
			},
		},
		{
			rangeHeader: "bytes=-1",
			size:        1000,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 999, End: 999}},
			},
		},
		{
			size:        1000,
			rangeHeader: "bytes=40-80,81-90,-1",
			expectedRanges: util.Ranges{
				Type: "bytes",
				Ranges: []*util.Range{
					{Start: 40, End: 80},
					{Start: 81, End: 90},
					{Start: 999, End: 999},
				},
			},
		},
		{
			rangeHeader: "bytes=0-499,1000-,500-999",
			size:        200,
			expectedRanges: util.Ranges{
				Type:   "bytes",
				Ranges: []*util.Range{{Start: 0, End: 199}},
			},
		},
		{
			rangeHeader: "items=0-5",
			size:        1000,
			expectedRanges: util.Ranges{
				Type:   "items",
				Ranges: []*util.Range{{Start: 0, End: 5}},
			},
		},
		{
			size:        150,
			rangeHeader: "bytes=0-4,90-99,5-75,100-199,101-102",
			combine:     true,
			expectedRanges: util.Ranges{
				Type: "bytes",
				Ranges: []*util.Range{
					{Start: 0, End: 75},
					{Start: 90, End: 149},
				},
			},
		},
		{
			size:        150,
			rangeHeader: "bytes=-1,20-100,0-1,101-120",
			combine:     true,
			expectedRanges: util.Ranges{
				Type: "bytes",
				Ranges: []*util.Range{
					{Start: 149, End: 149},
					{Start: 20, End: 120},
					{Start: 0, End: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		ranges, err := createRequest(tt.rangeHeader).Range(tt.size, tt.combine)
		require.NoError(t, err)
		assert.Equal(t, tt.expectedRanges.Type, ranges.Type)
		for i, r := range tt.expectedRanges.Ranges {
			assert.Equal(t, r.Start, ranges.Ranges[i].Start)
			assert.Equal(t, r.End, ranges.Ranges[i].End)
		}
	}
}
