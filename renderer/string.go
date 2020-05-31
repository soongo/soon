// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"io"
	"net/http"
)

// String contains the given string.
type String struct {
	Data string
}

var plainContentType = "text/plain; charset=utf-8"

// RenderHeader writes custom headers.
func (s String) RenderHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", plainContentType)
}

// Renderer writes data with custom ContentType.
func (s String) Render(w http.ResponseWriter) error {
	_, err := io.WriteString(w, s.Data)
	return err
}
