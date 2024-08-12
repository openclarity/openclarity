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

package azure

import (
	"encoding/base64"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const (
	DefaultEnvPrefix = "VMCLARITY_AZURE"
)

type AzurePublicKey string

func (a *AzurePublicKey) UnmarshalText(text []byte) error {
	if len(text) != 0 {
		publicKey, err := base64.StdEncoding.DecodeString(string(text))
		if err != nil {
			return fmt.Errorf("failed to decode azure scanner public key from base64: %w", err)
		}
		*a = AzurePublicKey(publicKey)
	}
	return nil
}

type Config struct {
	SubscriptionID              string         `mapstructure:"subscription_id"`
	ScannerLocation             string         `mapstructure:"scanner_location"`
	ScannerResourceGroup        string         `mapstructure:"scanner_resource_group"`
	ScannerSubnet               string         `mapstructure:"scanner_subnet_id"`
	ScannerPublicKey            AzurePublicKey `mapstructure:"scanner_public_key"`
	ScannerVMSize               string         `mapstructure:"scanner_vm_size"`
	ScannerImagePublisher       string         `mapstructure:"scanner_image_publisher"`
	ScannerImageOffer           string         `mapstructure:"scanner_image_offer"`
	ScannerImageSKU             string         `mapstructure:"scanner_image_sku"`
	ScannerImageVersion         string         `mapstructure:"scanner_image_version"`
	ScannerSecurityGroup        string         `mapstructure:"scanner_security_group"`
	ScannerStorageAccountName   string         `mapstructure:"scanner_storage_account_name"`
	ScannerStorageContainerName string         `mapstructure:"scanner_storage_container_name"`
}

func NewConfig() (*Config, error) {
	// Avoid modifying the global instance
	v := viper.New()

	v.SetEnvPrefix(DefaultEnvPrefix)
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	_ = v.BindEnv("subscription_id")
	_ = v.BindEnv("scanner_location")
	_ = v.BindEnv("scanner_resource_group")
	_ = v.BindEnv("scanner_subnet_id")
	_ = v.BindEnv("scanner_public_key")
	_ = v.BindEnv("scanner_vm_size")
	_ = v.BindEnv("scanner_image_publisher")
	_ = v.BindEnv("scanner_image_offer")
	_ = v.BindEnv("scanner_image_sku")
	_ = v.BindEnv("scanner_image_version")
	_ = v.BindEnv("scanner_security_group")
	_ = v.BindEnv("scanner_storage_account_name")
	_ = v.BindEnv("scanner_storage_container_name")

	config := &Config{}
	if err := v.Unmarshal(&config, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
		return nil, fmt.Errorf("failed to parse provider configuration. Provider=Azure: %w", err)
	}
	return config, nil
}

// nolint:cyclop
func (c Config) Validate() error {
	if c.SubscriptionID == "" {
		return fmt.Errorf("parameter SubscriptionID must be provided")
	}

	if c.ScannerLocation == "" {
		return fmt.Errorf("parameter ScannerLocation must be provided")
	}

	if c.ScannerResourceGroup == "" {
		return fmt.Errorf("parameter ScannerResourceGroup must be provided")
	}

	if c.ScannerSubnet == "" {
		return fmt.Errorf("parameter ScannerSubnet must be provided")
	}

	if c.ScannerVMSize == "" {
		return fmt.Errorf("parameter ScannerVMSize must be provided")
	}

	if c.ScannerImagePublisher == "" {
		return fmt.Errorf("parameter ScannerImagePublisher must be provided")
	}

	if c.ScannerImageOffer == "" {
		return fmt.Errorf("parameter ScannerImageOffer must be provided")
	}

	if c.ScannerImageSKU == "" {
		return fmt.Errorf("parameter ScannerImageSKU must be provided")
	}

	if c.ScannerImageVersion == "" {
		return fmt.Errorf("parameter ScannerImageVersion must be provided")
	}

	if c.ScannerSecurityGroup == "" {
		return fmt.Errorf("parameter ScannerSecurityGroup must be provided")
	}

	if c.ScannerStorageAccountName == "" {
		return fmt.Errorf("parameter ScannerStorageAccountName must be provided")
	}

	if c.ScannerStorageContainerName == "" {
		return fmt.Errorf("parameter ScannerStorageContainerName must be provided")
	}

	return nil
}
