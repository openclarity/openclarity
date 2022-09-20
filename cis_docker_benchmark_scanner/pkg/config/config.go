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
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	shared "github.com/openclarity/kubeclarity/shared/pkg/config"
)

const (
	Timeout = "TIMEOUT"
)

type Config struct {
	ResultServiceAddress string
	ScanUUID             string
	Timeout              time.Duration
	Registry             *shared.Registry
	ImageNameToScan      string // Not used
	ImageIDToScan        string
	ImageHashToScan      string // Not used
}

func LoadConfig() (*Config, error) {
	viper.SetDefault(Timeout, "2m")
	imageIDToScan := viper.GetString(shared.ImageIDToScan)
	config := &Config{
		ResultServiceAddress: viper.GetString(shared.ResultServiceAddress),
		ScanUUID:             viper.GetString(shared.ScanUUID),
		Registry:             shared.LoadRuntimeScannerRegistryConfig(imageIDToScan),
		ImageNameToScan:      viper.GetString(shared.ImageNameToScan),
		ImageIDToScan:        imageIDToScan,
		ImageHashToScan:      viper.GetString(shared.ImageHashToScan),
		Timeout:              viper.GetDuration(Timeout),
	}

	configB, err := json.Marshal(config)
	if err == nil {
		log.Infof("\n\nconfig=%s\n\n", configB)
	} else {
		log.Warningf("Failed to marshal config. %v", err)
	}

	return config, nil
}
