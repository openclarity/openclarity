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

package aws

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	apitypes "github.com/openclarity/openclarity/api/types"
)

const (
	DefaultEnvPrefix                   = "OPENCLARITY_AWS"
	DefaultScannerInstanceArchitecture = apitypes.Amd64
	DefaultBlockDeviceName             = "xvdh"
)

var (
	DefaultScannerInstanceArchitectureToTypeMapping = apitypes.FromArchitectureMapping{
		"x86_64": "t3.large",
		"arm64":  "t4g.large",
	}

	DefaultScannerInstanceArchitectureToAMIMapping = apitypes.FromArchitectureMapping{
		"x86_64": "ami-03f1cc6c8b9c0b899",
		"arm64":  "ami-06972d841707cc4cf",
	}
)

type Config struct {
	// Region where the Scanner instance needs to be created
	ScannerRegion string `mapstructure:"scanner_region"`
	// SubnetID where the Scanner instance needs to be created
	SubnetID string `mapstructure:"subnet_id"`
	// SecurityGroupID which needs to be attached to the Scanner instance
	SecurityGroupID string `mapstructure:"security_group_id"`
	// KeyPairName is the name of the SSH KeyPair to use for Scanner instance launch
	KeyPairName string `mapstructure:"keypair_name"`
	// ScannerInstanceArchitecture contains the architecture to be used for Scanner instance which prevents the Provider
	// to dynamically determine it based on the Target architecture.
	ScannerInstanceArchitecture apitypes.VMInfoArchitecture `mapstructure:"scanner_instance_architecture"`
	// ScannerInstanceArchitectureToTypeMapping contains Architecture:InstanceType pairs
	ScannerInstanceArchitectureToTypeMapping apitypes.FromArchitectureMapping `mapstructure:"scanner_instance_architecture_to_type_mapping"`
	// ScannerInstanceArchitectureToAMIMapping contains Architecture:AMI pairs
	ScannerInstanceArchitectureToAMIMapping apitypes.FromArchitectureMapping `mapstructure:"scanner_instance_architecture_to_ami_mapping"`
	// BlockDeviceName contains the block device name used for attaching Scanner volume to the Scanner instance
	BlockDeviceName string `mapstructure:"block_device_name"`
}

func (c *Config) Validate() error {
	if c.ScannerRegion == "" {
		return errors.New("parameter Region must be provided")
	}

	if c.SubnetID == "" {
		return errors.New("parameter SubnetID must be provided")
	}

	if c.SecurityGroupID == "" {
		return errors.New("parameter SecurityGroupID must be provided")
	}

	architecture, err := c.ScannerInstanceArchitecture.MarshalText()
	if err != nil {
		return fmt.Errorf("failed to marshal ScannerInstanceArchitecture into text: %w", err)
	}

	if _, ok := c.ScannerInstanceArchitectureToTypeMapping[architecture]; !ok {
		return fmt.Errorf("failed to find instance type for architecture. Arch=%s", architecture)
	}

	if _, ok := c.ScannerInstanceArchitectureToAMIMapping[architecture]; !ok {
		return fmt.Errorf("failed to find instance AMI for architecture. Arch=%s", architecture)
	}

	return nil
}

func NewConfig() (*Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("scanner_region")
	_ = v.BindEnv("subnet_id")
	_ = v.BindEnv("security_group_id")
	_ = v.BindEnv("keypair_name")

	_ = v.BindEnv("scanner_instance_architecture")
	v.SetDefault("scanner_instance_architecture", DefaultScannerInstanceArchitecture)

	_ = v.BindEnv("scanner_instance_architecture_to_type_mapping")
	v.SetDefault("scanner_instance_architecture_to_type_mapping", DefaultScannerInstanceArchitectureToTypeMapping)

	_ = v.BindEnv("scanner_instance_architecture_to_ami_mapping")
	v.SetDefault("scanner_instance_architecture_to_ami_mapping", DefaultScannerInstanceArchitectureToAMIMapping)

	_ = v.BindEnv("block_device_name")
	v.SetDefault("block_device_name", DefaultBlockDeviceName)

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		// TextUnmarshallerHookFunc is needed to decode the custom types
		mapstructure.TextUnmarshallerHookFunc(),
		// Default decoders
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)

	config := &Config{}
	if err := v.Unmarshal(config, viper.DecodeHook(decodeHooks)); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=AWS: %w", err)
	}

	return config, nil
}
