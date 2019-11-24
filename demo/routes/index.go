package routes

import (
	"log"
	"net/http"

	"github.com/soongo/soon"
)

var IndexRouter = soon.NewRouter()

func indexHandler(res *soon.Response, req *http.Request, next func()) {
	log.Println("start index route")
	_ = res.Send("index route")
	log.Println("end index route")
}

func v2Handler(res *soon.Response, req *http.Request, next func()) {
	log.Println("start v2 route")
	_ = res.Send("v2 route")
	log.Println("end v2 route")
}

func v3Handler(res *soon.Response, req *http.Request, next func()) {
	log.Println("start v3 route")
	_ = res.Send("v3 route")
	log.Println("end v3 route")
}

func panicHandler(res *soon.Response, req *http.Request, next func()) {
	log.Println("start panic route")
	panic("panic route")
}

func init() {
	IndexRouter.Use("/v2", func(writer *soon.Response, req *http.Request, next func()) {
		log.Println("[index router middleware] start")
		next()
		log.Println("[index router middleware] end")
	})
	IndexRouter.GET("", indexHandler)
	IndexRouter.GET("/", indexHandler)
	IndexRouter.HEAD("/v2", v2Handler)
	IndexRouter.POST("/v3/", v3Handler)
	IndexRouter.ALL("/panic", panicHandler)
}
