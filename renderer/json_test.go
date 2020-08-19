// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := JSON{nil}
	renderer.RenderHeader(w, nil)
	assert.Equal(t, jsonContentType, w.Header().Get("Content-Type"))
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

	assert := assert.New(t)
	for _, tt := range tests {
		w := httptest.NewRecorder()
		renderer := JSON{tt.data}
		err := renderer.Render(w, nil)
		if tt.err != nil {
			assert.NotNil(err)
			assert.Equal("", w.Body.String())
		} else {
			assert.Nil(err)
			assert.Equal(tt.expected+"\n", w.Body.String())
		}
	}
}
