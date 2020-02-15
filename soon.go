// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http"
)

// App represents an application with Soon framework.
type App struct {
	*Router
}

// New creates a Soon application.
func New() *App {
	return &App{Router: NewRouter()}
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, app)
func (app *App) Run(addr ...string) error {
	address := resolveAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)
	return http.ListenAndServe(address, app)
}
