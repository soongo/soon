// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"
)

// Context is the most important part of soon.
// It allows us to pass variables between middleware, manage the flow,
// validate the JSON of a request and render a JSON response for example.
type Context struct {
	Request  *Request
	response *response
	Writer   ResponseWriter

	next func(v ...interface{})

	// The finished property will be true if `context.End()` has been called.
	finished bool
}

// NewContext returns an instance of Context object
func NewContext(r *http.Request, w http.ResponseWriter) *Context {
	c := &Context{Request: NewRequest(r), response: newResponse(w)}
	c.Writer = c.response
	return c
}

// Next calls the next handler
func (c *Context) Next(v ...interface{}) {
	c.next(v...)
}

// Params is a shortcut method of request's params
func (c *Context) Params() Params {
	return c.Request.Params
}

// Locals is a shortcut method of request's locals
func (c *Context) Locals() Locals {
	return c.Request.Locals
}

// Query is a shortcut method for getting query of request
func (c *Context) Query() url.Values {
	return c.Request.URL.Query()
}

// HeadersSent indicates if the response header was already sent.
func (c *Context) HeadersSent() bool {
	return c.Writer.HeaderWritten()
}

// Append the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the
// specified value. The value parameter can be a string or a string slice.
// Note: calling c.Set() after c.Append() will reset the previously-set
// header value.
func (c *Context) Append(key string, value interface{}) {
	util.AddHeader(c.Writer, key, value)
}

// Get the first value of response header associated with the given key.
// If there are no values associated with the key, Get returns "".
// It is case insensitive
func (c *Context) Get(field string) string {
	return c.Writer.Header().Get(field)
}

// Set the response header entries associated with key to the
// single element value. It replaces any existing values
// associated with key. The key is case insensitive;
//
// To set multiple fields at once, pass a string map as the parameter.
func (c *Context) Set(value ...interface{}) {
	util.SetHeader(c.Writer, value...)
}

// Vary adds `field` to Vary. If already present in the Vary set, then
// this call is simply ignored.
func (c *Context) Vary(fields ...string) {
	util.Vary(c.Writer, fields)
}

// Status sets the HTTP status for the response.
func (c *Context) Status(code int) {
	c.Writer.WriteHeader(code)
}

// SendStatus sets the response HTTP status code to statusCode and
// send its string representation as the response body.
func (c *Context) SendStatus(code int) {
	c.Status(code)
	c.Send(http.StatusText(code))
}

// Type sets the Content-Type HTTP header to the MIME type as determined
// by LookupMimeType() for the specified type. If type contains the
// “/” character, then it sets the Content-Type to type.
func (c *Context) Type(s string) {
	util.SetContentType(c.Writer, s)
}

// Links sets Link header field with the given `links`.
func (c *Context) Links(links map[string]string) {
	link := strings.Trim(c.Get("Link"), " ")
	if link != "" {
		link += ", "
	}

	i, length := 0, len(links)
	for k, v := range links {
		link += "<" + v + ">; rel=\"" + k + "\""
		i++
		if i < length {
			link += ", "
		}
	}

	c.Set("Link", link)
}

// Location sets the location header to `url`.
// The given `url` can also be "back", which redirects
// to the _Referrer_ or _Referer_ headers or "/".
func (c *Context) Location(url string) {
	url = strings.Trim(url, " ")
	if url == "back" {
		url = c.Get("Referrer")
		if url == "" {
			url = "/"
		}
	}
	c.Set("Location", util.EncodeURI(url))
}

// Attachment sets the HTTP response Content-Disposition header field to
// “attachment”. If a filename is given, then it sets the Content-Type based
// on the extension name via c.Type(), and sets the Content-Disposition
// “filename=” parameter.
func (c *Context) Attachment(filename ...string) {
	contentDisposition := "attachment"
	if len(filename) >= 1 {
		name := filename[0]
		contentDisposition = fmt.Sprintf("attachment; filename=\"%s\"", name)
		c.Type(filepath.Ext(name))
	}
	c.Set("Content-Disposition", contentDisposition)
}

// Cookie sets cookie.
func (c *Context) Cookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// ClearCookie clears the specified cookie.
func (c *Context) ClearCookie(cookie *http.Cookie) {
	p := cookie.Path
	if p == "" {
		p = "/"
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:    cookie.Name,
		Value:   "",
		Path:    p,
		Expires: time.Unix(0, 0),
	})
}

// SendFile transfers the file at the given path. Sets the Content-Type
// response HTTP header field based on the filename’s extension.
// Unless the root option is set in the options object, path must be an
// absolute path to the file.
func (c *Context) SendFile(filePath string, options *renderer.FileOptions) {
	c.Render(renderer.File{FilePath: filePath, Options: options})
}

// Download transfers the file at path as an “attachment”. Typically, browsers will
// prompt the user for download. By default, the Content-Disposition header
// “filename=” parameter is path (this typically appears in the browser dialog).
// Override this default with the options.Name parameter.
//
// This method uses c.SendFile() to transfer the file. The optional options
// argument passes through to the underlying c.SendFile() call, and takes the
// exact same parameters.
func (c *Context) Download(filePath string, options *renderer.FileOptions) {
	name := filepath.Base(filePath)
	if options == nil {
		options = &renderer.FileOptions{}
	}
	if options.Name != "" {
		name = options.Name
	}
	if options.Header == nil {
		options.Header = make(map[string]string, 1)
	}
	options.Header["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\"", name)
	c.SendFile(filePath, options)
}

// End signals to the server that all of the response headers and body have been
// sent; other operations after it will be ignored.
//
// Use to quickly end the response without any data. If you need to respond
// with data, instead use methods such as c.Send() and c.Json().
func (c *Context) End() {
	c.finished = true
	c.Writer.Flush()
}

// Format responds to the Acceptable formats using an `map`
// of mime-type callbacks.
//
// This method uses `req.accepted`, an array of acceptable types ordered by
// their quality values. When "Accept" is not present the first after sorted
// callback is invoked, otherwise the first match is used. When no match is
// performed the server responds with 406 "Not Acceptable".
//
// Content-Type is set for you, however you may alter this within the callback
// using `c.Type()` or `c.Set("Content-Type", ...)`.
//
// By default Soon passes an `error` with a `.status` of 406 to `Next(err)`
// if a match is not made. If you provide a `.default` callback it will be
// invoked instead.
func (c *Context) Format(m map[string]Handle) {
	defaultHandler := m["default"]
	if defaultHandler != nil {
		handles := make(map[string]Handle, len(m))
		for k, v := range m {
			handles[k] = v
		}
		m = handles
	}
	delete(m, "default")
	keys, i := make([]string, len(m), len(m)), 0
	for k := range m {
		keys[i] = k
		i++
	}

	// TODO: As keys of map is not at determined order, so the first callback
	// TODO: is the first one after sorted, not the one first declared.
	sort.Strings(keys)

	var acceptsKeys []string
	if len(keys) > 0 {
		acceptsKeys = c.Request.Accepts(keys...)
	}

	c.Vary("Accept")

	if len(acceptsKeys) > 0 {
		key := acceptsKeys[0]
		c.Set("Content-Type", util.NormalizeType(key).Value)
		m[key](c)
	} else if defaultHandler != nil {
		defaultHandler(c)
	} else {
		status := http.StatusNotAcceptable
		c.Next(&statusError{http.StatusText(status), status})
	}
}

// Send is alias for String method
func (c *Context) Send(s string) {
	c.String(s)
}

// String sends a plain text response.
func (c *Context) String(s string) {
	c.Render(renderer.String{Data: s})
}

// Json sends a JSON response.
// This method sends a response (with the correct content-type) that is
// the parameter converted to a JSON string.
func (c *Context) Json(v interface{}) {
	c.Render(renderer.JSON{Data: v})
}

// Jsonp sends a JSON response with JSONP support. This method is identical
// to c.Json(), except that it opts-in to JSONP callback support.
func (c *Context) Jsonp(v interface{}) {
	c.Render(renderer.JSONP{Data: v})
}

// Redirect to the given `location` with `status`.
func (c *Context) Redirect(status int, location string) {
	c.Render(renderer.Redirect{Code: status, Location: location})
}

// sets the common http header.
func (c *Context) renderHeader() {
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Powered-By", "Soon")
}

// Render uses the specified renderer to deal with http response body.
func (c *Context) Render(r renderer.Renderer) {
	if !c.finished {
		c.renderHeader()
		r.RenderHeader(c.Writer, c.Request.Request)

		if !bodyAllowedForStatus(c.Writer.Status()) {
			c.Writer.WriteHeaderNow()
			return
		}

		if err := r.Render(c.Writer, c.Request.Request); err != nil {
			panic(err)
		}
	}
}

// bodyAllowedForStatus reports whether a given response status code
// permits a body. See RFC 7230, section 3.3.
//
// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
