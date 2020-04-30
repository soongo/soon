// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"
)

// Context is the most important part of soon. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
	*Request
	http.ResponseWriter

	next Next

	// The finished property will be true if response.end()
	// has been called.
	finished bool
}

type Next func(v ...interface{})

// Next calls the next handler
func (c *Context) Next() {
	c.next()
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

// Get returns the HTTP response header specified by field.
// The match is case-insensitive.
func (c *Context) Get(field string) string {
	return c.Header().Get(field)
}

// Sets the response’s HTTP header field to value.
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
