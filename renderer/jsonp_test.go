// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONP_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := JSONP{nil}
	renderer.RenderHeader(w, nil)
	assert.Equal(t, jsonpContentType, w.Header().Get("Content-Type"))
}

func TestJSONP_Render(t *testing.T) {
	tests := []struct {
		data     interface{}
		request  *http.Request
		expected string
		err      error
	}{
		{
			nil,
			nil,
			"/**/ typeof _jsonp_callback_ === 'function' && _jsonp_callback_(null);",
			nil,
		},
		{
			struct {
				id   uint32
				Name string
			}{1, "x"},
			nil,
			`/**/ typeof _jsonp_callback_ === 'function' && _jsonp_callback_({"Name":"x"});`,
			nil,
		},
		{
			struct {
				ID   uint32 `json:"id"`
				Name string `json:"name"`
			}{1, "x"},
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			`/**/ typeof jsonp12345 === 'function' && jsonp12345({"id":1,"name":"x"});`,
			nil,
		},
		{
			[]string{"foo", "bar"},
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			`/**/ typeof jsonp12345 === 'function' && jsonp12345(["foo","bar"]);`,
			nil,
		},
		{
			[]struct {
				ID   uint32 `json:"id"`
				Name string `json:"name"`
			}{{1, "x"}, {2, "y"}},
			httptest.NewRequest("GET", "http://a.com?callback=jsonp12345", nil),
			`/**/ typeof jsonp12345 === 'function' && jsonp12345([{"id":1,"name":"x"},{"id":2,"name":"y"}]);`,
			nil,
		},
		{func() {}, nil, "null", errors.New("")},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		renderer := JSONP{tt.data}
		err := renderer.Render(w, tt.request)
		if tt.err != nil {
			assert.NotNil(err)
			assert.Equal("", w.Body.String())
		} else {
			assert.Nil(err)
			assert.Equal(tt.expected, w.Body.String())
		}
	}
}
