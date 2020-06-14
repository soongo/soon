// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const (
	jsonpDefaultCallback = "_jsonp_callback_"
	jsonpContentType     = "text/javascript; charset=UTF-8"
)

// JSON contains the given interface object.
type JSONP struct {
	Data interface{}
}

// RenderHeader writes custom headers.
func (j JSONP) RenderHeader(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", jsonpContentType)
}

// Renderer writes data with custom ContentType.
func (j JSONP) Render(w http.ResponseWriter, req *http.Request) error {
	bs, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}

	body, callback := string(bs), jsonpDefaultCallback
	if req != nil {
		if values, ok := req.URL.Query()["callback"]; ok {
			if len(values) > 0 && strings.Trim(values[0], " ") != "" {
				callback = strings.Trim(values[0], " ")
			}
		}
	}

	// the /**/ is a specific security mitigation for "Rosetta Flash JSONP abuse"
	// the typeof check is just to reduce client error noise
	body = "/**/ typeof " + callback + " === 'function' && " + callback + "(" + body + ");"

	_, err = io.WriteString(w, body)
	return err
}
