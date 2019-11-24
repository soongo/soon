// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import "net/http"

type App struct {
	*Router
}

func New() *App {
	return &App{Router: NewRouter()}
}

func (app *App) setPanicHandler(h func(*Response, *http.Request, interface{})) {
	app.panicHandler = h
}
