// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soongo/soon/renderer"
	"github.com/soongo/soon/util"

	"github.com/stretchr/testify/assert"
)

var initDirnameErr = errors.New("InitDirnameErr")

func TestStatic(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Println(pwd)

	tests := []struct {
		route               string
		root                string
		options             renderer.FileOptions
		path                string
		header              http.Header
		expectedCode        int
		expectedContentType string
		expectedFile        string
		expectedContent     string
		expectedRange       *util.Range
		initDirnameErr      error
	}{
		{
			"",
			pwd,
			renderer.FileOptions{},
			"/README.md",
			nil,
			200,
			"text/markdown; charset=UTF-8",
			filepath.Join(pwd, "README.md"),
			"",
			nil,
			nil,
		},
		{
			"",
			pwd,
			renderer.FileOptions{},
			"/not_exist_file.txt",
			nil,
			404,
			"text/plain; charset=utf-8",
			"",
			"404 page not found",
			nil,
			nil,
		},
		{
			"",
			"./",
			renderer.FileOptions{},
			"/README.md",
			nil,
			404,
			"text/plain; charset=utf-8",
			"",
			"404 page not found",
			nil,
			nil,
		},
		{
			"",
			".",
			renderer.FileOptions{},
			"/README.md",
			nil,
			500,
			"text/plain; charset=utf-8",
			"",
			initDirnameErr.Error(),
			nil,
			initDirnameErr,
		},
		{
			"",
			pwd,
			renderer.FileOptions{},
			"/.travis.yml",
			nil,
			404,
			"text/plain; charset=utf-8",
			"",
			http.StatusText(404),
			nil,
			nil,
		},
		{
			"",
			pwd,
			renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyDeny},
			"/.travis.yml",
			nil,
			403,
			"text/plain; charset=utf-8",
			"",
			http.StatusText(403),
			nil,
			nil,
		},
		{
			"",
			pwd,
			renderer.FileOptions{DotfilesPolicy: renderer.DotfilesPolicyAllow},
			"/.travis.yml",
			nil,
			200,
			"text/yaml; charset=UTF-8",
			filepath.Join(pwd, ".travis.yml"),
			"",
			nil,
			nil,
		},
		{
			"",
			pwd,
			renderer.FileOptions{},
			"/README.md",
			http.Header{"Range": []string{"bytes=10-20,21-30"}},
			200,
			"text/markdown; charset=UTF-8",
			filepath.Join(pwd, "README.md"),
			"",
			&util.Range{Start: 10, End: 30},
			nil,
		},
		{
			"",
			pwd,
			renderer.FileOptions{},
			"/README.md",
			http.Header{"Range": []string{"bytes="}},
			400,
			"text/plain; charset=utf-8",
			"",
			renderer.RangeNotSatisfiableError.Error(),
			nil,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			router := NewRouter()
			if tt.route == "" {
				router.Use(Static(tt.root, tt.options))
			} else {
				router.Use(tt.route, Static(tt.root, tt.options))
			}
			server := httptest.NewServer(router)
			defer server.Close()

			util.InitDirnameErr = tt.initDirnameErr

			code, header, body, err := request("GET", server.URL+tt.path, tt.header)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.expectedContentType, header.Get("content-type"))

			content := tt.expectedContent
			if content == "" {
				_, content = getFileContent(tt.expectedFile, tt.expectedRange)
			}
			assert.Equal(t, strings.TrimSpace(content), strings.TrimSpace(body))
		})
	}
}
