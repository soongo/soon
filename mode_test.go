// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"testing"
)

func TestSetMode(t *testing.T) {
	tests := []struct {
		mode    string
		deferFn func()
	}{
		{ReleaseMode, nil},
		{TestMode, nil},
		{DebugMode, nil},
		{
			"Unknown",
			func() {
				if got := recover(); got == nil {
					t.Errorf(testErrorFormat, got, "none nil error")
				}
			},
		},
	}

	for _, tt := range tests {
		if tt.deferFn != nil {
			func() {
				defer tt.deferFn()
				SetMode(tt.mode)
			}()
		} else {
			SetMode(tt.mode)
			if got := Mode(); got != tt.mode {
				t.Errorf(testErrorFormat, got, tt.mode)
			}
		}
	}
}
