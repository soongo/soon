// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/url"
	"os"
	"strings"
)

// FileExists returns true if specified absolute file path exists.
func FileExists(absPath string) bool {
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Encodes a text string as a valid Uniform Resource Identifier (URI)
func EncodeURI(str string) string {
	excludes := ";/?:@&=+$,#"
	arr := strings.Split(str, "")
	result := ""
	for _, v := range arr {
		if strings.Contains(excludes, v) {
			result += v
		} else {
			result += EncodeURIComponent(v)
		}
	}
	return result
}

// EncodeURIComponent encodes a text string as a valid component of a Uniform
// Resource Identifier (URI).
func EncodeURIComponent(str string) string {
	r := url.QueryEscape(str)
	r = strings.Replace(r, "+", "%20", -1)
	return r
}

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

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		}
		debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}
