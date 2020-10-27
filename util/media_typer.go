// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"strings"

	"github.com/dlclark/regexp2"
)

type MediaType struct {
	MainType string
	Subtype  string
	Suffix   string
}

// RegExp to match type in RFC 6838
//
// type-name = restricted-name
// subtype-name = restricted-name
// restricted-name = restricted-name-first *126restricted-name-chars
// restricted-name-first  = ALPHA / DIGIT
// restricted-name-chars  = ALPHA / DIGIT / "!" / "#" /
//                          "$" / "&" / "-" / "^" / "_"
// restricted-name-chars =/ "." ; Characters before first dot always
//                              ; specify a facet name
// restricted-name-chars =/ "+" ; Characters after last plus always
//                              ; specify a structured syntax suffix
// ALPHA =  %x41-5A / %x61-7A   ; A-Z / a-z
// DIGIT =  %x30-39             ; 0-9
var (
	subtypeNameRegexp = regexp2.MustCompile("^[A-Za-z0-9][A-Za-z0-9!#$&^_.-]{0,126}$", regexp2.None)
	typeNameRegexp    = regexp2.MustCompile("^[A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126}$", regexp2.None)
	typeRegexp        = regexp2.MustCompile("^ *([A-Za-z0-9][A-Za-z0-9!#$&^_-]{0,126})\\/([A-Za-z0-9][A-Za-z0-9!#$&^_.+-]{0,126}) *$", regexp2.None)
)

// Format object to media type
func FormatMediaType(m MediaType) (string, error) {
	if m.MainType == "" {
		return "", errors.New("invalid main type")
	}

	if b, err := typeNameRegexp.MatchString(m.MainType); !b || err != nil {
		return "", errors.New("invalid main type")
	}

	if m.Subtype == "" {
		return "", errors.New("invalid subtype")
	}

	if b, err := subtypeNameRegexp.MatchString(m.Subtype); !b || err != nil {
		return "", errors.New("invalid subtype")
	}

	// format as type/subtype
	result := m.MainType + "/" + m.Subtype

	// append +suffix
	if m.Suffix != "" {
		if b, err := typeNameRegexp.MatchString(m.Suffix); !b || err != nil {
			return "", errors.New("invalid suffix")
		}

		result += "+" + m.Suffix
	}

	return result, nil
}

// Test media type
func TestMediaType(s string) bool {
	if s == "" {
		return false
	}

	b, err := typeRegexp.MatchString(strings.ToLower(s))

	return b && err == nil
}

// Parse media type to object
func ParseMediaType(s string) (MediaType, error) {
	if s == "" {
		return MediaType{}, errors.New("invalid argument")
	}

	match, err := typeRegexp.FindStringMatch(strings.ToLower(s))
	if err != nil || match == nil {
		return MediaType{}, errors.New("invalid media type")
	}

	mainType, subtype, suffix := match.GroupByNumber(1).String(), match.GroupByNumber(2).String(), ""

	// suffix after last +
	index := strings.LastIndex(subtype, "+")
	if index != -1 {
		suffix = subtype[index+1:]
		subtype = subtype[0:index]
	}

	return MediaType{mainType, subtype, suffix}, nil
}
