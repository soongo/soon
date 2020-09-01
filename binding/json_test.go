// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jsonChild struct {
	Name   string `json:"name" validate:"required,min=3"`
	Age    int    `json:"age" validate:"gte=0,max=150"`
	Gender string `json:"gender" validate:"oneof=male female"`
}

type jsonRoot struct {
	Foo   string    `json:"foo" validate:"required"`
	Child jsonChild `json:"child"`
}

var jsonBindTests = []struct {
	s        jsonRoot
	json     string
	errs     []string
	expected jsonRoot
}{
	{
		json: `{"foo": "FOO", "child": {"name": "matt", "age": 39, "gender": "male"}}`,
		expected: jsonRoot{
			Foo:   "FOO",
			Child: jsonChild{Name: "matt", Age: 39, Gender: "male"},
		},
	},
	{
		json: `{"foo": "FOO", "child": {"name": "hi", "age": -1, "gender": "x"}}`,
		expected: jsonRoot{
			Foo:   "FOO",
			Child: jsonChild{Name: "hi", Age: -1, Gender: "x"},
		},
		errs: []string{
			"Key: 'jsonRoot.Child.Name' Error:Field validation for 'Name' failed on the 'min' tag",
			"Key: 'jsonRoot.Child.Age' Error:Field validation for 'Age' failed on the 'gte' tag",
			"Key: 'jsonRoot.Child.Gender' Error:Field validation for 'Gender' failed on the 'oneof' tag",
		},
	},
	{
		json: `{"foo": ""}`,
		expected: jsonRoot{
			Foo: "",
		},
		errs: []string{
			"Key: 'jsonRoot.Foo' Error:Field validation for 'Foo' failed on the 'required' tag",
			"Key: 'jsonRoot.Child.Name' Error:Field validation for 'Name' failed on the 'required' tag",
			"Key: 'jsonRoot.Child.Gender' Error:Field validation for 'Gender' failed on the 'oneof' tag",
		},
	},
}

func TestJsonBinding_Bind(t *testing.T) {
	for _, tt := range jsonBindTests {
		req := httptest.NewRequest("GET", "/", strings.NewReader(tt.json))
		err := jsonBinding{}.Bind(req, &tt.s)
		if tt.errs == nil {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.s)
		} else {
			require.Error(t, err)
			require.EqualError(t, err, strings.Join(tt.errs, "\n"))
			assert.Equal(t, tt.expected, tt.s)
		}
	}

	req := httptest.NewRequest("GET", "/", nil)
	err := jsonBinding{}.Bind(req, &jsonRoot{})
	require.Error(t, err)

	err = jsonBinding{}.Bind(nil, &jsonRoot{})
	require.Error(t, err)
}

func TestJsonBinding_BindBody(t *testing.T) {
	for _, tt := range jsonBindTests {
		err := jsonBinding{}.BindBody([]byte(tt.json), &tt.s)
		if tt.errs == nil {
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tt.s)
		} else {
			require.Error(t, err)
			require.EqualError(t, err, strings.Join(tt.errs, "\n"))
			assert.Equal(t, tt.expected, tt.s)
		}
	}
}
