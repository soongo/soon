// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/url"
	"os"
	"strings"
)

// StringSlice attaches the methods of Interface to []string.
type StringSlice []string

// Contains reports whether str is within the string slice.
func (s StringSlice) Contains(str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// Filter returns a new string slice with matched values.
func (s StringSlice) Filter(f func(string) bool) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

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

func AddPrefixSlash(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func RouteJoin(routes ...string) string {
	var route string
	for _, r := range routes {
		if strings.HasSuffix(route, "/") && strings.HasPrefix(r, "/") {
			route = route[:len(route)-1]
		}
		route += r
	}
	return route
}
