// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import "testing"

func TestAddPrefixSlash(t *testing.T) {
	tests := []struct {
		name string
		p    string
		want string
	}{
		{"empty", "", "/"},
		{"root", "/", "/"},
		{"without-prefix", "abc", "/abc"},
		{"with-prefix", "/abc", "/abc"},
		{"with-multiple-prefix", "//abc", "//abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addPrefixSlash(tt.p); got != tt.want {
				t.Errorf("addPrefixSlash(%s) = %s, want %s", tt.p, got, tt.want)
			}
		})
	}
}

func TestRemoveSuffixSlash(t *testing.T) {
	tests := []struct {
		name string
		p    string
		want string
	}{
		{"empty", "", ""},
		{"root", "/", ""},
		{"without-suffix", "abc", "abc"},
		{"with-suffix", "abc/", "abc"},
		{"with-multiple-suffix", "abc//", "abc/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeSuffixSlash(tt.p); got != tt.want {
				t.Errorf("removeSuffixSlash(%s) = %s, want %s", tt.p, got, tt.want)
			}
		})
	}
}

func TestIsAncestor(t *testing.T) {
	tests := []struct {
		name     string
		ancestor string
		p        string
		want     bool
	}{
		{"empty-1", "", "", true},
		{"empty-2", "", "/", true},
		{"empty-3", "", "//", true},
		{"empty-4", "", "/sub", true},
		{"empty-5", "", "/sub/sub", true},
		{"empty-2-reverse", "/", "", true},
		{"empty-3-reverse", "//", "", false},
		{"empty-4-reverse", "/sub", "", false},
		{"empty-5-reverse", "/sub/sub", "", false},

		{"root-1", "/", "", true},
		{"root-2", "/", "/", true},
		{"root-3", "/", "//", true},
		{"root-4", "/", "/sub", true},
		{"root-5", "/", "/sub/sub", true},
		{"root-1-reverse", "", "/", true},
		{"root-3-reverse", "//", "/", false},
		{"root-4-reverse", "/sub", "/", false},
		{"root-5-reverse", "/sub/sub", "/", false},

		{"sub-1", "/sub", "/sub", true},
		{"sub-2", "/sub", "/sub/", true},
		{"sub-3", "/sub", "/sub//", true},
		{"sub-4", "/sub", "/sub/sub", true},
		{"sub-5", "/sub", "/sub//sub", true},
		{"sub-6", "/sub", "/sub/sub/", true},
		{"sub-7", "/sub", "/sub/sub//", true},
		{"sub-8", "/sub", "/sub//sub//", true},

		{"multiple-root-1", "//", "", false},
		{"multiple-root-2", "//", "/", false},
		{"multiple-root-3", "//", "abc", false},
		{"multiple-root-4", "//", "/abc", false},

		{"prefix-match", "abc", "abcd", false},

		{"sub-without-root-prefix-1", "abc/", "abc", true},
		{"sub-without-root-prefix-2", "abc/", "abc/", true},
		{"sub-without-root-prefix-3", "abc/", "abc/123", true},
		{"sub-without-root-prefix-4", "abc/", "/abc", true},
		{"sub-without-root-prefix-5", "abc/", "/abc/", true},
		{"sub-without-root-prefix-6", "abc/", "/abc/123", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAncestor(tt.ancestor, tt.p); got != tt.want {
				t.Errorf("isAncestor(%s, %s) = %v, want %v", tt.ancestor, tt.p, got, tt.want)
			}
		})
	}
}

func TestSimilar(t *testing.T) {
	tests := []struct {
		name string
		r    string
		p    string
		want bool
	}{
		{"empty-1", "", "", true},
		{"empty-2", "", "/", true},
		{"empty-3", "", "//", false},
		{"empty-4", "", "///", false},
		{"empty-2-reverse", "/", "", true},
		{"empty-3-reverse", "//", "", false},
		{"empty-4-reverse", "///", "", false},

		{"root-1", "/", "", true},
		{"root-2", "/", "/", true},
		{"root-3", "/", "//", false},
		{"root-4", "/", "///", false},
		{"root-1-reverse", "", "/", true},
		{"root-3-reverse", "//", "/", true},
		{"root-4-reverse", "///", "/", false},

		{"sub-1", "/sub", "/sub", true},
		{"sub-2", "/sub", "/sub/", true},
		{"sub-3", "/sub", "/sub//", false},
		{"sub-4", "/sub", "/sub///", false},
		{"sub-2-reverse", "/sub/", "/sub", true},
		{"sub-3-reverse", "/sub//", "/sub", false},
		{"sub-4-reverse", "/sub///", "/sub", false},

		{"sub-sub-1", "/sub/sub", "/sub/sub", true},
		{"sub-sub-2", "/sub/sub", "/sub/sub/", true},
		{"sub-sub-3", "/sub/sub", "/sub/sub//", false},
		{"sub-sub-4", "/sub/sub", "/sub/sub///", false},
		{"sub-sub-2-reverse", "/sub/sub/", "/sub/sub", true},
		{"sub-sub-3-reverse", "/sub/sub//", "/sub/sub", false},
		{"sub-sub-4-reverse", "/sub/sub///", "/sub/sub", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := similar(tt.r, tt.p); got != tt.want {
				t.Errorf("similar(%s, %s) = %v, want %v", tt.r, tt.p, got, tt.want)
			}
		})
	}
}
