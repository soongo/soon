// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/url"
	"os"
	"reflect"
	"strings"
	"unsafe"
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

// Index returns the index of str in the string slice.
func (s StringSlice) Index(str string) int {
	for i, v := range s {
		if v == str {
			return i
		}
	}
	return -1
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

// Map returns a new string slice with modified values.
func (s StringSlice) Map(f func(string) string) []string {
	result := make([]string, len(s), len(s))
	for i, v := range s {
		result[i] = f(v)
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

// EncodeURI encodes a text string as a valid Uniform Resource Identifier (URI)
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

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) (b []byte) {
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// AddPrefixSlash adds a slash to the start of given string
func AddPrefixSlash(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

// RouteJoin joins all given routes
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

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
