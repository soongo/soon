// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"
)

// Context is the most important part of soon.
// It allows us to pass variables between middleware, manage the flow,
// validate the JSON of a request and render a JSON response for example.
type Context struct {
	*Request
	http.ResponseWriter

	next Next

	// Locals contains response local variables scoped to the request,
	// and therefore available during that request / response cycle (if any).
	//
	// This property is useful for exposing request-level information such as
	// the request path name, authenticated user, user settings, and so on.
	Locals Locals

	// HeadersSent indicates if the app sent HTTP headers for the response.
	HeadersSent bool

	// The finished property will be true if response.end()
	// has been called.
	finished bool
}

func NewContext(req *http.Request, res http.ResponseWriter) *Context {
	c := &Context{Request: NewRequest(req), ResponseWriter: res}
	c.init()
	return c
}

type Next func(v ...interface{})

// Next calls the next handler
func (c *Context) Next(v ...interface{}) {
	c.next(v...)
}

type Locals map[interface{}]interface{}

func (l Locals) Get(k interface{}) interface{} {
	return l[k]
}

func (l Locals) Set(k interface{}, v interface{}) {
	l[k] = v
}

func (c *Context) init() {
	c.Locals = make(Locals, 0)
}

// Appends the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the
// specified value. The value parameter can be a string or a string slice.
// Note: calling c.Set() after c.Append() will reset the previously-set
// header value.
func (c *Context) Append(key string, value interface{}) {
	if s, ok := value.(string); ok {
		c.Header().Add(key, s)
	} else if arr, ok := value.([]string); ok {
		for _, s := range arr {
			c.Header().Add(key, s)
		}
	}
}

// Get gets the first value of response header associated with the given key.
// If there are no values associated with the key, Get returns "".
// It is case insensitive
func (c *Context) Get(field string) string {
	return c.Header().Get(field)
}

// Set sets the response header entries associated with key to the
// single element value. It replaces any existing values
// associated with key. The key is case insensitive;
//
// To set multiple fields at once, pass a string map as the parameter.
func (c *Context) Set(value ...interface{}) {
	util.SetHeader(c, value...)
}

// Status sets the HTTP status for the response.
func (c *Context) Status(code int) {
	c.WriteHeader(code)
}

// SendStatus sets the response HTTP status code to statusCode and
// send its string representation as the response body.
func (c *Context) SendStatus(code int) {
	c.Status(code)
	c.Send(http.StatusText(code))
}

// Sets the Content-Type HTTP header to the MIME type as determined
// by LookupMimeType() for the specified type. If type contains the
// “/” character, then it sets the Content-Type to type.
func (c *Context) Type(s string) {
	util.SetContentType(c, s)
}

// Sets the HTTP response Content-Disposition header field to “attachment”.
// If a filename is given, then it sets the Content-Type based on the extension
// name via c.Type(), and sets the Content-Disposition “filename=” parameter.
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
	http.SetCookie(c, cookie)
}

// ClearCookie clears the specified cookie.
func (c *Context) ClearCookie(cookie *http.Cookie) {
	p := cookie.Path
	if p == "" {
		p = "/"
	}

	http.SetCookie(c, &http.Cookie{
		Name:    cookie.Name,
		Value:   "",
		Path:    p,
		Expires: time.Unix(0, 0),
	})
}

// Transfers the file at the given path. Sets the Content-Type response HTTP
// header field based on the filename’s extension.
// Unless the root option is set in the options object, path must be an
// absolute path to the file.
func (c *Context) SendFile(filePath string, options *renderer.FileOptions) {
	c.Render(renderer.File{FilePath: filePath, Options: options})
}

// Transfers the file at path as an “attachment”. Typically, browsers will
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

// End marks the response is finished, and other send operations after end
// will be ignored.
//
// Use to quickly end the response without any data. If you need to respond
// with data, instead use methods such as c.Send() and c.Json().
func (c *Context) End() {
	c.finished = true
}

// Format responds to the Acceptable formats using an `map`
// of mime-type callbacks.
//
// This method uses `req.accepted`, an array of acceptable types ordered by
// their quality values. When "Accept" is not present the _first_ callback is
// invoked, otherwise the first match is used. When no match is performed the
// server responds with 406 "Not Acceptable".
//
// Content-Type is set for you, however you may alter this within the callback
// using `c.Type()` or `c.Set("Content-Type", ...)`.

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
	sort.Strings(keys)
	var acceptsKeys []string
	if len(keys) > 0 {
		acceptsKeys = c.Accepts(keys...)
	}

	util.Vary(c, []string{"Accept"})

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

// Sends string body
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

// sets the common http header.
func (c *Context) renderHeader() {
	c.HeadersSent = true
	c.Header().Set("Connection", "keep-alive")
	c.Header().Set("X-Powered-By", "Soon")
}

// Render uses the specified renderer to deal with http response body.
func (c *Context) Render(renderer renderer.Renderer) {
	if !c.finished {
		c.renderHeader()
		renderer.RenderHeader(c)

		if err := renderer.Render(c); err != nil {
			c.Status(http.StatusInternalServerError)
			panic(err)
		}
	}
}
