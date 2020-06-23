// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"fmt"
	"net/http"
)

// Redirect contains the http request reference and redirects status code
// and location.
type Redirect struct {
	Code     int
	Location string
}

// RenderHeader (Redirect) doesn't write any header.
func (r Redirect) RenderHeader(w http.ResponseWriter, _ *http.Request) {

}

// Render (Redirect) redirects the http request to new location
// and writes redirect response.
func (r Redirect) Render(w http.ResponseWriter, req *http.Request) error {
	if (r.Code < 300 || r.Code > 308) && r.Code != 201 {
		return fmt.Errorf("cannot redirect with status code %d", r.Code)
	}

	http.Redirect(w, req, r.Location, r.Code)
	return nil
}
