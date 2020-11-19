// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"path/filepath"
	"strings"

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

		// ignore root option in FileOptions
		var opts renderer.FileOptions
		if len(options) > 0 {
			opts = options[0]
			opts.Root = ""
		}

		c.SendFile(absPath, opts)
	}
}
