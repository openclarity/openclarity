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

package rootkits

import (
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits/common"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
)

type Config struct {
	Enabled         bool                   `yaml:"enabled" mapstructure:"enabled"`
	ScannersList    []string               `yaml:"scanners_list" mapstructure:"scanners_list"`
	StripInputPaths bool                   `yaml:"strip_input_paths" mapstructure:"strip_input_paths"`
	Inputs          []types.Input          `yaml:"inputs" mapstructure:"inputs"`
	ScannersConfig  *common.ScannersConfig `yaml:"scanners_config" mapstructure:"scanners_config"`
}

type Input struct {
	// StripPathFromResult overrides global StripInputPaths value
	StripPathFromResult *bool  `yaml:"strip_path_from_result" mapstructure:"strip_path_from_result"`
	Input               string `yaml:"input" mapstructure:"input"`
	InputType           string `yaml:"input_type" mapstructure:"input_type"`
}
