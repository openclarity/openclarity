// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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
	runnerconfig "github.com/openclarity/vmclarity/scanner/families/plugins/runner/config"
)

type Config struct {
	Enabled              bool               `yaml:"enabled" mapstructure:"enabled"`
	ScannersList         []string           `yaml:"scanners_list" mapstructure:"scanners_list"`
	Inputs               []common.ScanInput `yaml:"inputs" mapstructure:"inputs"`
	ScannersConfig       ScannersConfig     `yaml:"scanners_config" mapstructure:"scanners_config"`
	BinaryMode           bool               `yaml:"binary_mode" mapstructure:"binary_mode"`
	BinaryArtifactsPath  string             `yaml:"binary_artifacts_path" mapstructure:"binary_artifacts_path"`
	BinaryArtifactsClean bool               `yaml:"binary_artifacts_clean" mapstructure:"binary_artifacts_clean"`
}

type ScannersConfig map[string]runnerconfig.Config
