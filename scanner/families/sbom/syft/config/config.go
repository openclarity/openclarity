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
	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/syft/source"

	"github.com/openclarity/vmclarity/scanner/common"
)

type Config struct {
	Scope          string           `yaml:"scope" mapstructure:"scope" json:"scope"`
	ExcludePaths   []string         `yaml:"exclude_paths" mapstructure:"exclude_paths" json:"exclude_paths"`
	Registry       *common.Registry `yaml:"registry" mapstructure:"registry" json:"registry"`
	LocalImageScan bool             `yaml:"local_image_scan" mapstructure:"local_image_scan" json:"local_image_scan"`
}

func (c *Config) SetRegistry(registry *common.Registry) {
	c.Registry = registry
}

func (c *Config) SetLocalImageScan(localScan bool) {
	c.LocalImageScan = localScan
}

func (c *Config) GetScope() source.Scope {
	return source.ParseScope(c.Scope)
}

func (c *Config) GetExcludePaths() source.ExcludeConfig {
	return source.ExcludeConfig{
		Paths: c.ExcludePaths,
	}
}

func (c *Config) GetRegistryOptions() *image.RegistryOptions {
	if c.Registry == nil {
		return nil
	}

	credentials := make([]image.RegistryCredentials, len(c.Registry.Auths))

	for i, cred := range c.Registry.Auths {
		credentials[i] = image.RegistryCredentials{
			Authority: cred.Authority,
			Username:  cred.Username,
			Password:  cred.Password,
			Token:     cred.Token,
		}
	}

	return &image.RegistryOptions{
		InsecureSkipTLSVerify: c.Registry.SkipVerifyTLS,
		InsecureUseHTTP:       c.Registry.UseHTTP,
		Credentials:           credentials,
	}
}
