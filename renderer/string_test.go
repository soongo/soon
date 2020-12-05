// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	timeFormat = http.TimeFormat
)

func TestString_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := String{"hi"}
	renderer.RenderHeader(w, nil)
	assert.Equal(t, plainContentType, w.Header().Get("Content-Type"))

	w.Header().Set("Content-Type", jsonContentType)
	renderer.RenderHeader(w, nil)
	assert.Equal(t, jsonContentType, w.Header().Get("Content-Type"))
}

func TestString_Render(t *testing.T) {
	tests := []struct {
		s string
	}{
		{""},
		{"hi"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		renderer := String{tt.s}
		renderer.Render(w, nil)
		assert.Equal(t, tt.s, w.Body.String())
	}
}
