# Soon Web Framework

[![Build Status](https://travis-ci.org/soongo/soon.svg)](https://travis-ci.org/soongo/soon)
[![codecov](https://codecov.io/gh/soongo/soon/branch/master/graph/badge.svg)](https://codecov.io/gh/soongo/soon)
[![Go Report Card](https://goreportcard.com/badge/github.com/soongo/soon)](https://goreportcard.com/report/github.com/soongo/soon)
[![GoDoc](https://godoc.org/github.com/soongo/soon?status.svg)](https://godoc.org/github.com/soongo/soon)
[![License](https://img.shields.io/badge/MIT-green.svg)](https://opensource.org/licenses/MIT)

Soon is a web framework written in Go (Golang). It features an expressjs-like API.

## Installation

To install `Soon`, you need to install Go and set your Go workspace first.

The first need [Go](https://golang.org/) installed (**version 1.11+ is required**), then you can use the below Go command to install `Soon`.

```sh
$ go get -u github.com/soongo/soon
```

## Quick Start

```go
package main

import (
	"github.com/soongo/soon"
)

// an example middleware
func logger(c *soon.Context) {
	// do something before
	c.Next()
	// do something after
}

func main() {
	// soon.SetMode(soon.DebugMode) // enable soon framework debug mode

	// Create an app with default router
	app := soon.New()

	app.Use(logger) // use middleware

	app.GET("/", func(c *soon.Context) {
		c.Send("Hello World")
	})

	app.GET("/:foo", func(c *soon.Context) {
		c.Send(c.Params().Get("foo"))
	})

	// an example error handler
	app.Use(func(v interface{}, c *soon.Context) {
		msg := "Internal Server Error"
		switch err := v.(type) {
		case error:
			msg = err.Error()
		case string:
			msg = err
		}
		c.Status(500)
		c.Send(msg)
	})

	app.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```
