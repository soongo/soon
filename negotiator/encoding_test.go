// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"reflect"
	"testing"
)

func TestParseEncoding(t *testing.T) {
	// gzip, compress;q=0.2, identity;q=0.5
	tests := []struct {
		s        string
		i        int
		expected *acceptEncoding
	}{
		{"gzip", 0, &acceptEncoding{"gzip", 1, 0}},
		{"compress;q=0.2", 1, &acceptEncoding{"compress", .2, 1}},
		{" compress ; q=0.2 ", 2, &acceptEncoding{"compress", .2, 2}},
		{"gzip;q=x", 3, nil},
	}
	for _, tt := range tests {
		got := parseEncoding(tt.s, tt.i)
		if got == nil && tt.expected != nil || !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}
