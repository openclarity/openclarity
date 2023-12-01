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

package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"

	"github.com/openclarity/vmclarity/e2e/testenv/docker/asset"
	"github.com/openclarity/vmclarity/e2e/testenv/utils"
	"github.com/openclarity/vmclarity/installation"
)

func ProjectFromConfig(config *Config) (*types.Project, error) {
	if config.ctx == nil {
		config.ctx = context.Background()
	}

	if config.ComposeFiles == nil {
		config.ComposeFiles = []string{}
	}

	if len(config.ComposeFiles) == 0 {
		// Unpack VMClarity Docker Bundle
		if err := utils.ExportFS(installation.DockerManifestBundle, config.WorkDir); err != nil {
			return nil, fmt.Errorf("failed to unpack Docker manifest: %w", err)
		}

		composeFiles, err := findComposeFiles(config.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("failed to find Compose files in %s: %w", config.WorkDir, err)
		}

		config.ComposeFiles = composeFiles
	}

	if !config.SkipAssetInstall {
		// Unpack TestAssets Docker Bundle
		assetDir := filepath.Join(config.WorkDir, "asset")
		if err := utils.ExportFS(asset.AssetManifestFS, assetDir); err != nil {
			return nil, fmt.Errorf("failed to unpack Docker test assets manifest: %w", err)
		}

		composeFiles, err := findComposeFiles(assetDir)
		if err != nil {
			return nil, fmt.Errorf("failed to find Compose files in %s: %w", assetDir, err)
		}

		config.ComposeFiles = append(config.ComposeFiles, composeFiles...)
	}

	opts, err := cli.NewProjectOptions(
		config.ComposeFiles,
		cli.WithContext(config.ctx),
		cli.WithName(config.EnvName),
		cli.WithInterpolation(true),
		cli.WithResolvedPaths(true),
		cli.WithWorkingDirectory(config.WorkDir),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project options: %w", err)
	}

	WithContainerImages(opts, config.Images)

	project, err := cli.ProjectFromOptions(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create compose project: %w", err)
	}

	for i, service := range project.Services {
		service.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     service.Name,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False",
		}
		project.Services[i] = service
	}

	return project, nil
}

const (
	APIServerImageEnv    = "VMCLARITY_APISERVER_CONTAINER_IMAGE"
	OrchestratorImageEnv = "VMCLARITY_ORCHESTRATOR_CONTAINER_IMAGE"
	UIImageEnv           = "VMCLARITY_UI_CONTAINER_IMAGE"
	UIBackendImageEnv    = "VMCLARITY_UIBACKEND_CONTAINER_IMAGE"
	ScannerImageEnv      = "VMCLARITY_SCANNER_CONTAINER_IMAGE"
)

func WithContainerImages(o *cli.ProjectOptions, images ContainerImages) {
	o.Environment[APIServerImageEnv] = images.APIServer
	o.Environment[OrchestratorImageEnv] = images.Orchestrator
	o.Environment[UIImageEnv] = images.UI
	o.Environment[UIBackendImageEnv] = images.UIBackend
	o.Environment[ScannerImageEnv] = images.Scanner
}

func findComposeFiles(path string) ([]string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %w", err)
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid directory path: %w", err)
	}

	composeFiles := []string{}
	for _, compose := range cli.DefaultFileNames {
		f := filepath.Join(path, compose)
		if _, err = os.Stat(f); err == nil {
			composeFiles = append(composeFiles, f)
		}
	}

	for _, override := range cli.DefaultOverrideFileNames {
		f := filepath.Join(path, override)
		if _, err = os.Stat(f); err == nil {
			composeFiles = append(composeFiles, f)
		}
	}

	return composeFiles, nil
}
