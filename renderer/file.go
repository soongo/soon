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

	"github.com/soongo/soon/internal"
	"github.com/soongo/soon/util"
)

// DotfilesPolicy is the policy for dot files
type DotfilesPolicy uint8

const (
	// DotfilesPolicyIgnore ignore dot files
	DotfilesPolicyIgnore DotfilesPolicy = iota

	// DotfilesPolicyAllow allow dot files
	DotfilesPolicyAllow

	// DotfilesPolicyDeny deny dot files
	DotfilesPolicyDeny

	// IndexDisabled disable disable directory indexing
	IndexDisabled string = "IndexDisabled"
)

var (
	// ErrIsDir means the file is directory
	ErrIsDir = internal.NewStatusTextError(400, "file is directory")

	// RangeNotSatisfiableError indicates the range header is not satisfiable
	RangeNotSatisfiableError = internal.NewStatusTextError(400, "range not satisfiable")
)

// FileOptions contains all options for file renderer
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

	// Index sends the specified directory index file.
	// Set to `IndexDisabled` to disable directory indexing.
	Index string
}

// File contains the given path and options for file renderer.
type File struct {
	FilePath string
	Options  FileOptions
}

// RenderHeader writes custom headers.
func (f *File) RenderHeader(_ http.ResponseWriter, _ *http.Request) {
	// pass
}

// Render writes data with custom ContentType.
func (f *File) Render(w http.ResponseWriter, req *http.Request) error {
	absPath, options := strings.TrimSpace(f.FilePath), f.Options
	if absPath == "" {
		return errors.New("path argument is required")
	}

	if !filepath.IsAbs(absPath) {
		root := strings.TrimSpace(options.Root)
		if root == "" {
			return errors.New("path must be absolute or specify root")
		} else if !filepath.IsAbs(root) {
			return errors.New("root must be absolute")
		}

		absPath = filepath.Join(root, absPath)
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return internal.ErrNotFound
	}

	if fileInfo.IsDir() {
		if options.Index == IndexDisabled {
			return ErrIsDir
		}

		index := strings.TrimSpace(options.Index)
		if index == "" {
			index = "index.html"
		}

		absPath = filepath.Join(absPath, index)
		fileInfo, err = os.Stat(absPath)
		if err != nil {
			return internal.ErrNotFound
		}
	}

	if strings.HasPrefix(filepath.Base(absPath), ".") {
		if options.DotfilesPolicy == DotfilesPolicyIgnore {
			return internal.ErrNotFound
		}
		if options.DotfilesPolicy == DotfilesPolicyDeny {
			return internal.ErrForbidden
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
	if err == nil {
		defer file.Close()

		util.SetContentType(w, filepath.Ext(absPath))
		if !options.AcceptRangesDisabled {
			rangeHeader := strings.TrimSpace(req.Header.Get("range"))
			if rangeHeader != "" {
				ranges, err := util.RangeParser(fileInfo.Size(), rangeHeader, true)
				if err != nil {
					return RangeNotSatisfiableError
				}
				if ranges.Type == "bytes" && len(ranges.Ranges) == 1 {
					start, end := ranges.Ranges[0].Start, ranges.Ranges[0].End
					file.Seek(start, 0)
					_, err = io.CopyN(w, file, end-start+1)
					return err
				}
			}
		}

		_, err = io.Copy(w, file)
	}

	return err
}
