// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"bytes"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestJSON_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := JSON{nil}
	renderer.RenderHeader(w, nil)
	if got := w.Header().Get("Content-Type"); got != jsonContentType {
		t.Errorf(testErrorFormat, got, jsonContentType)
	}
}

func TestJSON_Render(t *testing.T) {
	tests := []struct {
		data     interface{}
		expected string
		err      error
	}{
		{nil, "null", nil},
		{
			struct {
				Name      string `json:"name"`
				PageTotal uint16 `json:"pageTotal"`
			}{"foo", 50},
			`{"name":"foo","pageTotal":50}`,
			nil,
		},
		{[]string{"foo", "bar"}, `["foo","bar"]`, nil},
		{
			[]struct {
				Name      string `json:"name"`
				PageTotal uint16 `json:"pageTotal"`
			}{{"foo", 50}, {"bar", 20}},
			`[{"name":"foo","pageTotal":50},{"name":"bar","pageTotal":20}]`,
			nil,
		},
		{func() {}, "null", errors.New("")},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		renderer := JSON{tt.data}
		err := renderer.Render(w, nil)
		if tt.err != nil {
			if err == nil {
				t.Errorf(testErrorFormat, err, "none nil error")
			}
			if got := w.Body.String(); got != "" {
				t.Errorf(testErrorFormat, got, "")
			}
		} else {
			if err != nil {
				t.Errorf(testErrorFormat, err, "nil")
			}
			buf := bytes.NewBuffer([]byte(tt.expected))
			buf.WriteByte('\n')
			expected := buf.String()
			if got := w.Body.String(); got != expected {
				t.Errorf(testErrorFormat, got, expected)
			}
		}
	}
}
