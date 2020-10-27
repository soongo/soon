// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupMimeType(t *testing.T) {
	tests := map[string]string{
		"examples/test.png": "image/png",
		".png":              "image/png",
		"png":               "image/png",
		"urlencoded":        "application/x-www-form-urlencoded",
		"examples/test":     "application/octet-stream",
	}

	for k, v := range tests {
		assert.Equal(t, v, LookupMimeType(k))
	}
}

func TestLookupCharset(t *testing.T) {
	tests := map[string]string{
		"text/html":                "UTF-8",
		"text/plain":               "UTF-8",
		"text/xxx":                 "UTF-8",
		"application/javascript":   "UTF-8",
		"application/json":         "UTF-8",
		"application/octet-stream": "",
		"application/xxx":          "",
		"image/png":                "",
	}

	for k, v := range tests {
		assert.Equal(t, v, LookupCharset(k))
	}
}
