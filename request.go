// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/soongo/negotiator"
	"github.com/soongo/soon/util"
)

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

// Locals contains local variables scoped to the request
type Locals map[string]interface{}

// Get returns the value of specified key in locals
func (l Locals) Get(k string) interface{} {
	return l[k]
}

// Set sets the value of specified key in locals
func (l Locals) Set(k string, v interface{}) {
	l[k] = v
}

// MarshalJSON transforms the Params object to json
func (p Params) MarshalJSON() ([]byte, error) {
	m := make(map[string]string, len(p))
	for k, v := range p {
		m[fmt.Sprint(k)] = v
	}
	return json.Marshal(m)
}

// Request is a wrapper for http.Request.
type Request struct {
	*http.Request

	// Params contains all matched url params of the request
	Params Params

	// Locals contains local variables scoped to the request,
	// and therefore available during that request / response cycle (if any).
	//
	// This property is useful for exposing request-level information such as
	// the request path name, authenticated user, user settings, and so on.
	Locals Locals
}

// NewRequest returns an instance of Request object
func NewRequest(req *http.Request) *Request {
	return &Request{req, make(Params, 0), make(Locals, 0)}
}

// Accepts checks if the specified content types are acceptable, based on the
// request’s Accept HTTP header field. The method returns the best match,
// or if none of the specified content types is acceptable, returns nil (in
// which case, the application should respond with 406 "Not Acceptable").
//
// The types value may be multiple MIME types string (such as “application/json”,
// "text/html"), extension names (such as “json”, "text").
// The method returns the best match (if any).
func (r *Request) Accepts(types ...string) []string {
	n := negotiator.New(r.Header)
	if len(types) == 0 {
		return n.MediaTypes()
	}

	// no accept header, return first given type
	if len(r.Header[negotiator.HeaderAccept]) == 0 {
		return types[0:1]
	}

	mimes := util.StringSlice(types).Map(func(s string) string {
		if strings.Index(s, "/") == -1 {
			return util.LookupMimeType(s)
		}
		return s
	})

	accept := n.MediaType(mimes...)
	if accept != "" {
		i := util.StringSlice(mimes).Index(accept)
		if i > -1 && i < len(types) {
			return types[i : i+1]
		}
	}

	return nil
}

// AcceptsEncodings reports accepted encodings or best fit based on `encodings`.
func (r *Request) AcceptsEncodings(encodings ...string) []string {
	n := negotiator.New(r.Header)
	if len(encodings) == 0 {
		return n.Encodings()
	}

	accepts := n.Encodings(encodings...)
	if len(accepts) > 0 {
		return accepts[0:1]
	}

	return nil
}

// AcceptsCharsets reports accepted charsets or best fit based on `charsets`.
func (r *Request) AcceptsCharsets(charsets ...string) []string {
	n := negotiator.New(r.Header)
	if len(charsets) == 0 {
		return n.Charsets()
	}

	accepts := n.Charsets(charsets...)
	if len(accepts) > 0 {
		return accepts[0:1]
	}

	return nil
}

// AcceptsLanguages reports accepted languages or best fit based on `languages`.
func (r *Request) AcceptsLanguages(languages ...string) []string {
	n := negotiator.New(r.Header)
	if len(languages) == 0 {
		return n.Languages()
	}

	accepts := n.Languages(languages...)
	if len(accepts) > 0 {
		return accepts[0:1]
	}

	return nil
}

// resetParams resets params to empty
func (r *Request) resetParams() {
	if r.Params == nil || len(r.Params) > 0 {
		r.Params = make(Params, 0)
	}
}
