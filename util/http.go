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

// Vary marks that a request is varied on a header field.
func Vary(w http.ResponseWriter, fields []string) {
	k := "Vary"
	header := strings.Join(w.Header().Values(k), ",")
	if val := AppendToVaryHeader(header, fields); val != "" {
		w.Header().Set(k, val)
	}
}

// AppendToVaryHeader appends fields to the vary field of header.
func AppendToVaryHeader(vary string, fields []string) string {
	if fields == nil {
		return vary
	}

	// existing, unspecified vary
	if vary == "*" {
		return vary
	}

	// enumerate current values
	val, vals := vary, StringSlice(ParseHeader(strings.ToLower(vary)))

	// unspecified vary
	if StringSlice(fields).Contains("*") || vals.Contains("*") {
		return "*"
	}

	for i := 0; i < len(fields); i++ {
		fld := strings.ToLower(fields[i])

		// append value (case-preserving)
		if !vals.Contains(fld) {
			vals = append(vals, fld)
			if val != "" {
				val += ", "
			}
			val += fields[i]
		}
	}

	return val
}

// ParseHeader parses header with type string into a slice.
func ParseHeader(header string) []string {
	start, end, length := 0, 0, len(header)
	values := make([]string, 0, length)

	// gather tokens
	for i := 0; i < length; i++ {
		switch header[i] {
		case 0x20: /*   */
			if start == end {
				start, end = i+1, i+1
			}
			break
		case 0x2c: /* , */
			values = append(values, header[start:end])
			start, end = i+1, i+1
			break
		default:
			end = i + 1
			break
		}
	}

	// final token
	values = append(values, header[start:end])

	return values
}
