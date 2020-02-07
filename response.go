// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/soongo/soon/renderer"
)

// Response is a custom http.ResponseWriter implementation.
type Response struct {
	http.ResponseWriter

	// The finished property will be true if response.end()
	// has been called.
	finished bool
}

var charsetRegexp = regexp.MustCompile(";\\s*charset\\s*=")

type DotfilesPolicy uint8

const (
	DotfilesPolicyIgnore DotfilesPolicy = iota
	DotfilesPolicyAllow
	DotfilesPolicyDeny
)

var (
	ErrIsDir     = errors.New("file is directory")
	ErrForbidden = errors.New(http.StatusText(http.StatusForbidden))
	ErrNotFound  = errors.New(http.StatusText(http.StatusNotFound))
)

type Callback func(err error)

type FileOptions struct {
	// Sets the max-age property of the Cache-Control header.
	MaxAge *time.Duration

	// Root directory for relative filenames.
	Root string

	// filename for download
	Name string

	// Whether sets the Last-Modified header to the last modified date of the
	// file on the OS. Set true to disable it.
	LastModifiedDisabled bool

	// HTTP headers to serve with the file.
	Header map[string]string

	// Option for serving dotfiles. Possible values are “DotfilesPolicyAllow”,
	// “DotfilesPolicyDeny”, “DotfilesPolicyIgnore”.
	DotfilesPolicy DotfilesPolicy

	// Enable or disable accepting ranged requests. Set true to disable it.
	AcceptRangesDisabled bool
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
	if len(value) == 2 {
		if k, ok := value[0].(string); ok {
			if v, ok := value[1].(string); ok {
				if strings.ToLower(k) == "content-type" && !charsetRegexp.MatchString(v) {
					charset := LookupCharset(strings.Split(v, ";")[0])
					if charset != "" {
						v += "; charset=" + charset
					}
				}
				r.Header().Set(k, v)
			}
		}
		return
	}

	if len(value) == 1 {
		if arr, ok := value[0].(map[string]string); ok {
			for k, v := range arr {
				r.Set(k, v)
			}
		}
	}
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
	k, s := "Content-Type", strings.Trim(s, " ")
	if strings.Contains(s, "/") {
		r.Set(k, s)
	} else {
		r.Set(k, LookupMimeType(s))
	}
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
func (r *Response) SendFile(p string, o *FileOptions, callback Callback) {
	p = strings.Trim(p, " ")
	if p == "" {
		panic("path argument is required")
	}

	if o == nil {
		o = &FileOptions{}
	}

	root := strings.Trim(o.Root, " ")
	if root == "" && !filepath.IsAbs(p) {
		panic("path must be absolute or specify root")
	}

	absPath := EncodeURI(filepath.Join(root, p))
	fileInfo, err := os.Stat(absPath)

	if callback == nil {
		callback = func(err error) {
			if err == ErrIsDir {
				r.Send(http.StatusText(http.StatusNotFound))
			}
		}
	}

	if err != nil {
		callback(err)
		return
	}

	if fileInfo.IsDir() {
		callback(ErrIsDir)
		return
	}

	if strings.HasPrefix(filepath.Base(absPath), ".") {
		if o.DotfilesPolicy == DotfilesPolicyIgnore {
			callback(ErrNotFound)
			return
		}
		if o.DotfilesPolicy == DotfilesPolicyDeny {
			callback(ErrForbidden)
			return
		}
	}

	if o.Header != nil {
		r.Set(o.Header)
	}

	if o.MaxAge != nil {
		r.Set("Cache-Control", fmt.Sprintf("max-age=%.0f", (*o.MaxAge).Seconds()))
	}

	if !o.LastModifiedDisabled {
		r.Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))
	}

	f, err := os.Open(absPath)
	if err != nil {
		callback(err)
		return
	}
	defer f.Close()

	r.Type(filepath.Ext(p))
	r.renderHeader()
	_, err = io.Copy(r, f)
	callback(err)
}

// Transfers the file at path as an “attachment”. Typically, browsers will
// prompt the user for download. By default, the Content-Disposition header
// “filename=” parameter is path (this typically appears in the browser dialog).
// Override this default with the filename parameter.
//
// When an error occurs or transfer is complete, the method calls the optional
// callback function fn. This method uses res.SendFile() to transfer the file.
//
// The optional options argument passes through to the underlying res.SendFile()
// call, and takes the exact same parameters.
func (r *Response) Download(p string, o *FileOptions, c Callback) {
	name := filepath.Base(p)
	if o == nil {
		o = &FileOptions{}
	}
	if o.Name != "" {
		name = o.Name
	}
	if o.Header == nil {
		o.Header = make(map[string]string)
	}
	o.Header["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\"", name)
	r.SendFile(p, o, c)
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
			panic(err)
		}
	}
}
