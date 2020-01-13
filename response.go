// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"

	"github.com/soongo/soon/renderer"
)

// Response is a custom http.ResponseWriter implementation.
type Response struct {
	http.ResponseWriter
}

// Status sets the HTTP status for the response.
func (r *Response) Status(code int) {
	r.WriteHeader(code)
}

// SendStatus sets the response HTTP status code to statusCode and
// send its string representation as the response body.
func (r *Response) SendStatus(code int) {
	r.Status(code)
	r.Send(http.StatusText(code))
}

// Send sends string body
func (r *Response) Send(s string) {
	r.String(s)
}

// String sends a plain text response.
func (r *Response) String(s string) {
	r.Render(renderer.String{Data: s})
}

// Json sends a JSON response.
// This method sends a response (with the correct content-type) that is
// the parameter converted to a JSON string.
func (r *Response) Json(v interface{}) {
	r.Render(renderer.JSON{Data: v})
}

func (r *Response) RenderHeader() {
	r.Header().Set("Connection", "keep-alive")
	r.Header().Set("X-Powered-By", "Soon")
}

func (r *Response) Render(renderer renderer.Renderer) {
	r.RenderHeader()
	renderer.RenderHeader(r)

	if err := renderer.Render(r); err != nil {
		panic(err)
	}
}
