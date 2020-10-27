// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func rawBytesToStr(b []byte) string {
	return string(b)
}

func rawStrToBytes(s string) []byte {
	return []byte(s)
}

var (
	testString = "你好中国abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	testBytes  = []byte(testString)
)

func generateRandomString(n int) string {
	testStrings := strings.Split(testString, "")
	rand.Seed(time.Now().UnixNano())
	str := strings.Builder{}
	str.Grow(n)
	for i := 0; i < n; i++ {
		str.WriteString(testStrings[rand.Intn(len(testStrings))])
	}
	return str.String()
}

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
		assert.Equal(t, tt.expected, tt.s.Contains(tt.str))
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
		assert.Equal(t, tt.expected, tt.s.Index(tt.str))
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
		assert.Equal(t, tt.expected, tt.s.Filter(tt.filter))
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
		assert.Equal(t, tt.expected, tt.s.Map(tt.fn))
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
		assert.Equal(t, tt.expected, FileExists(tt.p))
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
		assert.Equal(t, v, EncodeURI(k))
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
		assert.Equal(t, v, EncodeURIComponent(k))
	}
}

func TestStringToBytes(t *testing.T) {
	for i := 0; i < 100; i++ {
		str := generateRandomString(20)
		require.Equal(t, rawStrToBytes(str), StringToBytes(str))
	}
}

func TestBytesToString(t *testing.T) {
	data := make([]byte, 1024)
	for i := 0; i < 100; i++ {
		rand.Read(data)
		require.Equal(t, rawBytesToStr(data), BytesToString(data))
	}
}

func BenchmarkStringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringToBytes(testString)
	}
}

func BenchmarkRawStrToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rawStrToBytes(testString)
	}
}

func BenchmarkBytesToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToString(testBytes)
	}
}

func BenchmarkRawBytesToStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rawBytesToStr(testBytes)
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
		assert.Equal(t, tt.expected, AddPrefixSlash(tt.p))
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
		assert.Equal(t, tt.expected, RouteJoin(tt.p1, tt.p2))
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		x        int
		y        int
		expected int
	}{
		{0, 1, 0},
		{1, 0, 0},
		{1, 1, 1},
		{-1, 1, -1},
		{-1, 0, -1},
		{-1, -2, -2},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, Min(tt.x, tt.y))
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		x        int
		y        int
		expected int
	}{
		{0, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
		{-1, 1, 1},
		{-1, 0, 0},
		{-1, -2, -1},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, Max(tt.x, tt.y))
	}
}
