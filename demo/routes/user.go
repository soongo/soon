package routes

import (
	"io"
	"log"
	"net/http"

	"github.com/soongo/soon"
)

var UserRouter = soon.NewRouter()

func userRouterMiddleware(w http.ResponseWriter, req *http.Request, next func()) {
	log.Println("[user router middleware] start")
	next()
	log.Println("[user router middleware] end")
}

func userIndexHandler(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("start user index route")
	_, _ = io.WriteString(w, "user router")
	log.Println("end user index route")
}

func init() {
	UserRouter.Use(userRouterMiddleware)
	UserRouter.Get("/", userIndexHandler)
}
