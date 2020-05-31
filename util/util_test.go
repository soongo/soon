// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"os"
	"reflect"
	"testing"
)

var testErrorFormat = "got `%v`, expect `%v`"

func TestStringSlice_Contains(t *testing.T) {
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
}

func TestStringSlice_Index(t *testing.T) {
	arr := StringSlice{"foo", "*", "你好"}
	tests := []struct {
		s        StringSlice
		str      string
		expected int
	}{
		{arr, "foo", 0},
		{arr, "bar", -1},
		{arr, "*", 1},
		{arr, "你好", 2},
	}

	for _, tt := range tests {
		if got := tt.s.Index(tt.str); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestStringSlice_Filter(t *testing.T) {
	arr := StringSlice{"foo", "*", "你好"}
	tests := []struct {
		s        StringSlice
		filter   func(string) bool
		expected []string
	}{
		{
			arr,
			func(s string) bool {
				return s == "foo"
			},
			[]string{"foo"},
		},
		{
			arr,
			func(s string) bool {
				return len(s) >= 3
			},
			[]string{"foo", "你好"},
		},
		{
			arr,
			func(s string) bool {
				return len(s) >= 6
			},
			[]string{"你好"},
		},
	}

	for _, tt := range tests {
		if got := tt.s.Filter(tt.filter); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestStringSlice_Map(t *testing.T) {
	arr := StringSlice{"foo", "*", "你好"}
	tests := []struct {
		s        StringSlice
		fn       func(string) string
		expected []string
	}{
		{
			arr,
			func(s string) string {
				return s
			},
			[]string{"foo", "*", "你好"},
		},
		{
			arr,
			func(s string) string {
				return s + "bar"
			},
			[]string{"foobar", "*bar", "你好bar"},
		},
	}

	for _, tt := range tests {
		if got := tt.s.Map(tt.fn); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestFileExists(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tests := []struct {
		p        string
		expected bool
	}{
		{pwd, true},
		{"", false},
	}

	for _, tt := range tests {
		if got := FileExists(tt.p); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestEncodeURI(t *testing.T) {
	tests := map[string]string{
		"foo":                    "foo",
		"foo%bar":                "foo%25bar",
		"http://a.com/你好":        "http://a.com/%E4%BD%A0%E5%A5%BD",
		"https://a.com/Go_шеллы": "https://a.com/Go_%D1%88%D0%B5%D0%BB%D0%BB%D1%8B",
	}
	for k, v := range tests {
		if got := EncodeURI(k); got != v {
			t.Errorf(testErrorFormat, got, v)
		}
	}
}

func TestEncodeURIComponent(t *testing.T) {
	tests := map[string]string{
		"foo":                    "foo",
		"foo%bar":                "foo%25bar",
		"http://a.com/你好":        "http%3A%2F%2Fa.com%2F%E4%BD%A0%E5%A5%BD",
		"https://a.com/Go_шеллы": "https%3A%2F%2Fa.com%2FGo_%D1%88%D0%B5%D0%BB%D0%BB%D1%8B",
	}
	for k, v := range tests {
		if got := EncodeURIComponent(k); got != v {
			t.Errorf(testErrorFormat, got, v)
		}
	}
}

func TestAddPrefixSlash(t *testing.T) {
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
		if got := AddPrefixSlash(tt.p); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestRouteJoin(t *testing.T) {
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
		if got := RouteJoin(tt.p1, tt.p2); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}
