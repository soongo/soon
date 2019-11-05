# Soon Web Framework

Soon is a web framework written in Go (Golang). It features an expressjs-like API.

## Quick Start

```go
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/soongo/soon"
)

func main() {
	app := soon.New()
	app.Get("/", func(w http.ResponseWriter, req *http.Request, next func()) {
		io.WriteString(w, "Hello World")
	})
	log.Fatal(http.ListenAndServe(":3000", app))
}
```
