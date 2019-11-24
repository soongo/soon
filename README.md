# Soon Web Framework

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
	app.GET("/", func(res *soon.Response, req *http.Request, next func()) {
		res.Send("Hello World")
	})
	log.Fatal(http.ListenAndServe(":3000", app))
}
```
