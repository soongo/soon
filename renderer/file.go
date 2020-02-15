// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package renderer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/soongo/soon/util"
)

type DotfilesPolicy uint8

const (
	DotfilesPolicyIgnore DotfilesPolicy = iota
	DotfilesPolicyAllow
	DotfilesPolicyDeny
)

var (
	ErrIsDir     = errors.New("file is directory")
	ErrForbidden = errors.New(http.StatusText(http.StatusForbidden))
	ErrNotFound  = errors.New(http.StatusText(http.StatusNotFound))
)

type FileOptions struct {
	// Sets the max-age property of the Cache-Control header.
	MaxAge *time.Duration

	// Root directory for relative filenames.
	Root string

	// filename for download
	Name string

	// Whether sets the Last-Modified header to the last modified date of the
	// file on the OS. Set true to disable it.
	LastModifiedDisabled bool

	// HTTP headers to serve with the file.
	Header map[string]string

	// Option for serving dotfiles. Possible values are “DotfilesPolicyAllow”,
	// “DotfilesPolicyDeny”, “DotfilesPolicyIgnore”.
	DotfilesPolicy DotfilesPolicy

	// Enable or disable accepting ranged requests. Set true to disable it.
	AcceptRangesDisabled bool
}

type File struct {
	FilePath string
	Options  *FileOptions
}

// RenderHeader writes custom headers.
func (f File) RenderHeader(w http.ResponseWriter) {

}

// Renderer writes data with custom ContentType.
func (f File) Render(w http.ResponseWriter) error {
	filePath := strings.Trim(f.FilePath, " ")
	if filePath == "" {
		return errors.New("path argument is required")
	}

	options := f.Options
	if options == nil {
		options = &FileOptions{}
	}

	root := strings.Trim(options.Root, " ")
	if root == "" && !filepath.IsAbs(filePath) {
		return errors.New("path must be absolute or specify root")
	}

	absPath := util.EncodeURI(filepath.Join(root, filePath))
	fileInfo, err := os.Stat(absPath)

	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return ErrIsDir
	}

	if strings.HasPrefix(filepath.Base(absPath), ".") {
		if options.DotfilesPolicy == DotfilesPolicyIgnore {
			return ErrNotFound
		}
		if options.DotfilesPolicy == DotfilesPolicyDeny {
			return ErrForbidden
		}
	}

	if options.Header != nil {
		util.SetHeader(w, options.Header)
	}

	if options.MaxAge != nil {
		t := fmt.Sprintf("max-age=%.0f", (*options.MaxAge).Seconds())
		w.Header().Set("Cache-Control", t)
	}

	if !options.LastModifiedDisabled {
		w.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))
	}

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	util.SetContentType(w, filepath.Ext(filePath))
	_, err = io.Copy(w, file)
	return err
}
