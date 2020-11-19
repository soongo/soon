// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package internal

import (
	"errors"
	"net/http"
)

type HttpError interface {
	error
	Status() int
}

type statusError struct {
	code int
	err  error
}

// NewStatusError creates *statusError by code and error
func NewStatusError(code int, err error) *statusError {
	return &statusError{code, err}
}

// NewStatusError creates *statusError by code
func NewStatusCodeError(code int) *statusError {
	return NewStatusTextError(code, http.StatusText(code))
}

// NewStatusError creates *statusError by code and text
func NewStatusTextError(code int, s string) *statusError {
	return NewStatusError(code, errors.New(s))
}

// Error returns the error text
func (e *statusError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

// Status returns status code
func (e *statusError) Status() int {
	return e.code
}
