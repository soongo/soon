package routes

import (
	"io"
	"log"
	"net/http"

	"github.com/soongo/soon"
)

var IndexRouter = soon.NewRouter()

func indexHandler(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("start index route")
	_, _ = io.WriteString(w, "index route")
	log.Println("end index route")
}

func v2Handler(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("start v2 route")
	_, _ = io.WriteString(w, "v2 route")
	log.Println("end v2 route")
}

func v3Handler(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("start v3 route")
	_, _ = io.WriteString(w, "v3 route")
	log.Println("end v3 route")
}

func panicHandler(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("start panic route")
	panic("panic route")
}

func init() {
	IndexRouter.Use(func(writer http.ResponseWriter, req *http.Request, next func()) {
		log.Println("[index router middleware] start")
		next()
		log.Println("[index router middleware] end")
	})
	IndexRouter.Get("", indexHandler)
	IndexRouter.Get("/", indexHandler)
	IndexRouter.Head("/v2", v2Handler)
	IndexRouter.Post("/v3/", v3Handler)
	IndexRouter.All("/panic", panicHandler)
}
