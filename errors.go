// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

type httpError interface {
	error
	status() int
}

type statusError struct {
	text string
	code int
}

var _ httpError = &statusError{}

func (e *statusError) Error() string {
	return e.text
}

func (e *statusError) status() int {
	return e.code
}
