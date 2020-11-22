// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"

	pathToRegexp "github.com/soongo/path-to-regexp"
)

// Static is a built-in middleware function in Soon. It serves static files.
//
// The root argument specifies the root directory from which to serve static
// assets. The function determines the file to serve by combining req.url with
// the provided root directory. When a file is not found, instead of sending a
// 404 response, it instead calls next() to move on to the next middleware,
// allowing for stacking and fall-backs.
func Static(root string, options ...renderer.FileOptions) Handle {
	return func(c *Context) {
		if !filepath.IsAbs(root) {
			dirname, err := util.Dirname()
			if err != nil {
				panic(err)
			}

			root = filepath.Join(dirname, root)
		}

		path := strings.TrimPrefix(c.Request.Path, c.Request.BaseUrl)
		absPath := pathToRegexp.DecodeURIComponent(filepath.Join(root, path))

		if !util.IsFileExist(absPath) {
			c.Next()
			return
		}

		c.SendFile(absPath, options...)
	}
}

// DevLog is a built-in middleware function in Soon.
// It just print simple log for development.
// For production environment, you can use `https://github.com/sirupsen/logrus`
func DevLog() Handle {
	g, r, y := color.New(color.FgGreen).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgYellow).SprintFunc()

	return func(c *Context) {
		req, w := c.Request, c.Writer
		method, path := req.Method, req.URL.String()
		start := time.Now()
		c.Next()
		cost, status, size := time.Since(start), w.Status(), w.Size()
		statusText := g(status)
		if status >= 500 {
			statusText = r(status)
		} else if status >= 400 {
			statusText = y(status)
		}
		fmt.Printf("%s %s %s %s - %d\n", method, path, statusText, cost, size)
	}
}
