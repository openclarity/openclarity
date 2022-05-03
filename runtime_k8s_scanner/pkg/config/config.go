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

package config

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	shared "github.com/openclarity/kubeclarity/shared/pkg/config"
)

const (
	SBOMDBAddress = "SBOM_DB_ADDR"
)

type Config struct {
	ResultServiceAddress string
	SBOMDBAddress        string
	ImageIDToScan        string
	ImageHashToScan      string
	ImagePullSecret      string `yaml:"-" json:"-"`
	ScanUUID             string
	RegistryInsecure     bool
	SharedConfig         *shared.Config
	ImageNameToScan      string
}

func LoadConfig() (*Config, error) {
	imageIDToScan := viper.GetString(shared.ImageIDToScan)
	config := &Config{
		ResultServiceAddress: viper.GetString(shared.ResultServiceAddress),
		SBOMDBAddress:        viper.GetString(SBOMDBAddress),
		ImageIDToScan:        imageIDToScan,
		ImageHashToScan:      viper.GetString(shared.ImageHashToScan),
		ImageNameToScan:      viper.GetString(shared.ImageNameToScan),
		ImagePullSecret:      viper.GetString(shared.ImagePullSecret),
		ScanUUID:             viper.GetString(shared.ScanUUID),
		SharedConfig: &shared.Config{
			Registry: shared.LoadRuntimeScannerRegistryConfig(imageIDToScan),
			Analyzer: shared.LoadAnalyzerConfig(),
			Scanner:  shared.LoadScannerConfig(),
		},
	}

	configB, _ := json.Marshal(config)
	log.Infof("\n\nconfig=%s\n\n", configB)

	return config, nil
}
