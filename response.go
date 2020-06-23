// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"io"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter interface
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher

	// Returns the HTTP response status code of the current request.
	Status() int

	// Returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Writes the string into the response body.
	WriteString(string) (int, error)

	// Returns true if the response header was already written.
	HeaderWritten() bool

	// Returns true if the response body was already written.
	Written() bool

	// Forces to write the http header (status code + headers).
	// Http header changes After this will not be sent with response.
	WriteHeaderNow()
}

type response struct {
	http.ResponseWriter

	status        int
	size          int
	headerWritten bool
}

var _ ResponseWriter = &response{}

func NewResponse(w http.ResponseWriter) *response {
	r := &response{}
	r.reset(w)
	return r
}

func (r *response) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.size = noWritten
	r.status = defaultStatus
}

// HeaderWritten returns true if the response header was already written.
func (c *response) HeaderWritten() bool {
	return c.headerWritten
}

// WriteHeader sends an HTTP response header with the provided status code.
func (r *response) WriteHeader(statusCode int) {
	if statusCode > 0 && r.status != statusCode {
		if r.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to "+
				"override status code %d with %d", r.status, statusCode)
		}
		r.status = statusCode
	}
}

// WriteHeaderNow forces to write the HTTP response header immediately.
// Http header changes After this will not be sent with response.
func (r *response) WriteHeaderNow() {
	if !r.Written() {
		r.size = 0
		r.ResponseWriter.WriteHeader(r.status)
		r.headerWritten = true
	}
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *response) Write(data []byte) (n int, err error) {
	r.WriteHeaderNow()
	n, err = r.ResponseWriter.Write(data)
	r.size += n
	return
}

// WriteString writes the contents of the string s into the HTTP response body.
func (r *response) WriteString(s string) (n int, err error) {
	r.WriteHeaderNow()
	n, err = io.WriteString(r.ResponseWriter, s)
	r.size += n
	return
}

// Flush implements the http.Flush interface. It will sends the HTTP
// response immediately.
func (r *response) Flush() {
	r.WriteHeaderNow()
	r.ResponseWriter.(http.Flusher).Flush()
}

// Status returns the HTTP response status code of the current request.
func (r *response) Status() int {
	return r.status
}

// Size returns the number of bytes already written into the response http body.
func (r *response) Size() int {
	return r.size
}

// Written returns true if the response body was already written.
func (w *response) Written() bool {
	return w.size != noWritten
}
