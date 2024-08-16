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
	AMD64 = "x86_64"
	ARM64 = "arm64"

	DefaultEnvPrefix = "OPENCLARITY_GCP"

	projectID                  = "project_id"
	scannerZone                = "scanner_zone"
	scannerSubnetwork          = "scanner_subnetwork"
	scannerMachineTypeMapping  = "scanner_machine_type_mapping"
	scannerMachineArchitecture = "scanner_machine_architecture"
	scannerSourceImagePrefix   = "scanner_source_image_prefix"
	scannerSourceImageMapping  = "scanner_source_image_mapping"
	scannerSSHPublicKey        = "scanner_ssh_public_key"
)

var (
	DefaultScannerMachineTypeMapping = Mapping{
		AMD64: "e2-standard-2",
		ARM64: "t2a-standard-2",
	}

	DefaultScannerSourceImagePrefix  = "projects/ubuntu-os-cloud/global/images/"
	DefaultScannerSourceImageMapping = Mapping{
		AMD64: "ubuntu-2204-jammy-v20230630",
		ARM64: "ubuntu-2204-jammy-arm64-v20230630",
	}
)

type Config struct {
	ProjectID                  string  `mapstructure:"project_id"`
	ScannerZone                string  `mapstructure:"scanner_zone"`
	ScannerSubnetwork          string  `mapstructure:"scanner_subnetwork"`
	ScannerMachineTypeMapping  Mapping `mapstructure:"scanner_machine_type_mapping"`
	ScannerMachineArchitecture string  `mapstructure:"scanner_machine_architecture"`
	ScannerSourceImagePrefix   string  `mapstructure:"scanner_source_image_prefix"`
	ScannerSourceImageMapping  Mapping `mapstructure:"scanner_source_image_mapping"`
	ScannerSSHPublicKey        string  `mapstructure:"scanner_ssh_public_key"`
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

	_ = v.BindEnv(scannerMachineTypeMapping)
	v.SetDefault(scannerMachineTypeMapping, DefaultScannerMachineTypeMapping)

	_ = v.BindEnv(scannerMachineArchitecture)

	_ = v.BindEnv(scannerSourceImagePrefix)
	v.SetDefault(scannerSourceImagePrefix, DefaultScannerSourceImagePrefix)

	_ = v.BindEnv(scannerSourceImageMapping)
	v.SetDefault(scannerSourceImageMapping, DefaultScannerSourceImageMapping)

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

	if c.ScannerMachineArchitecture == "" {
		return fmt.Errorf("parameter ScannerMachineArchitecture must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(scannerMachineArchitecture))
	}

	if _, ok := c.ScannerMachineTypeMapping[c.ScannerMachineArchitecture]; !ok {
		return fmt.Errorf("failed to find machine type for architecture %s", c.ScannerMachineArchitecture)
	}

	if c.ScannerSourceImagePrefix == "" {
		return fmt.Errorf("parameter ScannerSourceImageMappingPrefix must be provided by setting %v_%v environment variable", DefaultEnvPrefix, strings.ToUpper(scannerSourceImagePrefix))
	}

	if _, ok := c.ScannerSourceImageMapping[c.ScannerMachineArchitecture]; !ok {
		return fmt.Errorf("failed to find source image for architecture %s", c.ScannerMachineArchitecture)
	}

	return nil
}

type Mapping map[string]string

func (m *Mapping) UnmarshalText(text []byte) error {
	mapping := make(Mapping)
	items := strings.Split(string(text), ",")

	numOfParts := 2
	for _, item := range items {
		pair := strings.Split(item, ":")
		if len(pair) != numOfParts {
			continue
		}

		switch pair[0] {
		case AMD64, ARM64:
			mapping[pair[0]] = pair[1]
		default:
			return fmt.Errorf("unsupported architecture: %s", pair[0])
		}
	}
	*m = mapping

	return nil
}
