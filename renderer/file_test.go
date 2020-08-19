// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFile_RenderHeader(t *testing.T) {
	w := httptest.NewRecorder()
	renderer := File{"", nil}
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
		options             *FileOptions
		expectedStatus      int
		expectedContentType string
		expectedError       error
	}{
		{
			"normal-1",
			path.Join(pwd, "../README.md"),
			nil,
			200,
			"text/markdown; charset=UTF-8",
			nil,
		},
		{
			"normal-2",
			path.Join(pwd, "../README.md"),
			&FileOptions{
				MaxAge: &maxAge,
				Header: map[string]string{
					"Accept-Charset":  "utf-8",
					"Accept-Language": "en;q=0.5, zh;q=0.8",
				},
			},
			200,
			"text/markdown; charset=UTF-8",
			nil,
		},
		{"empty-filepath", "", nil, 200, "", errors.New("")},
		{
			"non-exist-file",
			path.Join(pwd, "../xx.md"),
			nil,
			200,
			"",
			errors.New(""),
		},
		{
			"with-root-path",
			"../README.md",
			&FileOptions{Root: pwd, LastModifiedDisabled: true},
			200,
			"text/markdown; charset=UTF-8",
			nil,
		},
		{"not-root-filepath", "../README.md", nil, 200, "", errors.New("")},
		{"directory", pwd, nil, 200, "", ErrIsDir},
		{
			"hidden-default",
			path.Join(pwd, "../.travis.yml"),
			nil,
			200,
			"",
			ErrNotFound,
		},
		{
			"hidden-allow",
			path.Join(pwd, "../.travis.yml"),
			&FileOptions{DotfilesPolicy: DotfilesPolicyAllow},
			200,
			"text/yaml; charset=UTF-8",
			nil,
		},
		{
			"hidden-deny",
			path.Join(pwd, "../.travis.yml"),
			&FileOptions{DotfilesPolicy: DotfilesPolicyDeny},
			200,
			"",
			ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			renderer := File{tt.filePath, tt.options}
			w := httptest.NewRecorder()
			err := renderer.Render(w, nil)
			if tt.expectedError != nil {
				assert.NotNil(err)
				assert.Equal(tt.expectedStatus, w.Code)
				assert.Equal("", w.Body.String())
			} else {
				fileInfo, fileContent := getFileContent(tt.filePath)
				lastModified := fileInfo.ModTime().UTC().Format(timeFormat)
				assert.Equal(tt.expectedStatus, w.Code)
				assert.Equal(fileContent, w.Body.String())
				assert.Equal(tt.expectedContentType, w.Header().Get("Content-Type"))
				if tt.options != nil {
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
				} else {
					assert.Equal(lastModified, w.Header().Get("Last-Modified"))
				}
			}
		})
	}
}

func getFileContent(p string) (os.FileInfo, string) {
	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		panic(err)
	}

	return fileInfo, string(bts)
}
