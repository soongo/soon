// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import "testing"

var testErrorFormat = "got `%v`, expect `%v`"

func TestStringSlice(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		arr := StringSlice{"foo", "*", "你好"}
		tests := []struct {
			s        StringSlice
			str      string
			expected bool
		}{
			{arr, "foo", true},
			{arr, "bar", false},
			{arr, "*", true},
			{arr, "你好", true},
		}
		for _, tt := range tests {
			if got := tt.s.Contains(tt.str); got != tt.expected {
				t.Errorf(testErrorFormat, got, tt.expected)
			}
		}
	})
}

func TestAddPrefixSlash(t *testing.T) {
	name := "addPrefixSlash"
	tests := []struct {
		p        string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{"abc", "/abc"},
		{"/abc", "/abc"},
		{"//abc", "//abc"},
	}
	for _, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := AddPrefixSlash(tt.p); got != tt.expected {
				t.Errorf(testErrorFormat, got, tt.expected)
			}
		})
	}
}

func TestRouteJoin(t *testing.T) {
	name := "routeJoin"
	tests := []struct {
		p1       string
		p2       string
		expected string
	}{
		{"", "", ""},
		{"/", "/", "/"},
		{"abc", "/123", "abc/123"},
		{"abc/", "/123", "abc/123"},
		{"abc//", "123", "abc//123"},
		{"abc//", "/123", "abc//123"},
		{"abc//", "//123", "abc///123"},
	}
	for _, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := RouteJoin(tt.p1, tt.p2); got != tt.expected {
				t.Errorf(testErrorFormat, got, tt.expected)
			}
		})
	}
}
