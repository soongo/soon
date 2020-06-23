// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"net/http"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
)

var (
	charsetRegexp           = regexp.MustCompile(";\\s*charset\\s*=")
	acceptParamsRegexp      = regexp.MustCompile(" *; *")
	acceptParamsPartsRegexp = regexp.MustCompile(" *= *")
)

// AcceptParams is an object with `.value`, `.quality` and `.params`.
type AcceptParams struct {
	Value   string
	Quality float64
	Params  map[string]string
}

// AddHeader adds the specified value to the HTTP response header field.
// If the header is not already set, it creates the header with the specified
// value. The value parameter can be a string or a string slice.
func AddHeader(w http.ResponseWriter, k string, v interface{}) {
	if s, ok := v.(string); ok {
		if strings.ToLower(k) == "content-type" && !charsetRegexp.MatchString(s) {
			charset := LookupCharset(strings.Split(s, ";")[0])
			if charset != "" {
				s += "; charset=" + charset
			}
		}
		w.Header().Add(k, s)
	} else if arr, ok := v.([]string); ok {
		for _, s := range arr {
			AddHeader(w, k, s)
		}
	}
}

// SetHeader sets the response’s HTTP header field to value.
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
			} else if arr, ok := value[1].([]string); ok {
				for i, v := range arr {
					if i == 0 {
						SetHeader(w, k, v)
					} else {
						AddHeader(w, k, v)
					}
				}
			}
		}
		return
	}

	if len(value) == 1 {
		if m, ok := value[0].(map[string]string); ok {
			for k, v := range m {
				SetHeader(w, k, v)
			}
		}
	}
}

// SetContentType sets the Content-Type HTTP header to the MIME type as
// determined by LookupMimeType() for the specified type. If type contains
// the “/” character, then it sets the Content-Type to type.
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
	header := strings.Join(GetHeaderValues(w.Header(), k), ",")
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

// GetHeaderValues returns the header values of specified key.
// This is a patch function of http.Header.Values for go version lower than 1.4
func GetHeaderValues(h http.Header, key string) []string {
	if h == nil {
		return nil
	}
	return h[textproto.CanonicalMIMEHeaderKey(key)]
}

// NormalizeType normalizes the given `type`, for example "html" becomes "text/html".
func NormalizeType(t string) AcceptParams {
	if strings.Index(t, "/") >= 0 {
		return acceptParams(t, 0)
	}
	return AcceptParams{Value: LookupMimeType(t)}
}

// Parse accept params `str` returning an
// object with `.value`, `.quality` and `.params`.
func acceptParams(str string, index int) AcceptParams {
	parts := acceptParamsRegexp.Split(str, -1)
	ret := AcceptParams{Value: parts[0], Quality: 1}

	for i := 1; i < len(parts); i++ {
		pms := acceptParamsPartsRegexp.Split(parts[i], -1)
		if "q" == pms[0] {
			q, err := strconv.ParseFloat(pms[1], 64)
			if err == nil {
				ret.Quality = q
			}
		} else {
			if ret.Params == nil {
				ret.Params = make(map[string]string)
			}
			ret.Params[pms[0]] = pms[1]
		}
	}

	return ret
}
