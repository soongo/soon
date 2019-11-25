// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import "net/http"

// App represents an application with Soon framework.
type App struct {
	*Router
}

// New creates a Soon application.
func New() *App {
	return &App{Router: NewRouter()}
}

// SetPanicHandler sets the panic handler to handle panics recovered from
// http handlers.
func (app *App) SetPanicHandler(h func(*http.Request, *Response, interface{})) {
	app.panicHandler = h
}
