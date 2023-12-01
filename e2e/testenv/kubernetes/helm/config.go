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
	"fmt"
	"path/filepath"

	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

const (
	// https://helm.sh/docs/topics/advanced/#storage-backends
	DefaultHelmDriver    = "secret"
	DefaultHelmNamespace = "default"
)

var applyConfigWithOpts = envtypes.WithOpts[Config, ConfigOptFn]

type ContainerImages = envtypes.ContainerImages[envtypes.ImageRef]

type Config struct {
	Namespace      string `mapstructure:"namespace"`
	ReleaseName    string `mapstructure:"release_name"`
	ChartPath      string `mapstructure:"chart_path"`
	StorageDriver  string `mapstructure:"driver"`
	KubeConfigPath string `mapstructure:"kubeconfig"`

	valuesOpts []ValuesOpts
}

// ConfigOptFn defines transformer function for Config.
type ConfigOptFn func(*Config) error

func WithNamespace(ns string) ConfigOptFn {
	return func(c *Config) error {
		if c.Namespace != "" {
			return nil
		}
		c.Namespace = ns

		return nil
	}
}

func WithKubeConfigPath(path string) ConfigOptFn {
	return func(c *Config) error {
		if c.KubeConfigPath != "" {
			return nil
		}
		c.KubeConfigPath = path

		return nil
	}
}

func WithValuesOpts(opts []ValuesOpts) ConfigOptFn {
	return func(c *Config) error {
		if c.valuesOpts == nil {
			c.valuesOpts = make([]ValuesOpts, 0, len(opts))
		}
		c.valuesOpts = append(c.valuesOpts, opts...)

		return nil
	}
}

func withResolvedChartPath() ConfigOptFn {
	return func(config *Config) error {
		if config.ChartPath == "" {
			return nil
		}

		p, err := filepath.Abs(config.ChartPath)
		if err != nil {
			return fmt.Errorf("failed to resolve Chart directory: %s: %w", config.ChartPath, err)
		}
		config.ChartPath = p

		return nil
	}
}

func withResolvedKubeConfigPath() ConfigOptFn {
	return func(config *Config) error {
		p, err := filepath.Abs(config.KubeConfigPath)
		if err != nil {
			return fmt.Errorf("failed to resolve KubeConfig: %s: %w", config.KubeConfigPath, err)
		}
		config.KubeConfigPath = p

		return nil
	}
}
