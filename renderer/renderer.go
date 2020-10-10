// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import "net/http"

// Renderer interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Renderer interface {
	// RenderHeader writes custom headers.
	RenderHeader(http.ResponseWriter, *http.Request)

	// Renderer writes data with custom ContentType.
	Render(http.ResponseWriter, *http.Request) error
}

var (
	_ Renderer = &String{}
	_ Renderer = &JSON{}
	_ Renderer = &File{}
	_ Renderer = &JSONP{}
	_ Renderer = &Redirect{}
)
