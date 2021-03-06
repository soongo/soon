// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	// BaseUrl is the URL path on which a router instance was mounted.
	//
	// Even if you use a path pattern to load the router,
	// the baseUrl property returns the matched string, not the pattern(s).
	BaseUrl string

	Hostname string
	Protocol string
	Path     string
	Query    url.Values
	Secure   bool
	Xhr      bool

	writer ResponseWriter
}

// NewRequest returns an instance of Request object
func NewRequest(req *http.Request) *Request {
	r := &Request{
		Request:  req,
		Params:   make(Params, 0),
		Hostname: req.URL.Host,
		Protocol: req.URL.Scheme,
		Path:     req.URL.EscapedPath(),
		Query:    req.URL.Query(),
		Secure:   req.URL.Scheme == "https",
	}

	r.Xhr = strings.ToLower(r.Get("X-Requested-With")) == "xmlhttprequest"

	return r
}

// Get returns the specified HTTP request header field (case-insensitive match).
func (r *Request) Get(key string) string {
	return r.Header.Get(key)
}

// ContentType returns the Content-Type HTTP header of request
func (r *Request) ContentType() string {
	contentType := strings.TrimSpace(r.Get("Content-Type"))
	for i, char := range contentType {
		if char == ' ' || char == ';' {
			return contentType[:i]
		}
	}
	return contentType
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

// Fresh checks if the request is fresh, aka Last-Modified and/or the ETag still match.
func (r *Request) Fresh() bool {
	// GET or HEAD for weak freshness validation only
	method := r.Method
	if http.MethodGet != method && http.MethodHead != method {
		return false
	}

	status := r.writer.Status()

	// 2xx or 304 as per rfc2616 14.26
	if (status >= 200 && status < 300) || 304 == status {
		return util.Fresh(r.Header, r.writer.Header())
	}

	return false
}

// Is checks if the incoming request contains the "Content-Type"
// header field, and it contains the give mime `type`.
func (r *Request) Is(types ...string) string {
	return util.RequestTypeIs(r.Request, types...)
}

// Range parses `Range` header field, capping to the given `size`.
//
// The "combine" argument can be set to `true` and overlapping & adjacent ranges
// * will be combined into a single range.
func (r *Request) Range(size int64, combine bool) (util.Ranges, error) {
	return util.RangeParser(size, r.Get("Range"), combine)
}

// resetParams resets params to empty
func (r *Request) resetParams() {
	if r.Params == nil || len(r.Params) > 0 {
		r.Params = make(Params, 0)
	}
}
