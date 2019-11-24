// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import "testing"

func TestAddPrefixSlash(t *testing.T) {
	name := "addPrefixSlash"
	tests := []struct {
		p    string
		want string
	}{
		{"", "/"},
		{"/", "/"},
		{"abc", "/abc"},
		{"/abc", "/abc"},
		{"//abc", "//abc"},
	}
	for _, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := addPrefixSlash(tt.p); got != tt.want {
				t.Errorf("addPrefixSlash(%s) = %s, want %s", tt.p, got, tt.want)
			}
		})
	}
}

func TestRouteJoin(t *testing.T) {
	name := "routeJoin"
	tests := []struct {
		p1   string
		p2   string
		want string
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
			if got := routeJoin(tt.p1, tt.p2); got != tt.want {
				t.Errorf("routeJoin(%s, %s) = %s, want %s", tt.p1, tt.p2, got, tt.want)
			}
		})
	}
}
