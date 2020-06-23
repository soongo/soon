// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"encoding/json"
	"net/http"
)

// JSON contains the given interface object.
type JSON struct {
	Data interface{}
}

const jsonContentType = "application/json; charset=UTF-8"

// RenderHeader writes custom headers.
func (j JSON) RenderHeader(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", jsonContentType)
}

// Render writes data with custom ContentType.
func (j JSON) Render(w http.ResponseWriter, _ *http.Request) error {
	return json.NewEncoder(w).Encode(j.Data)
}
