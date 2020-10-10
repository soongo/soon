// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	urlpkg "net/url"
)

// Redirect contains the http request reference and redirects status code
// and location.
type Redirect struct {
	Code     int
	Location string

	url   string
	hadCT bool
}

// RenderHeader writes custom headers.
func (r *Redirect) RenderHeader(w http.ResponseWriter, req *http.Request) {
	url := r.Location
	if u, err := urlpkg.Parse(url); err == nil {
		// If url was relative, make its path absolute by
		// combining with request path.
		// The client would probably do this for us,
		// but doing it ourselves is more reliable.
		// See RFC 7231, section 7.1.2
		if u.Scheme == "" && u.Host == "" {
			oldpath := req.URL.Path
			if oldpath == "" { // should not happen, but avoid a crash if it does
				oldpath = "/"
			}

			// no leading http://server
			if url == "" || url[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				url = olddir + url
			}

			var query string
			if i := strings.Index(url, "?"); i != -1 {
				url, query = url[:i], url[i:]
			}

			// clean up but preserve trailing slash
			trailing := strings.HasSuffix(url, "/")
			url = path.Clean(url)
			if trailing && !strings.HasSuffix(url, "/") {
				url += "/"
			}
			url += query
		}
	}

	r.url = url

	h := w.Header()

	// RFC 7231 notes that a short HTML body is usually included in
	// the response because older user agents may not understand 301/307.
	// Do it only if the request didn't already have a Content-Type header.
	_, hadCT := h["Content-Type"]
	r.hadCT = hadCT

	h.Set("Location", hexEscapeNonASCII(url))
	if !hadCT && (req.Method == "GET" || req.Method == "HEAD") {
		h.Set("Content-Type", "text/html; charset=utf-8")
	}
	w.WriteHeader(r.Code)
}

// Render (Redirect) redirects the http request to new location
// and writes redirect response.
func (r *Redirect) Render(w http.ResponseWriter, req *http.Request) error {
	if (r.Code < 300 || r.Code > 308) && r.Code != 201 {
		return fmt.Errorf("cannot redirect with status code %d", r.Code)
	}

	// Shouldn't send the body for POST or HEAD; that leaves GET.
	if !r.hadCT && req.Method == "GET" {
		body := "<a href=\"" + htmlEscape(r.url) + "\">" + http.StatusText(r.Code) + "</a>.\n"
		_, err := fmt.Fprintln(w, body)
		return err
	}

	return nil
}

// hexEscapeNonASCII is a copy of http.hexEscapeNonASCII non-exported function.
func hexEscapeNonASCII(s string) string {
	newLen := 0
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			newLen += 3
		} else {
			newLen++
		}
	}
	if newLen == len(s) {
		return s
	}
	b := make([]byte, 0, newLen)
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			b = append(b, '%')
			b = strconv.AppendInt(b, int64(s[i]), 16)
		} else {
			b = append(b, s[i])
		}
	}
	return string(b)
}

var htmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	// "&#34;" is shorter than "&quot;".
	`"`, "&#34;",
	// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	"'", "&#39;",
)

func htmlEscape(s string) string {
	return htmlReplacer.Replace(s)
}
