package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/soongo/soon"
	"github.com/soongo/soon/demo/routes"
)

func logger(w http.ResponseWriter, r *http.Request, next func()) {
	log.Println("[log middleware] start")
	next()
	log.Println("[log middleware] end")
}

func timer(w http.ResponseWriter, r *http.Request, next func()) {
	start := time.Now()
	log.Println("[time middleware] start")
	next()
	log.Printf("[time middleware] end, cost: %s", time.Since(start))
}

func notFound(w http.ResponseWriter, r *http.Request, next func()) {
	http.Error(w, "404", http.StatusNotFound)
}

func main() {
	app := soon.New()
	app.Use(logger)
	app.Use(timer)
	app.Use("/", routes.IndexRouter)
	app.Use("/user", routes.UserRouter)
	app.Get("/users", func(w http.ResponseWriter, req *http.Request, next func()) {
		log.Println("start users router")
		_, _ = io.WriteString(w, "users router")
		log.Println("end users router")
	})
	app.Use(notFound)
	log.Fatal(http.ListenAndServe(":8081", app))
}
