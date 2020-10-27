// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/soongo/soon/binding"

	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"
)

// BodyBytesKey indicates a default body bytes key.
const BodyBytesKey = "_soongo/soon/bodybyteskey"

// Context is the most important part of soon.
// It allows us to pass variables between middleware, manage the flow,
// validate the JSON of a request and render a JSON response for example.
type Context struct {
	Request  *Request
	response *response
	Writer   ResponseWriter

	next func(v ...interface{})

	// Locals contains local variables scoped to the request,
	// and therefore available during that request / response cycle (if any).
	//
	// This property is useful for exposing request-level information such as
	// the request path name, authenticated user, user settings, and so on.
	Locals map[string]interface{}

	// This mutex protect locals map
	mu sync.RWMutex

	// The finished property will be true if `context.End()` has been called.
	finished bool
}

// NewContext returns an instance of Context object
func NewContext(r *http.Request, w http.ResponseWriter) *Context {
	c := &Context{Request: NewRequest(r), response: newResponse(w)}
	c.Writer = c.response
	c.Request.writer = c.response
	return c
}

// Next calls the next handler
func (c *Context) Next(v ...interface{}) {
	c.next(v...)
}

// SetLocal is used to store a new key/value pair in locals for this context.
// It also lazy initializes c.Locals if it was not used previously.
func (c *Context) SetLocal(k string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Locals == nil {
		c.Locals = make(map[string]interface{})
	}
	c.Locals[k] = v
}

// SetLocals is used to store multi new key/value pairs in locals for this context.
// It also lazy initializes c.locals if it was not used previously.
func (c *Context) SetLocals(m map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Locals == nil {
		c.Locals = make(map[string]interface{})
	}
	for k, v := range m {
		c.Locals[k] = v
	}
}

// GetLocal returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) GetLocal(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.Locals[key]
	return
}

// MustGetLocal returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGetLocal(key string) interface{} {
	if value, exists := c.GetLocal(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist in locals map")
}

// ResetLocals is used to reset and store new key/value pairs in locals for this context.
func (c *Context) ResetLocals(m map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Locals = make(map[string]interface{})
	for k, v := range m {
		c.Locals[k] = v
	}
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

func (c *Context) del(field string) {
	c.Writer.Header().Del(field)
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
	c.Render(&renderer.File{FilePath: filePath, Options: options})
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
		c.Next(&statusError{status, errors.New(http.StatusText(status))})
	}
}

// BindJSON is a shortcut for c.BindWith(obj, binding.JSON).
func (c *Context) BindJSON(obj interface{}) error {
	return c.BindWith(obj, binding.JSON)
}

// BindQuery is a shortcut for c.BindWith(obj, binding.Query).
func (c *Context) BindQuery(obj interface{}) error {
	return c.BindWith(obj, binding.Query)
}

// BindHeader is a shortcut for c.BindWith(obj, binding.Header).
func (c *Context) BindHeader(obj interface{}) error {
	return c.BindWith(obj, binding.Header)
}

// BindUri binds the passed struct pointer using binding.Uri.
func (c *Context) BindUri(obj interface{}) error {
	m := make(map[string][]string)
	for k, v := range c.Request.Params {
		if s, ok := k.(string); ok {
			m[s] = []string{v}
		} else if i, ok := k.(int); ok {
			m[strconv.Itoa(i)] = []string{v}
		}
	}
	return binding.Uri.BindUri(m, obj)
}

// BindWith binds the passed struct pointer using the specified binding engine.
// See the binding package.
func (c *Context) BindWith(obj interface{}, b binding.Binding) error {
	return b.Bind(c.Request.Request, obj)
}

// MustBindJSON is a shortcut for c.MustBindWith(obj, binding.JSON).
func (c *Context) MustBindJSON(obj interface{}) {
	c.MustBindWith(obj, binding.JSON)
}

// MustBindQuery is a shortcut for c.MustBindWith(obj, binding.Query).
func (c *Context) MustBindQuery(obj interface{}) {
	c.MustBindWith(obj, binding.Query)
}

// MustBindHeader is a shortcut for c.MustBindWith(obj, binding.Header).
func (c *Context) MustBindHeader(obj interface{}) {
	c.MustBindWith(obj, binding.Header)
}

// MustBindUri binds the passed struct pointer using binding.Uri.
// It will panic with HTTP 400 if any error occurs.
// See the binding package.
func (c *Context) MustBindUri(obj interface{}) {
	if err := c.BindUri(obj); err != nil {
		panic(&statusError{http.StatusBadRequest, err})
	}
}

// MustBindWith binds the passed struct pointer using the specified binding engine.
// It will panic with HTTP 400 if any error occurs.
// See the binding package.
func (c *Context) MustBindWith(obj interface{}, b binding.Binding) {
	if err := c.BindWith(obj, b); err != nil {
		panic(&statusError{http.StatusBadRequest, err})
	}
}

// BindBodyWith is similar with BindWith, but it stores the request
// body into the context, and reuse when it is called again.
//
// NOTE: This method reads the body before binding. So you should use
// BindWith for better performance if you need to call only once.
func (c *Context) BindBodyWith(obj interface{}, bb binding.BindingBody) (err error) {
	var body []byte
	if cb, ok := c.GetLocal(BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			return err
		}
		c.SetLocal(BodyBytesKey, body)
	}
	return bb.BindBody(body, obj)
}

// Send is alias for String method
func (c *Context) Send(s string) {
	c.String(s)
}

// String sends a plain text response.
func (c *Context) String(s string) {
	c.Render(&renderer.String{Data: s})
}

// Json sends a JSON response.
// This method sends a response (with the correct content-type) that is
// the parameter converted to a JSON string.
func (c *Context) Json(v interface{}) {
	c.Render(&renderer.JSON{Data: v})
}

// Jsonp sends a JSON response with JSONP support. This method is identical
// to c.Json(), except that it opts-in to JSONP callback support.
func (c *Context) Jsonp(v interface{}) {
	c.Render(&renderer.JSONP{Data: v})
}

// Redirect to the given `location` with `status`.
func (c *Context) Redirect(status int, location string) {
	c.Render(&renderer.Redirect{Code: status, Location: location})
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

		if c.Request.Fresh() {
			c.Status(304)
		}

		status := c.Writer.Status()

		// strip irrelevant headers
		if status == 204 || status == 304 {
			c.del("Content-Type")
			c.del("Content-Length")
			c.del("Transfer-Encoding")
		}

		if c.Request.Method == http.MethodHead || !bodyAllowedForStatus(status) {
			c.Writer.WriteHeaderNow()
			return
		}

		if err := r.Render(c.Writer, c.Request.Request); err != nil {
			panic(err)
		}

		c.finished = true
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
