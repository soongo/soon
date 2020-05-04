// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"net/http"
	"net/textproto"
	"strings"
)

var AcceptCharsetHeader = textproto.CanonicalMIMEHeaderKey("Accept-Charset")

type Negotiator struct {
	req *http.Request
}

func (n *Negotiator) Charset(available ...string) string {
	charsets := n.Charsets(available...)
	if len(charsets) == 0 {
		return ""
	}
	return charsets[0]
}

func (n *Negotiator) Charsets(available ...string) []string {
	accept, values := "", n.req.Header.Values(AcceptCharsetHeader)
	// RFC 2616 sec 14.2: no header = *
	if values == nil {
		accept = "*"
	} else {
		accept = strings.Join(values, ",")
	}

	return PreferredCharsets(accept, available...)
}
