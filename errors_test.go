// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusError_Error(t *testing.T) {
	tests := []struct {
		statusError      *statusError
		expectedCode     int
		expectedErrorStr string
	}{
		{&statusError{200, errors.New("error")}, 200, "error"},
		{&statusError{200, nil}, 200, ""},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expectedCode, tt.statusError.status())
		assert.Equal(t, tt.expectedErrorStr, tt.statusError.Error())
	}
}
