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
	"github.com/openclarity/vmclarity/scanner/common"
	syft "github.com/openclarity/vmclarity/scanner/families/sbom/syft/config"
	trivy "github.com/openclarity/vmclarity/scanner/families/sbom/trivy/config"
)

const (
	DefaultOutputFormat = "cyclonedx-json"
)

type MergeWith struct {
	SbomPath string `yaml:"sbom_path" mapstructure:"sbom_path"`
}

type Config struct {
	Enabled         bool               `yaml:"enabled" mapstructure:"enabled"`
	AnalyzersList   []string           `yaml:"analyzers_list" mapstructure:"analyzers_list"`
	Inputs          []common.ScanInput `yaml:"inputs" mapstructure:"inputs"`
	MergeWith       []MergeWith        `yaml:"merge_with" mapstructure:"merge_with"`
	Registry        *common.Registry   `yaml:"registry" mapstructure:"registry"`
	LocalImageScan  bool               `yaml:"local_image_scan" mapstructure:"local_image_scan"`
	OutputFormat    string             `yaml:"output_format" mapstructure:"output_format"`
	AnalyzersConfig AnalyzersConfig    `yaml:"analyzers_config" mapstructure:"analyzers_config"`
}

func (c *Config) GetOutputFormat() string {
	if c.OutputFormat != "" {
		return c.OutputFormat
	}

	return DefaultOutputFormat
}

type AnalyzersConfig struct {
	Syft  syft.Config  `yaml:"syft" mapstructure:"syft"`
	Trivy trivy.Config `yaml:"trivy" mapstructure:"trivy"`
}
