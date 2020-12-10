// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/soongo/soon/internal"
	"github.com/soongo/soon/util"

	"github.com/stretchr/testify/assert"
)

func TestFile_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := File{"", FileOptions{}}
	renderer.RenderHeader(w, nil)
	assert.Equal(t, "", w.Header().Get("Content-Type"))
}

func TestFile_Render(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	maxAge := time.Hour
	tests := []struct {
		name                string
		filePath            string
		options             FileOptions
		rangeHeader         string
		expectedRange       *util.Range
		expectedStatus      int
		expectedContentType string
		expectedError       error
	}{
		{
			"normal-1",
			path.Join(pwd, "../README.md"),
			FileOptions{},
			"",
			nil,
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"normal-2",
			path.Join(pwd, "../README.md"),
			FileOptions{
				MaxAge: &maxAge,
				Header: map[string]string{
					"Accept-Charset":  "utf-8",
					"Accept-Language": "en;q=0.5, zh;q=0.8",
				},
			},
			"",
			nil,
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{"empty-filepath", "", FileOptions{}, "", nil, 200, "", errors.New("")},
		{
			"non-exist-file",
			path.Join(pwd, "../xx.md"),
			FileOptions{},
			"",
			nil,
			200,
			"",
			errors.New(""),
		},
		{
			"abs-path-will-ignore-root-specified",
			path.Join(pwd, "../README.md"),
			FileOptions{Root: "static"},
			"",
			nil,
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"with-abs-root-path",
			"../README.md",
			FileOptions{Root: pwd, LastModifiedDisabled: true},
			"",
			nil,
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"with-non-abs-root-path",
			"../README.md",
			FileOptions{Root: ".", LastModifiedDisabled: true},
			"",
			nil,
			200,
			"",
			errors.New(""),
		},
		{"not-root-filepath", "../README.md", FileOptions{}, "", nil, 200, "", errors.New("")},
		{
			"index-not-exists",
			pwd,
			FileOptions{},
			"",
			nil,
			200,
			"",
			errors.New(""),
		},
		{
			"directory",
			pwd,
			FileOptions{Index: IndexDisabled},
			"",
			nil,
			400,
			"",
			ErrIsDir,
		},
		{
			"index-exists",
			pwd,
			FileOptions{Index: "file_test.go"},
			"",
			nil,
			200,
			"application/octet-stream",
			nil,
		},
		{
			"hidden-default",
			path.Join(pwd, "../.testkeep.yml"),
			FileOptions{},
			"",
			nil,
			404,
			"",
			internal.ErrNotFound,
		},
		{
			"hidden-allow",
			path.Join(pwd, "../.testkeep.yml"),
			FileOptions{DotfilesPolicy: DotfilesPolicyAllow},
			"",
			nil,
			200,
			"text/yaml; charset=utf-8",
			nil,
		},
		{
			"hidden-deny",
			path.Join(pwd, "../.testkeep.yml"),
			FileOptions{DotfilesPolicy: DotfilesPolicyDeny},
			"",
			nil,
			403,
			"",
			internal.ErrForbidden,
		},
		{
			"range-0",
			path.Join(pwd, "../README.md"),
			FileOptions{},
			"bytes=10-20",
			&util.Range{Start: 10, End: 20},
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"range-1",
			path.Join(pwd, "../README.md"),
			FileOptions{},
			"bytes=10-20,21-30",
			&util.Range{Start: 10, End: 30},
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"range-2",
			path.Join(pwd, "../README.md"),
			FileOptions{},
			"bytes=10-20,30-50",
			nil,
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"range-3",
			path.Join(pwd, "../README.md"),
			FileOptions{AcceptRangesDisabled: true},
			"bytes=10-20",
			nil,
			200,
			"text/markdown; charset=utf-8",
			nil,
		},
		{
			"range-error",
			path.Join(pwd, "../README.md"),
			FileOptions{},
			"bytes=",
			nil,
			400,
			"text/markdown; charset=utf-8",
			RangeNotSatisfiableError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			renderer := File{tt.filePath, tt.options}
			w, req := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
			if tt.rangeHeader != "" {
				req.Header.Set("range", tt.rangeHeader)
			}
			err = renderer.Render(w, req)
			if tt.expectedError != nil {
				assert.NotNil(err)
				if tt.expectedError.Error() != "" {
					assert.Equal(tt.expectedError, err)
				}
				if httpErr, ok := tt.expectedError.(internal.HttpError); ok {
					assert.Equal(tt.expectedStatus, httpErr.Status())
				} else {
					assert.Equal(tt.expectedStatus, w.Code)
				}
				assert.Equal("", w.Body.String())
			} else {
				if tt.options.Index != "" {
					tt.filePath = filepath.Join(tt.filePath, tt.options.Index)
				}
				fileInfo, fileContent := getFileContent(tt.filePath, tt.expectedRange)
				lastModified := fileInfo.ModTime().UTC().Format(timeFormat)
				assert.Equal(tt.expectedStatus, w.Code)
				assert.Equal(fileContent, w.Body.String())
				assert.Equal(tt.expectedContentType, w.Header().Get("Content-Type"))
				if tt.options.MaxAge != nil {
					cc := fmt.Sprintf("max-age=%.0f", maxAge.Seconds())
					assert.Equal(cc, w.Header().Get("Cache-Control"))
				}
				if tt.options.Header != nil {
					for k, v := range tt.options.Header {
						assert.Equal(v, w.Header().Get(k))
					}
				}
				expectedLastModified := lastModified
				if tt.options.LastModifiedDisabled {
					expectedLastModified = ""
				}
				assert.Equal(expectedLastModified, w.Header().Get("Last-Modified"))
			}
		})
	}
}

func getFileContent(p string, r *util.Range) (os.FileInfo, string) {
	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	fileInfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	start, size := int64(0), fileInfo.Size()
	if r != nil {
		start, size = r.Start, r.End-r.Start+1
	}

	bts := make([]byte, size)
	_, err = f.ReadAt(bts, start)
	if err != nil {
		panic(err)
	}

	return fileInfo, string(bts)
}
