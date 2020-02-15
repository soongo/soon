// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/soongo/soon/util"

	"github.com/soongo/soon/renderer"
)

// Response is a custom http.ResponseWriter implementation.
type Response struct {
	http.ResponseWriter

	// The finished property will be true if response.end()
	// has been called.
	finished bool
}

// Appends the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the
// specified value. The value parameter can be a string or a string slice.
// Note: calling r.Set() after r.Append() will reset the previously-set
// header value.
func (r *Response) Append(key string, value interface{}) {
	if s, ok := value.(string); ok {
		r.Header().Add(key, s)
	} else if arr, ok := value.([]string); ok {
		for _, s := range arr {
			r.Header().Add(key, s)
		}
	}
}

// Get returns the HTTP response header specified by field.
// The match is case-insensitive.
func (r *Response) Get(field string) string {
	return r.Header().Get(field)
}

// Sets the response’s HTTP header field to value.
// To set multiple fields at once, pass a string map as the parameter.
func (r *Response) Set(value ...interface{}) {
	util.SetHeader(r, value...)
}

// Status sets the HTTP status for the response.
func (r *Response) Status(code int) {
	r.WriteHeader(code)
}

// SendStatus sets the response HTTP status code to statusCode and
// send its string representation as the response body.
func (r *Response) SendStatus(code int) {
	r.Status(code)
	r.Send(http.StatusText(code))
}

// Sets the Content-Type HTTP header to the MIME type as determined
// by LookupMimeType() for the specified type. If type contains the
// “/” character, then it sets the Content-Type to type.
func (r *Response) Type(s string) {
	util.SetContentType(r, s)
}

// Sets the HTTP response Content-Disposition header field to “attachment”.
// If a filename is given, then it sets the Content-Type based on the extension
// name via res.Type(), and sets the Content-Disposition “filename=” parameter.
func (r *Response) Attachment(filename ...string) {
	contentDisposition := "attachment"
	if len(filename) >= 1 {
		name := filename[0]
		contentDisposition = fmt.Sprintf("attachment; filename=\"%s\"", name)
		r.Type(filepath.Ext(name))
	}
	r.Set("Content-Disposition", contentDisposition)
}

// Cookie sets cookie.
func (r *Response) Cookie(c *http.Cookie) {
	http.SetCookie(r, c)
}

// ClearCookie clears the specified cookie.
func (r *Response) ClearCookie(c *http.Cookie) {
	p := c.Path
	if p == "" {
		p = "/"
	}

	cookie := http.Cookie{
		Name:    c.Name,
		Value:   "",
		Path:    p,
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(r, &cookie)
}

// Transfers the file at the given path. Sets the Content-Type response HTTP
// header field based on the filename’s extension.
// Unless the root option is set in the options object, path must be an
// absolute path to the file.
func (r *Response) SendFile(filePath string, options *renderer.FileOptions) {
	r.Render(renderer.File{FilePath: filePath, Options: options})
}

// Transfers the file at path as an “attachment”. Typically, browsers will
// prompt the user for download. By default, the Content-Disposition header
// “filename=” parameter is path (this typically appears in the browser dialog).
// Override this default with the options.Name parameter.
//
// This method uses res.SendFile() to transfer the file. The optional options
// argument passes through to the underlying res.SendFile() call, and takes the
// exact same parameters.
func (r *Response) Download(filePath string, options *renderer.FileOptions) {
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
	r.SendFile(filePath, options)
}

// End marks the response is finished, and other send operations after end
// will be ignored.
//
// Use to quickly end the response without any data. If you need to respond
// with data, instead use methods such as res.Send() and res.Json().
func (r *Response) End() {
	r.finished = true
}

// Sends string body
func (r *Response) Send(s string) {
	r.String(s)
}

// String sends a plain text response.
func (r *Response) String(s string) {
	r.Render(renderer.String{Data: s})
}

// Json sends a JSON response.
// This method sends a response (with the correct content-type) that is
// the parameter converted to a JSON string.
func (r *Response) Json(v interface{}) {
	r.Render(renderer.JSON{Data: v})
}

// sets the common http header.
func (r *Response) renderHeader() {
	r.Header().Set("Connection", "keep-alive")
	r.Header().Set("X-Powered-By", "Soon")
}

// Render uses the specified renderer to deal with http response body.
func (r *Response) Render(renderer renderer.Renderer) {
	if !r.finished {
		r.renderHeader()
		renderer.RenderHeader(r)

		if err := renderer.Render(r); err != nil {
			r.Status(http.StatusInternalServerError)
			panic(err)
		}
	}
}
