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

package types

import (
	"time"

	cliconfig "github.com/openclarity/vmclarity/scanner/config"
	"github.com/openclarity/vmclarity/scanner/families/types"
)

type Config struct {
	Enabled         bool           `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	ScannersList    []string       `yaml:"scanners_list" mapstructure:"scanners_list"`
	StripInputPaths bool           `yaml:"strip_input_paths" mapstructure:"strip_input_paths"`
	Inputs          []types.Input  `yaml:"inputs" mapstructure:"inputs"`
	ScannersConfig  ScannersConfig `yaml:"scanners_config" mapstructure:"scanners_config"`
}

// Add scanner specific configurations here, where the key is the scanner name,
// and the value is the scanner specific configuration.
//
// For example if the scanner name is "lynis":
//
//	Lynis LynisConfig `yaml:"lynis" mapstructure:"lynis"`
type ScannersConfig struct {
	Lynis     LynisConfig     `yaml:"lynis" mapstructure:"lynis"`
	CISDocker CISDockerConfig `yaml:"cisdocker" mapstructure:"cisdocker"`
}

func (ScannersConfig) IsConfig() {}

type LynisConfig struct {
	BinaryPath string `yaml:"binary_path" mapstructure:"binary_path"`
}

// nolint:musttag
type CISDockerConfig struct {
	Timeout  time.Duration       `yaml:"timeout" mapstructure:"timeout"`
	Registry *cliconfig.Registry `yaml:"registry" mapstructure:"registry"`
}
