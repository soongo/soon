// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"regexp"
	"strings"
)

var charsetRegexp = regexp.MustCompile(";\\s*charset\\s*=")

// Sets the response’s HTTP header field to value.
// To set multiple fields at once, pass a string map as the parameter.
func SetHeader(w http.ResponseWriter, value ...interface{}) {
	if len(value) == 2 {
		if k, ok := value[0].(string); ok {
			if v, ok := value[1].(string); ok {
				if strings.ToLower(k) == "content-type" && !charsetRegexp.MatchString(v) {
					charset := LookupCharset(strings.Split(v, ";")[0])
					if charset != "" {
						v += "; charset=" + charset
					}
				}
				w.Header().Set(k, v)
			}
		}
		return
	}

	if len(value) == 1 {
		if arr, ok := value[0].(map[string]string); ok {
			for k, v := range arr {
				SetHeader(w, k, v)
			}
		}
	}
}

// Sets the Content-Type HTTP header to the MIME type as determined
// by LookupMimeType() for the specified type. If type contains the
// “/” character, then it sets the Content-Type to type.
func SetContentType(w http.ResponseWriter, s string) {
	k, s := "Content-Type", strings.Trim(s, " ")
	if strings.Contains(s, "/") {
		SetHeader(w, k, s)
	} else {
		SetHeader(w, k, LookupMimeType(s))
	}
}
