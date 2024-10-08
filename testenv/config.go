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

package testenv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	awsenv "github.com/openclarity/openclarity/testenv/aws"
	azureenv "github.com/openclarity/openclarity/testenv/azure"
	dockerenv "github.com/openclarity/openclarity/testenv/docker"
	gcpenv "github.com/openclarity/openclarity/testenv/gcp"
	k8senv "github.com/openclarity/openclarity/testenv/kubernetes"
	"github.com/openclarity/openclarity/testenv/types"
)

const (
	DefaultEnvName  = "testenv"
	DefaultPlatform = types.EnvironmentTypeDocker
)

const (
	DefaultAPIServer         = "ghcr.io/openclarity/openclarity-api-server:latest"
	DefaultOrchestrator      = "ghcr.io/openclarity/openclarity-orchestrator:latest"
	DefaultUI                = "ghcr.io/openclarity/openclarity-ui:latest"
	DefaultUIBackend         = "ghcr.io/openclarity/openclarity-ui-backend:latest"
	DefaultScanner           = "ghcr.io/openclarity/openclarity-cli:latest"
	DefaultCRDiscoveryServer = "ghcr.io/openclarity/openclarity-cr-discovery-server:latest"
	DefaultPluginKics        = "ghcr.io/openclarity/openclarity-plugin-kics:latest"
)

// Config is the configuration for testenv.
//
//nolint:containedctx
type Config struct {
	// Platform defines the platform to be used for running end-to-end test suite.
	Platform types.EnvironmentType `mapstructure:"platform"`
	// EnvName the name of the environment to be created.
	EnvName string `mapstructure:"env_name"`
	// Images provides a list of container images used for deployment.
	Images types.ContainerImages[string] `mapstructure:",squash"`
	// Docker contains the configuration for Docker platform.
	Docker *dockerenv.Config `mapstructure:"docker,omitempty"`
	// Kubernetes contains the configuration for Kubernetes platform.
	Kubernetes *k8senv.Config `mapstructure:"kubernetes,omitempty"`
	// AWS contains the configuration for AWS platform.
	AWS *awsenv.Config `mapstructure:"aws,omitempty"`
	// GCP contains the configuration for GCP platform.
	GCP *gcpenv.Config `mapstructure:"gcp,omitempty"`
	// Azure contains the configuration for Azure platform.
	Azure *azureenv.Config `mapstructure:"azure,omitempty"`
	// WorkDir contains the path to the work directory.
	WorkDir string `mapstructure:"work_dir"`

	// logger is an initialized logrus.Entry which is used by some components during initialization,
	// where context.Context is not available to retrieve logger from.
	logger *logrus.Entry
	// ctx provides context.Context for environment which require it at initialization time (e.g. Docker Compose).
	ctx context.Context
}

// ConfigOptFn defines transformer function for Config.
type ConfigOptFn func(*Config) error

var applyConfigWithOpts = types.WithOpts[Config, ConfigOptFn]

// WithLogger sets logger for Config.
func WithLogger(logger *logrus.Entry) ConfigOptFn {
	return func(config *Config) error {
		config.logger = logger

		return nil
	}
}

// WithContext sets context for Config.
func WithContext(ctx context.Context) ConfigOptFn {
	return func(config *Config) error {
		config.ctx = ctx

		return nil
	}
}

// WithWorkDir sets Config.WorkDir in config.
func WithWorkDir(dir string) ConfigOptFn {
	return func(config *Config) error {
		config.WorkDir = dir

		return nil
	}
}

// withResolvedWorkDirPath returns a ConfigOptFn which validates the Config.WorkDir parameter and returns error
// if it is invalid.
func withResolvedWorkDirPath() ConfigOptFn {
	return func(config *Config) error {
		if config.WorkDir != "" {
			absDir, err := filepath.Abs(config.WorkDir)
			if err != nil {
				return fmt.Errorf("invalid work directory path %s: %w", config.WorkDir, err)
			}

			stat, err := os.Stat(absDir)
			if err != nil {
				return fmt.Errorf("invalid work directory path %s: %w", config.WorkDir, err)
			}

			if !stat.IsDir() {
				return fmt.Errorf("invalid path %s: not a directory", config.WorkDir)
			}

			config.WorkDir = absDir
		}

		return nil
	}
}

// withDefaultWorkDir returns a ConfigOptFn which creates and sets a work directory in case
// none is provided in Config.WorkDir.
func withDefaultWorkDir() ConfigOptFn {
	return func(config *Config) error {
		if config.WorkDir != "" {
			return nil
		}

		userConfDir, err := ConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get user's config directory path: %w", err)
		}

		config.WorkDir = filepath.Join(userConfDir, config.EnvName)

		var perm os.FileMode = 0o755
		if err = os.MkdirAll(config.WorkDir, perm); err != nil {
			return fmt.Errorf("failed to create work directory at %s: %w", config.WorkDir, err)
		}

		return nil
	}
}

const TestEnvDataDir = "openclarity-testenv"

func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, TestEnvDataDir), nil
	}

	if dir := os.Getenv("AppData"); runtime.GOOS == "windows" && dir != "" {
		return filepath.Join(dir, TestEnvDataDir), nil
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %w", err)
	}

	return filepath.Join(dir, ".config", TestEnvDataDir), nil
}
