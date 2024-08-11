// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func ExportFS(fsys fs.ReadDirFS, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	info, err := os.Stat(absPath)
	switch {
	case os.IsNotExist(err):
		var dirPerm fs.FileMode = 0o700
		if err = os.MkdirAll(absPath, dirPerm); err != nil {
			return fmt.Errorf("failed to create directory %s,: %w", absPath, err)
		}
	case err != nil:
		return fmt.Errorf("invalid path: %w", err)
	case !info.IsDir():
		return fmt.Errorf("invalid path: not a directory: %s", absPath)
	}

	if err = fs.WalkDir(fsys, ".", newExporter(fsys, absPath)); err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	return nil
}

func newExporter(fsys fs.ReadDirFS, dir string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d != nil && d.IsDir() {
			// Get original file permission if available
			var dirPerm fs.FileMode = 0o700
			info, err := d.Info()
			if err == nil {
				dirPerm |= info.Mode()
			}

			// Create directory
			if err = os.MkdirAll(filepath.Join(dir, path), dirPerm); err != nil {
				return fmt.Errorf("failed to created directory: %w", err)
			}

			return nil
		}

		// Open original file from fsys
		src, err := fsys.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}
		defer src.Close()

		// Open destination file
		dest, err := os.Create(filepath.Join(dir, path))
		if err != nil {
			return fmt.Errorf("failed to open destination file: %w", err)
		}
		defer dest.Close()

		// Copy content
		_, err = io.Copy(bufio.NewWriter(dest), bufio.NewReader(src))
		if err != nil {
			return fmt.Errorf("failed to cpy content: %w", err)
		}

		// Get original file permission if available
		var filePerm fs.FileMode = 0o600
		if d != nil {
			info, err := d.Info()
			if err == nil {
				filePerm |= info.Mode()
			}
		}

		// Set file permission
		if err = os.Chmod(filepath.Join(dir, path), filePerm); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}

		return nil
	}
}
