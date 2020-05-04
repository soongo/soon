// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"testing"
)

var dotRegexp = regexp.MustCompile("\\s*,\\s*")

type charsetTestObj struct {
	negotiator *Negotiator
	available  []string
	expected   []string
}

func TestNegotiator_Charset(t *testing.T) {
	for _, tt := range newCharsetTestObjs() {
		expected := ""
		if len(tt.expected) > 0 {
			expected = tt.expected[0]
		}
		if got := tt.negotiator.Charset(tt.available...); !reflect.DeepEqual(got, expected) {
			t.Errorf(testErrorFormat, got, expected)
		}
	}
}

func TestNegotiator_Charsets(t *testing.T) {
	for _, tt := range newCharsetTestObjs() {
		if got := tt.negotiator.Charsets(tt.available...); !reflect.DeepEqual(got, tt.expected) {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func newCharsetTestObjs() []charsetTestObj {
	results := make([]charsetTestObj, len(preferredCharsetTestObjs), len(preferredCharsetTestObjs))
	for i, obj := range preferredCharsetTestObjs {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header = http.Header{AcceptCharsetHeader: dotRegexp.Split(obj.accept, -1)}
		results[i] = charsetTestObj{&Negotiator{req}, obj.provided, obj.expected}
	}
	return results
}
