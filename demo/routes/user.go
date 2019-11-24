package routes

import (
	"log"
	"net/http"

	"github.com/soongo/soon"
)

var UserRouter = &soon.Router{Sensitive: true, Strict: true}

func userRouterMiddleware(res *soon.Response, req *http.Request, next func()) {
	log.Println("[user router middleware] start")
	next()
	log.Println("[user router middleware] end")
}

func userIndexHandler(res *soon.Response, req *http.Request, next func()) {
	log.Println("start user index route")
	_ = res.Send("user router")
	log.Println("end user index route")
}

func init() {
	UserRouter.Use(userRouterMiddleware)
	UserRouter.GET("/Hello/", userIndexHandler)
}
