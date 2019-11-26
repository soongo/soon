// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import "net/http"

// Params contains all matched url params
type Params map[interface{}]string

// Get one param by key
func (p Params) Get(k interface{}) string {
	return p[k]
}

// Set one param with key
func (p Params) Set(k interface{}, v string) {
	p[k] = v
}

// Request is a wrapper for http.Request.
type Request struct {
	*http.Request

	// Params contains all matched url params of the request
	Params Params
}

// resetParams resets params to empty
func (r *Request) resetParams() {
	if r.Params == nil || len(r.Params) > 0 {
		r.Params = make(Params, 0)
	}
}
