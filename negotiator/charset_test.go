// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"reflect"
	"testing"
)

var testErrorFormat = "got `%v`, expect `%v`"

var preferredCharsetTestObjs = []struct {
	accept   string
	provided []string
	expected []string
}{
	{
		"utf-8",
		nil,
		[]string{"utf-8"},
	},
	{"utf-8, iso-8859-1",
		nil,
		[]string{"utf-8", "iso-8859-1"},
	},
	{"utf-8, iso-8859-1;q=0.8",
		nil,
		[]string{"utf-8", "iso-8859-1"},
	},
	{"utf-8, iso-8859-1;q=0.8, utf-7;q=0.2",
		nil,
		[]string{"utf-8", "iso-8859-1", "utf-7"},
	},
	{"utf-8, iso-8859-1;q=0.8, utf-7;q=0.9",
		nil,
		[]string{"utf-8", "utf-7", "iso-8859-1"},
	},
	{"utf-8",
		[]string{"utf-8", "iso-8859-1"},
		[]string{"utf-8"},
	},
	{"utf-8, iso-8859-1",
		[]string{"utf-8", "iso-8859-1"},
		[]string{"utf-8", "iso-8859-1"},
	},
	{"utf-8, iso-8859-1",
		[]string{"utf-8"},
		[]string{"utf-8"},
	},
	{"utf-8, iso-8859-1;q=0.8",
		[]string{"utf-8", "iso-8859-1"},
		[]string{"utf-8", "iso-8859-1"},
	},
	{"utf-8, iso-8859-2;q=0.8",
		[]string{"utf-8", "iso-8859-1"},
		[]string{"utf-8"},
	},
	{"utf-8, iso-8859-1;q=0.8, utf-7;q=0.2",
		[]string{"utf-8", "iso-8859-1"},
		[]string{"utf-8", "iso-8859-1"},
	},
	{"utf-8, iso-8859-1;q=0.8, utf-7;q=0.2",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{"utf-8", "iso-8859-1", "utf-7"},
	},
	{"utf-8;q=0.1, iso-8859-1;q=0.1, utf-7;q=0.2",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{"utf-7", "utf-8", "iso-8859-1"},
	},
	{"utf-8;q=0.1, iso-8859-1;q=0.2, utf-7;q=0.3",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{"utf-7", "iso-8859-1", "utf-8"},
	},
	{"utf-8;q=0.1, iso-8859-1;q=0.2, utf-7;q=0.2",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{"iso-8859-1", "utf-7", "utf-8"},
	},
	{"utf-8;q=0.1, iso-8859-1;q=x, utf-7;q=0.2",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{"utf-7", "utf-8"},
	},
	{"utf-8, iso-8859-1;q=0.8, utf-7;q=0.2",
		[]string{"iso-8859-12", "utf-7"},
		[]string{"utf-7"},
	},
	{"",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{},
	},
	{"utf-8, iso-8859-1;q=0.8, utf-7;q=0.2",
		[]string{},
		[]string{"utf-8", "iso-8859-1", "utf-7"},
	},
	{"*",
		[]string{},
		[]string{"*"},
	},
	{"*",
		[]string{"utf-8"},
		[]string{"utf-8"},
	},
	{"*",
		[]string{"utf-8", "iso-8859-1", "utf-7"},
		[]string{"utf-8", "iso-8859-1", "utf-7"},
	},
	{"*, utf-8",
		[]string{},
		[]string{"*", "utf-8"},
	},
	{"*, utf-8;q=x",
		[]string{},
		[]string{"*"},
	},
}

func TestPreferredCharsets(t *testing.T) {
	for _, tt := range preferredCharsetTestObjs {
		if got := PreferredCharsets(tt.accept, tt.provided...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseAcceptCharset(t *testing.T) {
	tests := []struct {
		s        string
		expected acceptCharsets
	}{
		{"utf-8", acceptCharsets{{"utf-8", 1, 0}}},
		{
			"utf-8, iso-8859-1;q=0.8, utf-7;q=0.2",
			acceptCharsets{
				{"utf-8", 1, 0},
				{"iso-8859-1", .8, 1},
				{"utf-7", .2, 2},
			},
		},
	}
	for _, tt := range tests {
		if got := parseAcceptCharset(tt.s); !acceptCharsetEquals(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestParseCharset(t *testing.T) {
	tests := []struct {
		s        string
		i        int
		expected *acceptCharset
	}{
		{"utf-8", 0, &acceptCharset{"utf-8", 1, 0}},
		{"iso-8859-1;q=0.8", 1, &acceptCharset{"iso-8859-1", .8, 1}},
		{" utf-7 ; q=0.2 ", 2, &acceptCharset{"utf-7", .2, 2}},
		{"utf-16;q=x", 3, nil},
	}
	for _, tt := range tests {
		got := parseCharset(tt.s, tt.i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestGetCharsetPriority(t *testing.T) {
	acs := acceptCharsets{
		{"utf-8", 1, 0},
		{"iso-8859-1", .8, 1},
		{"utf-7", .2, 2},
	}
	tests := []struct {
		charset  string
		acs      acceptCharsets
		index    int
		expected specificity
	}{
		{"utf-8", acceptCharsets{}, 0, specificity{0, -1, 0, 0}},
		{"iso-8859-1", acs, 1, specificity{1, 1, 0.8, 1}},
		{"utf-7", acs, 2, specificity{2, 2, 0.2, 1}},
	}
	for _, tt := range tests {
		got := getCharsetPriority(tt.charset, tt.acs, tt.index)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestSpecify(t *testing.T) {
	tests := []struct {
		charset  string
		ac       acceptCharset
		index    int
		expected *specificity
	}{
		{"utf-8",
			acceptCharset{"utf-8", 1, 0},
			0,
			&specificity{0, 0, 1, 1},
		},
		{"iso-8859-1",
			acceptCharset{"iso-8859-1", .8, 1},
			1,
			&specificity{1, 1, .8, 1},
		},
		{"utf-7",
			acceptCharset{"utf-7", .2, 2},
			2,
			&specificity{2, 2, .2, 1},
		},
		{"utf-16",
			acceptCharset{"utf-32", .3, 3},
			3,
			nil,
		},
		{"utf-16",
			acceptCharset{"*", .4, 4},
			4,
			&specificity{4, 4, .4, 0},
		},
		{"*",
			acceptCharset{"utf-8", .5, 5},
			5,
			nil,
		},
		{"*",
			acceptCharset{"*", .6, 6},
			6,
			&specificity{6, 6, .6, 1},
		},
	}
	for i, tt := range tests {
		got := specify(tt.charset, tt.ac, i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func acceptCharsetEquals(a, b acceptCharsets) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}
