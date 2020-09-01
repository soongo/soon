// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"io"
	"os"
	"strings"

	"github.com/soongo/soon/binding"
)

// EnvSoonMode indicates environment name for soon mode.
const EnvSoonMode = "SOON_MODE"

const (
	// DebugMode indicates soon mode is debug.
	DebugMode = "debug"

	// ReleaseMode indicates soon mode is release.
	ReleaseMode = "release"

	// TestMode indicates soon mode is test.
	TestMode = "test"
)

const (
	debugCode = iota
	releaseCode
	testCode
)

// DefaultWriter is the default io.Writer used by Soon for debug output.
// To support coloring in Windows use:
// 		import "github.com/mattn/go-colorable"
// 		soon.DefaultWriter = colorable.NewColorableStdout()
var DefaultWriter io.Writer = os.Stdout

// DefaultErrorWriter is the default io.Writer used by Soon to debug errors
var DefaultErrorWriter io.Writer = os.Stderr

var mode = ReleaseMode
var modeCode = releaseCode

func init() {
	mode := os.Getenv(EnvSoonMode)
	SetMode(mode)
}

// SetMode sets soon mode according to input string.
func SetMode(value string) {
	value = strings.Trim(value, " ")
	switch value {
	case ReleaseMode, "":
		value = ReleaseMode
		modeCode = releaseCode
	case DebugMode:
		modeCode = debugCode
	case TestMode:
		modeCode = testCode
	default:
		panic("unknown mode: " + value)
	}
	mode = value
}

// Mode returns currently soon mode.
func Mode() string {
	return mode
}

// DisableBindValidation closes the default validator.
func DisableBindValidation() {
	binding.Validator = nil
}

// EnableJsonDecoderUseNumber sets true for binding.EnableDecoderUseNumber to
// call the UseNumber method on the JSON Decoder instance.
func EnableJsonDecoderUseNumber() {
	binding.EnableDecoderUseNumber = true
}

// EnableJsonDecoderDisallowUnknownFields sets true for binding.EnableDecoderDisallowUnknownFields
// to call the DisallowUnknownFields method on the JSON Decoder instance.
func EnableJsonDecoderDisallowUnknownFields() {
	binding.EnableDecoderDisallowUnknownFields = true
}
