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
	"fmt"

	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix           = "VMCLARITY_AWS"
	DefaultScannerInstanceType = "t2.large"
	DefaultBlockDeviceName     = "xvdh"
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
	// ScannerImage is the AMI image used for creating Scanner instance
	ScannerImage string `mapstructure:"scanner_ami_id"`
	// ScannerInstanceType is the instance type used for Scanner instance
	ScannerInstanceType string `mapstructure:"scanner_instance_type"`
	// BlockDeviceName contains the block device name used for attaching Scanner volume to the Scanner instance
	BlockDeviceName string `mapstructure:"block_device_name"`
}

func (c *Config) Validate() error {
	if c.ScannerRegion == "" {
		return fmt.Errorf("parameter Region must be provided")
	}

	if c.SubnetID == "" {
		return fmt.Errorf("parameter SubnetID must be provided")
	}

	if c.SecurityGroupID == "" {
		return fmt.Errorf("parameter SecurityGroupID must be provided")
	}

	if c.ScannerImage == "" {
		return fmt.Errorf("parameter ScannerImage must be provided")
	}

	if c.ScannerInstanceType == "" {
		return fmt.Errorf("parameter ScannerInstanceType must be provided")
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
	_ = v.BindEnv("scanner_ami_id")

	_ = v.BindEnv("scanner_instance_type")
	v.SetDefault("scanner_instance_type", DefaultScannerInstanceType)

	_ = v.BindEnv("block_device_name")
	v.SetDefault("block_device_name", DefaultBlockDeviceName)

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=AWS: %w", err)
	}

	return config, nil
}
