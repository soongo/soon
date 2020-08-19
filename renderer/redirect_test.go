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

func TestRedirect_Render(t *testing.T) {
	req := httptest.NewRequest("GET", "/test-redirect", nil)
	tests := []struct {
		code     int
		location string
		req      *http.Request
		err      error
	}{
		{301, "/new/location", req, nil},
		{302, "/new/location", req, nil},
		{200, "/new/location", req, errors.New("")},
		{201, "/new/location", req, nil},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		r := Redirect{tt.code, tt.location}
		err := r.Render(w, tt.req)
		if tt.err != nil {
			assert.NotNil(err)
		} else {
			assert.Nil(err)
		}

		// just for improving test coverage
		r.RenderHeader(w, tt.req)
	}
}
