// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testErrorFormat = "got `%v`, expect `%v`"
	timeFormat      = http.TimeFormat
)

func TestString_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := String{"hi"}
	renderer.RenderHeader(w, nil)
	if got := w.Header().Get("Content-Type"); got != plainContentType {
		t.Errorf(testErrorFormat, got, plainContentType)
	}
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
		if got := w.Body.String(); got != tt.s {
			t.Errorf(testErrorFormat, got, tt.s)
		}
	}
}
