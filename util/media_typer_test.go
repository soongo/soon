// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var invalidTypes = []string{
	" ",
	"null",
	"undefined",
	"/",
	"text/;plain",
	"text/\"plain\"",
	"text/p£ain",
	"text/(plain)",
	"text/@plain",
	"text/plain,wrong",
}

func TestFormat(t *testing.T) {
	tests := []struct {
		desc           string
		mediaType      MediaType
		expected       string
		expectedErrMsg string
	}{
		{
			desc:      "should format basic type",
			mediaType: MediaType{MainType: "text", Subtype: "html"},
			expected:  "text/html",
		},
		{
			desc:      "should format type with suffix",
			mediaType: MediaType{MainType: "image", Subtype: "svg", Suffix: "xml"},
			expected:  "image/svg+xml",
		},
		{
			desc:           "should require main type",
			mediaType:      MediaType{},
			expectedErrMsg: "invalid main type",
		},
		{
			desc:           "should reject invalid main type",
			mediaType:      MediaType{MainType: "text/"},
			expectedErrMsg: "invalid main type",
		},
		{
			desc:           "should require subtype",
			mediaType:      MediaType{MainType: "text"},
			expectedErrMsg: "invalid subtype",
		},
		{
			desc:           "should reject invalid subtype",
			mediaType:      MediaType{MainType: "text", Subtype: "html/"},
			expectedErrMsg: "invalid subtype",
		},
		{
			desc:           "should reject invalid suffix",
			mediaType:      MediaType{MainType: "image", Subtype: "svg", Suffix: "xml\\\\"},
			expectedErrMsg: "invalid suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result, err := FormatMediaType(tt.mediaType)
			if tt.expectedErrMsg == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		desc           string
		s              string
		expected       MediaType
		expectedErrMsg string
	}{
		{
			desc:     "should parse basic type",
			s:        "text/html",
			expected: MediaType{MainType: "text", Subtype: "html"},
		},
		{
			desc:     "should parse with suffix",
			s:        "image/svg+xml",
			expected: MediaType{MainType: "image", Subtype: "svg", Suffix: "xml"},
		},
		{
			desc:     "should lower-case type",
			s:        "IMAGE/SVG+XML",
			expected: MediaType{MainType: "image", Subtype: "svg", Suffix: "xml"},
		},
		{
			desc:           "should require argument",
			expectedErrMsg: "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result, err := ParseMediaType(tt.s)
			if tt.expectedErrMsg == "" {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErrMsg, err.Error())
			}
		})
	}

	for _, invalidType := range invalidTypes {
		t.Run("should throw on invalid media type "+invalidType, func(t *testing.T) {
			_, err := ParseMediaType(invalidType)
			require.Error(t, err)
			assert.Equal(t, "invalid media type", err.Error())
		})
	}
}

func TestTest(t *testing.T) {
	tests := []struct {
		desc     string
		s        string
		expected bool
	}{
		{
			desc:     "should pass basic type",
			s:        "text/html",
			expected: true,
		},
		{
			desc:     "should pass with suffix",
			s:        "image/svg+xml",
			expected: true,
		},
		{
			desc:     "should pass upper-case type",
			s:        "IMAGE/SVG+XML",
			expected: true,
		},
		{
			desc:     "should require argument",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.expected, TestMediaType(tt.s))
		})
	}

	for _, invalidType := range invalidTypes {
		t.Run("should fail invalid media type "+invalidType, func(t *testing.T) {
			assert.Equal(t, false, TestMediaType(invalidType))
		})
	}
}
