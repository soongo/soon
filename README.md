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
	"log"
	"net/http"

	"github.com/soongo/soon"
)

func main() {
	app := soon.New()
	app.GET("/", func(req *http.Request, res *soon.Response, next func()) {
		res.Send("Hello World")
	})
	log.Fatal(http.ListenAndServe(":3000", app))
}
```
