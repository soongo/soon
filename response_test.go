// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http/httptest"
	"testing"
)

func TestResponse_WriteHeader(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		tests := []struct {
			code     int
			expected int
		}{
			{0, 200},
			{100, 100},
			{200, 200},
		}

		for _, tt := range tests {
			r := newResponse(httptest.NewRecorder())
			r.WriteHeader(tt.code)
			if got := r.Status(); got != tt.expected {
				t.Errorf(testErrorFormat, got, tt.expected)
			}
		}
	})

	t.Run("write-after-send", func(t *testing.T) {
		got := captureOutput(t, func() {
			SetMode(DebugMode)
			r := newResponse(httptest.NewRecorder())
			r.WriteHeader(100)
			_, _ = r.WriteString("hello")
			r.WriteHeader(200)
			SetMode(TestMode)
		})
		expected := "[SOON-debug] [WARNING] Headers were already written. " +
			"Wanted to override status code 100 with 200\n"
		if got != expected {
			t.Errorf(testErrorFormat, got, expected)
		}
	})
}

func TestResponse_WriteHeaderNow(t *testing.T) {
	tests := []struct {
		code     int
		again    int
		expected int
	}{
		{0, 0, 200},
		{200, 0, 200},
		{500, 404, 500},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		r.WriteHeader(tt.code)
		r.WriteHeaderNow()
		r.WriteHeader(tt.again)
		r.WriteHeaderNow()
		if got := r.ResponseWriter.(*httptest.ResponseRecorder).Code; got != tt.expected {
			t.Errorf(testErrorFormat, got, tt.expected)
		}
	}
}

func TestResponse_Write(t *testing.T) {
	tests := []struct {
		code         int
		data         []byte
		size         int
		expectedCode int
	}{
		{0, []byte("hello"), 5, 200},
		{302, []byte("你好"), 6, 302},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		r.WriteHeader(tt.code)
		size, err := r.Write(tt.data)
		if err != nil {
			t.Error(err)
		} else {
			if got := r.Status(); got != tt.expectedCode {
				t.Errorf(testErrorFormat, got, tt.expectedCode)
			}
			recorder := r.ResponseWriter.(*httptest.ResponseRecorder)
			if got := recorder.Code; got != tt.expectedCode {
				t.Errorf(testErrorFormat, got, tt.expectedCode)
			}
			if got := recorder.Body.String(); got != string(tt.data) {
				t.Errorf(testErrorFormat, got, tt.data)
			}
			if size != tt.size {
				t.Errorf(testErrorFormat, size, tt.size)
			}
		}
	}
}

func TestResponse_WriteString(t *testing.T) {
	tests := []struct {
		code         int
		data         string
		size         int
		expectedCode int
	}{
		{0, "hello", 5, 200},
		{302, "你好", 6, 302},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		recorder := r.ResponseWriter.(*httptest.ResponseRecorder)
		r.WriteHeader(tt.code)
		if got := recorder.Code; got != 200 {
			t.Errorf(testErrorFormat, got, 200)
		}
		size, err := r.WriteString(tt.data)
		if err != nil {
			t.Error(err)
		} else {
			if got := r.Status(); got != tt.expectedCode {
				t.Errorf(testErrorFormat, got, tt.expectedCode)
			}
			if got := recorder.Code; got != tt.expectedCode {
				t.Errorf(testErrorFormat, got, tt.expectedCode)
			}
			if got := recorder.Body.String(); got != tt.data {
				t.Errorf(testErrorFormat, got, tt.data)
			}
			if size != tt.size {
				t.Errorf(testErrorFormat, size, tt.size)
			}
		}
	}
}

func TestResponse_Flush(t *testing.T) {
	tests := []struct {
		code         int
		expectedCode int
	}{
		{0, 200},
		{302, 302},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		recorder := r.ResponseWriter.(*httptest.ResponseRecorder)
		r.WriteHeader(tt.code)
		if got := r.Status(); got != tt.expectedCode {
			t.Errorf(testErrorFormat, got, tt.expectedCode)
		}
		if got := recorder.Code; got != 200 {
			t.Errorf(testErrorFormat, got, 200)
		}
		if got := recorder.Flushed; got != false {
			t.Errorf(testErrorFormat, got, false)
		}
		r.Flush()
		if got := recorder.Code; got != tt.expectedCode {
			t.Errorf(testErrorFormat, got, tt.expectedCode)
		}
		if got := recorder.Flushed; got != true {
			t.Errorf(testErrorFormat, got, true)
		}
	}
}

func TestResponse_Status(t *testing.T) {
	tests := []struct {
		code         int
		expectedCode int
	}{
		{0, 200},
		{302, 302},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		r.WriteHeader(tt.code)
		if got := r.Status(); got != tt.expectedCode {
			t.Errorf(testErrorFormat, got, tt.expectedCode)
		}
	}
}

func TestResponse_Size(t *testing.T) {
	tests := []struct {
		data string
		size int
	}{
		{"hello", 5},
		{"你好", 6},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		_, err := r.WriteString(tt.data)
		if err != nil {
			t.Error(err)
		} else if r.Size() != tt.size {
			t.Errorf(testErrorFormat, r.Size(), tt.size)
		}
	}
}

func TestResponse_Written(t *testing.T) {
	tests := []struct {
		code int
		data []byte
	}{
		{200, nil},
		{0, []byte("hello")},
	}

	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		r.WriteHeader(tt.code)
		if got := r.Written(); got != false {
			t.Errorf(testErrorFormat, got, false)
		}
		if tt.code > 0 {
			r.WriteHeaderNow()
		} else if tt.data != nil {
			_, _ = r.Write(tt.data)
		}
		if got := r.Written(); got != true {
			t.Errorf(testErrorFormat, got, true)
		}
	}
}
