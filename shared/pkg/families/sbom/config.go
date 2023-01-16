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

package sbom

import "github.com/openclarity/kubeclarity/shared/pkg/config"

type Config struct {
	Enabled         bool           `yaml:"enabled" mapstructure:"enabled"`
	AnalyzersList   []string       `yaml:"analyzers_list" mapstructure:"analyzers_list"`
	Inputs          []Input        `yaml:"inputs" mapstructure:"inputs"`
	MergeWith       []MergeWith    `yaml:"merge_with" mapstructure:"merge_with"`
	AnalyzersConfig *config.Config `yaml:"analyzers_config" mapstructure:"analyzers_config"`
}

type Input struct {
	Input     string `yaml:"input" mapstructure:"input"`
	InputType string `yaml:"input_type" mapstructure:"input_type"`
}

type MergeWith struct {
	SbomPath string `yaml:"sbom_path" mapstructure:"sbom_path"`
}
