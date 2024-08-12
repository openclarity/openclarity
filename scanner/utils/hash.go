// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GenerateHash(inputType SourceType, source string) (string, error) {
	absPath, err := filepath.Abs(source)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of the source %s: %w", source, err)
	}
	switch inputType {
	case IMAGE, DOCKERARCHIVE, OCIARCHIVE, OCIDIR:
		log.Infof("Skip generating hash in the case of image")
		return "", nil
	case DIR, ROOTFS:
		hash, err := hashDir(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to create hash for directory %s: %w", absPath, err)
		}
		return hash, nil
	case FILE:
		input, err := os.Open(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s for generating hash: %w", absPath, err)
		}
		defer input.Close()
		hash := sha256.New()
		if _, err := io.Copy(hash, input); err != nil {
			return "", fmt.Errorf("failed to create hash for file %s: %w", absPath, err)
		}
		return fmt.Sprintf("%x", hash.Sum(nil)), nil // nolint:perfsprint
	case SBOM:
		log.Infof("Skip generating hash in the case of sbom")
		return "", nil
	default:
		return "", fmt.Errorf("unsupported input type %s", inputType)
	}
}

// hashDir gathers files from directory and
// pass the file list and the open function as arguments to the generateHash function,
// which creates hashes for all files and generates a hash for the hashes.
func hashDir(dir string) (string, error) {
	files, err := dirFiles(dir)
	if err != nil {
		return "", fmt.Errorf("failed to gathering files from directory %s: %w", dir, err)
	}
	osOpen := func(name string) (io.ReadCloser, error) {
		return os.Open(filepath.Join(dir, name)) // nolint:wrapcheck
	}
	return generateHash(files, osOpen)
}

// dirFiles gathering files from a directory.
func dirFiles(dir string) ([]string, error) {
	var files []string
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel := file
		if dir != "." {
			rel = file[len(dir)+1:]
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk over files: %w", err)
	}
	return files, nil
}

// generateHash creates hashes for all files along with filenames and generates a hash for the hashes and filenames.
func generateHash(files []string, open func(string) (io.ReadCloser, error)) (string, error) {
	h := sha256.New()
	files = append([]string(nil), files...)
	sort.Strings(files)
	for _, file := range files {
		if strings.Contains(file, "\n") {
			return "", errors.New("filenames with newlines are not supported")
		}
		r, err := open(file)
		if err != nil {
			return "", fmt.Errorf("failed to open file %s: %w", file, err)
		}
		hf := sha256.New()
		_, err = io.Copy(hf, r)
		r.Close()
		if err != nil {
			return "", fmt.Errorf("failed to create hash for file %s: %w", file, err)
		}
		fmt.Fprintf(h, "%x  %s\n", hf.Sum(nil), file)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil // nolint:perfsprint
}
