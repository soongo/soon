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

func TestRedirect_RenderHeader(t *testing.T) {
	emptyPathReq := httptest.NewRequest("GET", "/", nil)
	emptyPathReq.URL.Path = ""

	tests := []struct {
		code                int
		location            string
		req                 *http.Request
		contentType         string
		expectedLocation    string
		expectedContentType string
	}{
		{
			301,
			"/new/location",
			httptest.NewRequest("GET", "/", nil),
			"",
			"/new/location",
			"text/html; charset=utf-8",
		},
		{
			302,
			"/new/location/",
			httptest.NewRequest("HEAD", "/", nil),
			"",
			"/new/location/",
			"text/html; charset=utf-8",
		},
		{
			301,
			"/new/地址",
			httptest.NewRequest("GET", "/", nil),
			"application/json",
			"/new/" + hexEscapeNonASCII("地址"),
			"application/json",
		},
		{
			302,
			"/new/location/",
			httptest.NewRequest("HEAD", "/", nil),
			"application/json",
			"/new/location/",
			"application/json",
		},
		{
			200,
			"/new/location?id=1",
			httptest.NewRequest("POST", "/", nil),
			"",
			"/new/location?id=1",
			"",
		},
		{
			200,
			"/new/location?id=1",
			httptest.NewRequest("POST", "/", nil),
			"application/json",
			"/new/location?id=1",
			"application/json",
		},
		{201, "", emptyPathReq, "", "/", "text/html; charset=utf-8"},
		{201, "", emptyPathReq, "application/json", "/", "application/json"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		if tt.contentType != "" {
			w.Header().Set("content-type", tt.contentType)
		}
		r := &Redirect{Code: tt.code, Location: tt.location}
		r.RenderHeader(w, tt.req)
		assert.Equal(t, tt.code, w.Code)
		assert.Equal(t, tt.expectedLocation, w.Header().Get("location"))
		assert.Equal(t, tt.expectedContentType, w.Header().Get("content-type"))
	}
}

func TestRedirect_Render(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	tests := []struct {
		code     int
		location string
		req      *http.Request
		err      error
	}{
		{301, "/new/location", req, nil},
		{302, "/new/location", httptest.NewRequest("POST", "/", nil), nil},
		{200, "/new/location", req, errors.New("")},
		{201, "/new/location", req, nil},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r := &Redirect{Code: tt.code, Location: tt.location}
		err := r.Render(w, tt.req)
		if tt.err != nil {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
		}
	}
}
