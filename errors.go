// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

type httpError interface {
	error
	status() int
}

type statusError struct {
	code int
	err  error
}

var _ httpError = &statusError{}

func (e *statusError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e *statusError) status() int {
	return e.code
}
