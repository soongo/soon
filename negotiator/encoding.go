// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

type acceptEncoding struct {
	encoding string
	q        float64
	i        int
}

var simpleEncodingRegExp = regexp2.MustCompile("^\\s*([^\\s;]+)\\s*(?:;(.*))?$", regexp2.None)

// Parse an encoding from the Accept-Encoding header.
func parseEncoding(s string, i int) *acceptEncoding {
	match, err := simpleEncodingRegExp.FindStringMatch(s)
	if match == nil || match.GroupCount() == 0 || err != nil {
		return nil
	}

	encoding, q := match.Groups()[1].String(), 1.0
	if match.Groups()[2].String() != "" {
		params := strings.Split(match.Groups()[2].String(), ";")
		for j := 0; j < len(params); j++ {
			p := strings.Split(strings.Trim(params[j], " "), "=")
			if p[0] == "q" {
				q1, err := strconv.ParseFloat(p[1], 64)
				if err != nil {
					return nil
				}
				q = q1
				break
			}
		}
	}

	return &acceptEncoding{encoding, q, i}
}
