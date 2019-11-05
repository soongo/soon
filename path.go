// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"regexp"
	"strings"
)

func addPrefixSlash(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}

func addSuffixSlash(p string) string {
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

func removeSuffixSlash(p string) string {
	if strings.HasSuffix(p, "/") {
		p = p[:len(p)-1]
	}
	return p
}

func isAncestor(ancestor, p string) bool {
	if ancestor == "/" {
		return true
	}

	return strings.HasPrefix(addSuffixSlash(addPrefixSlash(p)),
		addSuffixSlash(addPrefixSlash(ancestor)))
}

func similar(r, p string) bool {
	return regexp.MustCompile("^" + removeSuffixSlash(r) + "/?$").MatchString(p)
}
