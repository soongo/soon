package soon

import (
	"net/http"
	"strings"
)

type App struct {
	*Router
}

func New() *App {
	return &App{Router: NewRouter()}
}

func (app *App) Use(params ...interface{}) {
	length := len(params)
	if length == 2 {
		if m, ok := params[0].(string); ok {
			if r, ok := params[1].(*Router); ok {
				app.mount(m, r)
				return
			}
			panic("second param should be Router")
		}
		panic("mount point should be string")
	}

	if length == 1 {
		if m, ok := params[0].(func(http.ResponseWriter, *http.Request, func())); ok {
			app.Router.Use(m)
			return
		}

		if r, ok := params[0].(*Router); ok {
			app.mount("/", r)
			return
		}

		panic("params should be middleware function or Router")
	}

	panic("params count should be 1 or 2")
}

func (app *App) mount(mountPoint string, r *Router) {
	for _, v := range r.routes {
		app.routes = append(app.routes, &node{
			method:       v.method,
			route:        routeJoin(mountPoint, v.route),
			isMiddleware: v.isMiddleware,
			handle:       v.handle,
		})
	}
}

func routeJoin(routes ...string) string {
	var route string
	for _, r := range routes {
		if strings.HasSuffix(route, "/") && strings.HasPrefix(r, "/") {
			route = route[:len(route)-1]
		}
		route += r
	}
	return route
}
