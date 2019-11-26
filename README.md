# Soon Web Framework

[![Build Status](https://travis-ci.org/soongo/soon.svg)](https://travis-ci.org/soongo/soon)
[![codecov](https://codecov.io/gh/soongo/soon/branch/master/graph/badge.svg)](https://codecov.io/gh/soongo/soon)
[![Go Report Card](https://goreportcard.com/badge/github.com/soongo/soon)](https://goreportcard.com/report/github.com/soongo/soon)
[![GoDoc](https://godoc.org/github.com/soongo/soon?status.svg)](https://godoc.org/github.com/soongo/soon)
[![License](https://img.shields.io/badge/MIT-green.svg)](https://opensource.org/licenses/MIT)

Soon is a web framework written in Go (Golang). It features an expressjs-like API.

## Quick Start

```go
package main

import (
	"github.com/soongo/soon"
)

// an example middleware
func logger(req *soon.Request, res *soon.Response, next soon.Next) {
	// do something before
	next(nil)
	// do something after
}

func main() {
	// soon.SetMode(soon.DebugMode) // enable soon framework debug mode

	// Create an app with default router
	app := soon.New()

	app.Use(logger) // use middleware

	app.GET("/", func(req *soon.Request, res *soon.Response, next soon.Next) {
		res.Send("Hello World")
	})

	app.GET("/:foo", func(req *soon.Request, res *soon.Response, next soon.Next) {
		res.Send(req.Params.Get("foo"))
	})

	app.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```
