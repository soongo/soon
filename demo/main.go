package main

import (
	"log"
	"net/http"
	"time"

	"github.com/soongo/soon"
	"github.com/soongo/soon/demo/routes"
)

func logger(res *soon.Response, req *http.Request, next func()) {
	log.Println("[log middleware] start")
	next()
	log.Println("[log middleware] end")
}

func timer(res *soon.Response, req *http.Request, next func()) {
	start := time.Now()
	log.Println("[time middleware] start")
	next()
	log.Printf("[time middleware] end, cost: %s", time.Since(start))
}

func notFound(res *soon.Response, req *http.Request, next func()) {
	http.Error(res, "404", http.StatusNotFound)
}

func main() {
	app := soon.New()
	app.Use(logger)
	app.Use(timer)
	app.Use("/", routes.IndexRouter)
	app.Use("/user/:id", routes.UserRouter)
	app.GET("/users", func(res *soon.Response, req *http.Request, next func()) {
		log.Println("start users router")
		_ = res.Send("users router")
		log.Println("end users router")
	})
	app.Use(notFound)
	log.Fatal(http.ListenAndServe(":8081", app))
}
