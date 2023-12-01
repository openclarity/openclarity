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

package helm

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"

	"github.com/openclarity/vmclarity/installation"
)

func LoadPathOrEmbedded(path string) (*chart.Chart, error) {
	if path != "" {
		return LoadPath(path)
	}

	return LoadFS(installation.HelmManifestBundle)
}

func LoadPath(path string) (*chart.Chart, error) {
	chartLoader, err := loader.Loader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create Chart loader: %w", err)
	}

	loadedChart, err := chartLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load Chart: %w", err)
	}

	return loadedChart, nil
}

func LoadFS(fsys fs.FS) (*chart.Chart, error) {
	if fsys == nil {
		return nil, errors.New("failed to load chart: invalid filesystem")
	}

	bufferedFiles := make([]*loader.BufferedFile, 0)
	fsLoader := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d != nil && d.IsDir() {
			return nil
		}

		// Helm expects unix / separator, but on windows this will be \
		name := strings.ReplaceAll(path, string(filepath.Separator), "/")

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		bufferedFiles = append(bufferedFiles, &loader.BufferedFile{
			Name: name,
			Data: data,
		})

		return nil
	}

	err := fs.WalkDir(fsys, ".", fsLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load files from bundle: %w", err)
	}

	loadedChart, err := loader.LoadFiles(bufferedFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to load Chart: %w", err)
	}

	return loadedChart, nil
}
