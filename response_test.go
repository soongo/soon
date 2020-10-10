// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Equal(t, tt.expected, r.Status())
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
		assert.Equal(t, expected, got)
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
		assert.Equal(t, tt.expected, r.ResponseWriter.(*httptest.ResponseRecorder).Code)
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

	assert := assert.New(t)
	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		recorder := r.ResponseWriter.(*httptest.ResponseRecorder)
		r.WriteHeader(tt.code)
		size, err := r.Write(tt.data)
		assert.Nil(err)
		assert.Equal(tt.expectedCode, r.Status())
		assert.Equal(tt.expectedCode, recorder.Code)
		assert.Equal(string(tt.data), recorder.Body.String())
		assert.Equal(tt.size, size)
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

	assert := assert.New(t)
	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		recorder := r.ResponseWriter.(*httptest.ResponseRecorder)
		r.WriteHeader(tt.code)
		assert.Equal(200, recorder.Code)
		size, err := r.WriteString(tt.data)
		assert.Nil(err)
		assert.Equal(tt.expectedCode, r.Status())
		assert.Equal(tt.expectedCode, recorder.Code)
		assert.Equal(tt.data, recorder.Body.String())
		assert.Equal(tt.size, size)
	}
}

func TestResponse_Hijack(t *testing.T) {
	router := NewRouter()
	router.Use(func(c *Context) {
		conn, buf, err := c.Writer.Hijack()
		defer conn.Close()
		require.NoError(t, err)
		assert.NotNil(t, conn)
		assert.NotNil(t, buf)
		_, err = buf.WriteString(body200)
		require.NoError(t, err)
		require.NoError(t, buf.Flush())
	})
	server := httptest.NewServer(router)
	defer server.Close()
	request("GET", server.URL+"/", nil)
}

func TestResponse_Flush(t *testing.T) {
	tests := []struct {
		code         int
		expectedCode int
	}{
		{0, 200},
		{302, 302},
	}

	assert := assert.New(t)
	for _, tt := range tests {
		r := newResponse(httptest.NewRecorder())
		recorder := r.ResponseWriter.(*httptest.ResponseRecorder)
		r.WriteHeader(tt.code)
		assert.Equal(tt.expectedCode, r.Status())
		assert.Equal(200, recorder.Code)
		assert.Equal(false, recorder.Flushed)

		r.Flush()
		assert.Equal(tt.expectedCode, recorder.Code)
		assert.Equal(true, recorder.Flushed)
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
		assert.Equal(t, tt.expectedCode, r.Status())
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
		assert.Nil(t, err)
		assert.Equal(t, tt.size, r.Size())
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
		assert.Equal(t, false, r.Written())
		if tt.code > 0 {
			r.WriteHeaderNow()
		} else if tt.data != nil {
			_, _ = r.Write(tt.data)
		}
		assert.Equal(t, true, r.Written())
	}
}
