// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import "net/http"

// Renderer interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Renderer interface {
	// RenderHeader writes custom headers.
	RenderHeader(w http.ResponseWriter)

	// Renderer writes data with custom ContentType.
	Render(http.ResponseWriter) error
}

var (
	_ Renderer = JSON{}
	_ Renderer = String{}
)
