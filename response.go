// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"io"
	"net/http"
)

// Response is a custom http.ResponseWriter implementation.
type Response struct {
	http.ResponseWriter
}

// Send sends string body
func (r *Response) Send(s string) error {
	_, err := io.WriteString(r, s)
	return err
}
