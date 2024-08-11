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

package gcp

import (
	"fmt"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix = "VMCLARITY_GCP"

	projectID           = "project_id"
	scannerZone         = "scanner_zone"
	scannerSubnetwork   = "scanner_subnetwork"
	scannerMachineType  = "scanner_machine_type"
	scannerSourceImage  = "scanner_source_image"
	scannerSSHPublicKey = "scanner_ssh_public_key"
)

type Config struct {
	ProjectID           string `mapstructure:"project_id"`
	ScannerZone         string `mapstructure:"scanner_zone"`
	ScannerSubnetwork   string `mapstructure:"scanner_subnetwork"`
	ScannerMachineType  string `mapstructure:"scanner_machine_type"`
	ScannerSourceImage  string `mapstructure:"scanner_source_image"`
	ScannerSSHPublicKey string `mapstructure:"scanner_ssh_public_key"`
}

func NewConfig() (*Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv(projectID)
	_ = v.BindEnv(scannerZone)
	_ = v.BindEnv(scannerSubnetwork)
	_ = v.BindEnv(scannerMachineType)
	_ = v.BindEnv(scannerSourceImage)
	_ = v.BindEnv(scannerSSHPublicKey)

	config := &Config{}
	if err := v.Unmarshal(&config, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=GCP: %w", err)
	}
	return config, nil
}

// nolint:cyclop
func (c Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("parameter ProjectID must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(projectID))
	}

	if c.ScannerZone == "" {
		return fmt.Errorf("parameter ScannerZone must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(scannerZone))
	}

	if c.ScannerSubnetwork == "" {
		return fmt.Errorf("parameter ScannerSubnetwork must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(scannerSubnetwork))
	}

	if c.ScannerMachineType == "" {
		return fmt.Errorf("parameter ScannerMachineType must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(scannerMachineType))
	}

	if c.ScannerSourceImage == "" {
		return fmt.Errorf("parameter ScannerSourceImage must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(scannerSourceImage))
	}

	return nil
}
