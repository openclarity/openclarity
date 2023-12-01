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

package kubernetes

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes/helm"
	"github.com/openclarity/vmclarity/e2e/testenv/kubernetes/types"
	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

const (
	DefaultProvider               = types.ProviderTypeKind
	DefaultInstaller              = types.InstallerTypeHelm
	DefaultClusterCreationTimeout = 30 * time.Minute
	DefaultKubernetesVersion      = "1.28"
	DefaultNamespace              = "default"
)

var applyConfigWithOpts = envtypes.WithOpts[Config, ConfigOptFn]

type ContainerImages = envtypes.ContainerImages[envtypes.ImageRef]

type Config struct {
	// EnvName the name of the environment to be created
	EnvName string `mapstructure:"env_name"`
	// Namespace where to deploy
	Namespace string `mapstructure:"namespace"`
	// Provider defines the infrastructure/platform provider
	Provider types.ProviderType `mapstructure:"provider"`
	// Installer defines the installer to be used for deployment
	Installer types.InstallerType `mapstructure:"installer"`
	// HelmConfig defines the Helm specific configuration
	HelmConfig *helm.Config `mapstructure:"helm,omitempty"`
	// Images contains the list of container images to be used for deployment
	Images ContainerImages `mapstructure:",squash"`
	// WorkDir absolute path to the directory where the deployment files prior performing actions
	WorkDir string `mapstructure:"work_dir"`
	// SkipAssetInstall defines whether to deploy test assets to environment or not
	SkipAssetInstall bool `mapstructure:"skip_asset_install"`

	// ProviderConfig contains the configuration for the Kubernetes provider.
	// NOTE(chrisgacsal): mapstructure does not support *squash* for ptr types
	types.ProviderConfig `mapstructure:",squash"`

	// logger is an initialized logrus.Entry which is used by some components during initialization,
	// where context.Context is not available to retrieve logger from.
	logger *logrus.Entry
}

// ConfigOptFn defines transformer function for Config.
type ConfigOptFn func(*Config) error

func WithLogger(logger *logrus.Entry) ConfigOptFn {
	return func(config *Config) error {
		config.logger = logger

		return nil
	}
}

// WithWorkDir set workDir for Config.
func WithWorkDir(dir string) ConfigOptFn {
	return func(config *Config) error {
		config.WorkDir = dir

		return nil
	}
}

func withResolvedKubeConfigPath() ConfigOptFn {
	return func(config *Config) error {
		if filepath.IsAbs(config.KubeConfigPath) {
			return nil
		}

		kubeConfig := config.KubeConfigPath
		if config.KubeConfigPath == "" {
			kubeConfig = "kubeConfig"
		}

		absPath, err := filepath.Abs(filepath.Join(config.WorkDir, kubeConfig))
		if err != nil {
			return fmt.Errorf("invalid KubeConfig path: %w", err)
		}

		config.KubeConfigPath = absPath

		return nil
	}
}
