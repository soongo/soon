// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"testing"

	"github.com/soongo/soon/binding"

	"github.com/stretchr/testify/assert"
)

func TestSetMode(t *testing.T) {
	tests := []struct {
		mode    string
		deferFn func()
	}{
		{ReleaseMode, nil},
		{DebugMode, nil},
		{TestMode, nil},
		{
			"Unknown",
			func() {
				assert.NotNil(t, recover())
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
			assert.Equal(t, tt.mode, Mode())
		}
	}
}

func TestDisableBindValidation(t *testing.T) {
	v := binding.Validator
	assert.NotNil(t, binding.Validator)
	DisableBindValidation()
	assert.Nil(t, binding.Validator)
	binding.Validator = v
}

func TestEnableJsonDecoderUseNumber(t *testing.T) {
	assert.False(t, binding.EnableDecoderUseNumber)
	EnableJsonDecoderUseNumber()
	assert.True(t, binding.EnableDecoderUseNumber)
}

func TestEnableJsonDecoderDisallowUnknownFields(t *testing.T) {
	assert.False(t, binding.EnableDecoderDisallowUnknownFields)
	EnableJsonDecoderDisallowUnknownFields()
	assert.True(t, binding.EnableDecoderDisallowUnknownFields)
}
