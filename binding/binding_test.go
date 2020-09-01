// Copyright 2014 Manu Martinez-Almeida.
// Portions copyright 2020 Guoyao Wu.
// All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FooStruct struct {
	Foo string `json:"foo" validate:"required"`
}

type FooStructUseNumber struct {
	Foo interface{} `json:"foo" validate:"required"`
}

type FooStructDisallowUnknownFields struct {
	Foo interface{} `json:"foo" validate:"required"`
}

func TestValidate(t *testing.T) {
	v := Validator
	Validator = nil
	assert.Equal(t, nil, validate(FooStruct{"foo"}))
	Validator = v
}

func TestBindingJSONUseNumber(t *testing.T) {
	testBodyBindingUseNumber(t,
		JSON,
		"/", "/",
		`{"foo": 123}`, `{"bar": "foo"}`)
}

func TestBindingJSONUseNumber2(t *testing.T) {
	testBodyBindingUseNumber2(t,
		JSON,
		"/", "/",
		`{"foo": 123}`, `{"bar": "foo"}`)
}

func TestBindingJSONDisallowUnknownFields(t *testing.T) {
	testBodyBindingDisallowUnknownFields(t, JSON,
		"/", "/",
		`{"foo": "bar"}`, `{"foo": "bar", "what": "this"}`)
}

func testBodyBindingUseNumber(t *testing.T, b Binding, path, badPath, body, badBody string) {
	EnableDecoderUseNumber = true
	defer func() {
		EnableDecoderUseNumber = false
	}()
	obj := FooStructUseNumber{}
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	// we hope it is int64(123)
	v, e := obj.Foo.(json.Number).Int64()
	assert.NoError(t, e)
	assert.Equal(t, int64(123), v)

	obj = FooStructUseNumber{}
	req = httptest.NewRequest("POST", badPath, strings.NewReader(badBody))
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func testBodyBindingUseNumber2(t *testing.T, b Binding, path, badPath, body, badBody string) {
	obj := FooStructUseNumber{}
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	EnableDecoderUseNumber = false
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	// it will return float64(123) if not use EnableDecoderUseNumber
	// maybe it is not hoped
	assert.Equal(t, float64(123), obj.Foo)

	obj = FooStructUseNumber{}
	req = httptest.NewRequest("POST", badPath, strings.NewReader(badBody))
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func testBodyBindingDisallowUnknownFields(t *testing.T, b Binding, path, badPath, body, badBody string) {
	EnableDecoderDisallowUnknownFields = true
	defer func() {
		EnableDecoderDisallowUnknownFields = false
	}()

	obj := FooStructDisallowUnknownFields{}
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStructDisallowUnknownFields{}
	req = httptest.NewRequest("POST", badPath, strings.NewReader(badBody))
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "what")
}
