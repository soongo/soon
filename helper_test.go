// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAddress(t *testing.T) {
	tests := []struct {
		p        []string
		envPort  string
		expected string
		deferFn  func()
	}{
		{nil, "", ":8080", nil},
		{[]string{}, "", ":8080", nil},
		{[]string{"3000"}, "", "3000", nil},
		{
			[]string{"3000", "8000"},
			"",
			"",
			func() {
				assert.NotNil(t, recover())
			},
		},
		{nil, "3000", ":3000", nil},
	}

	for _, tt := range tests {
		if tt.deferFn != nil {
			func() {
				defer tt.deferFn()
				resolveAddress(tt.p)
			}()
		} else {
			os.Setenv("PORT", tt.envPort)
			assert.Equal(t, tt.expected, resolveAddress(tt.p))
		}
	}
}
