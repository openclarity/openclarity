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
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix       = "OPENCLARITY_AWS"
	DefaultImageNameFilter = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-*"
	DefaultBlockDeviceName = "xvdh"

	AMD64 = "x86_64"
	ARM64 = "arm64"
)

var (
	DefaultImageOwners = []string{
		"099720109477", // Official Ubuntu Cloud account
	}

	DefaultInstanceTypeMapping = map[string]string{
		AMD64: "t3.large",
		ARM64: "t4g.large",
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
	KeyPairName     string `mapstructure:"keypair_name"`
	ImageNameFilter string `mapstructure:"image_filter_name"`
	// ImageOwners is a comma separated list of OwnerID(s)/OwnerAliases used as Owners filter for finding AMI
	// to instantiate Scanner instance.
	// See: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeImages.html
	ImageOwners []string `mapstructure:"image_owners"`
	// InstanceTypeMapping contains Architecture:InstanceType pairs
	InstanceTypeMapping InstanceTypeMapping `mapstructure:"instance_type_mapping"`
	// InstanceArchToUse contains the architecture to be used for Scanner instance which prevents the Provider
	// to dynamically determine it based on the Target architecture. The Provider will use this value to lookup
	// for InstanceType in InstanceTypeMapping.
	InstanceArchToUse string `mapstructure:"instance_arch_to_use"`
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

	switch c.InstanceArchToUse {
	case AMD64, ARM64:
		if _, ok := c.InstanceTypeMapping[c.InstanceArchToUse]; !ok {
			return fmt.Errorf("invalid InstanceArchToUse: %s architecture is missing from InstanceTypeMapping", c.InstanceArchToUse)
		}
	case "":
	default:
		return fmt.Errorf("invalid InstanceArchToUse: %s is not supported", c.InstanceArchToUse)
	}

	return nil
}

type InstanceTypeMapping map[string]string

func (m *InstanceTypeMapping) UnmarshalText(text []byte) error {
	mapping := make(InstanceTypeMapping)
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

	_ = v.BindEnv("image_filter_name")
	v.SetDefault("image_filter_name", DefaultImageNameFilter)

	_ = v.BindEnv("image_owners")
	v.SetDefault("image_owners", DefaultImageOwners)

	_ = v.BindEnv("instance_type_mapping")
	v.SetDefault("instance_type_mapping", DefaultInstanceTypeMapping)

	_ = v.BindEnv("instance_arch_to_use")

	_ = v.BindEnv("block_device_name")
	v.SetDefault("block_device_name", DefaultBlockDeviceName)

	decodeHooks := mapstructure.ComposeDecodeHookFunc(
		// TextUnmarshallerHookFunc is needed to decode InstanceTypeMapping
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
