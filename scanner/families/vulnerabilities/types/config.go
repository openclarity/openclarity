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
	"github.com/openclarity/openclarity/scanner/common"
	grypeconfig "github.com/openclarity/openclarity/scanner/families/vulnerabilities/grype/config"
	trivyconfig "github.com/openclarity/openclarity/scanner/families/vulnerabilities/trivy/config"
)

type Config struct {
	Enabled        bool               `yaml:"enabled" mapstructure:"enabled" json:"enabled"`
	ScannersList   []string           `yaml:"scanners_list" mapstructure:"scanners_list" json:"scanners_list"`
	Inputs         []common.ScanInput `yaml:"inputs" mapstructure:"inputs" json:"inputs"`
	InputFromSbom  bool               `yaml:"input_from_sbom" mapstructure:"input_from_sbom" json:"input_from_sbom"`
	Registry       *common.Registry   `yaml:"registry" mapstructure:"registry" json:"registry"`
	LocalImageScan bool               `yaml:"local_image_scan" mapstructure:"local_image_scan" json:"local_image_scan"`
	ScannersConfig ScannersConfig     `yaml:"scanners_config" mapstructure:"scanners_config" json:"scanners_config"`
}

type ScannersConfig struct {
	Grype grypeconfig.Config `yaml:"grype" mapstructure:"grype" json:"grype"`
	Trivy trivyconfig.Config `yaml:"trivy" mapstructure:"trivy" json:"trivy"`
}
