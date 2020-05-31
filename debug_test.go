// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
	"testing"
)

func TestIsDebugging(t *testing.T) {
	tests := []struct {
		mode     string
		expected bool
	}{
		{DebugMode, true},
		{ReleaseMode, false},
		{TestMode, false},
	}

	for _, tt := range tests {
		SetMode(tt.mode)
		if got := IsDebugging(); got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestDebugPrint(t *testing.T) {
	got := captureOutput(t, func() {
		SetMode(DebugMode)
		SetMode(ReleaseMode)
		debugPrint("DEBUG this!")
		SetMode(TestMode)
		debugPrint("DEBUG this!")
		SetMode(DebugMode)
		debugPrint("these are %d %s", 2, "error messages")
		SetMode(TestMode)
	})
	expected := "[SOON-debug] these are 2 error messages\n"
	if got != expected {
		t.Errorf(testErrorFormat, got, expected)
	}
}

func captureOutput(t *testing.T, f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defaultWriter := DefaultWriter
	defaultErrorWriter := DefaultErrorWriter
	defer func() {
		DefaultWriter = defaultWriter
		DefaultErrorWriter = defaultErrorWriter
		log.SetOutput(os.Stderr)
	}()
	DefaultWriter = writer
	DefaultErrorWriter = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		_, err := io.Copy(&buf, reader)
		if err != nil {
			t.Error(testErrorFormat, err, "nil")
		}
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}
