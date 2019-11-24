// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"strings"
)

func addPrefixSlash(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func routeJoin(routes ...string) string {
	var route string
	for _, r := range routes {
		if strings.HasSuffix(route, "/") && strings.HasPrefix(r, "/") {
			route = route[:len(route)-1]
		}
		route += r
	}
	return route
}
