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
	"time"

	"github.com/anchore/stereoscope/pkg/image"
	"github.com/anchore/syft/syft/source"

	"github.com/openclarity/vmclarity/scanner/common"
)

const (
	DefaultListingURL         = "https://toolbox-data.anchore.io/grype/databases/listing.json"
	DefaultListingFileTimeout = 60 * time.Second
	DefaultUpdateTimeout      = 60 * time.Second
	DefaultMaxDatabaseAge     = 120 * time.Hour
)

type LocalGrypeConfig struct {
	UpdateDB           bool             `yaml:"update_db" mapstructure:"update_db" json:"update_db"`
	DBRootDir          string           `yaml:"db_root_dir" mapstructure:"db_root_dir" json:"db_root_dir"`                               // Location to write the vulnerability database cache.
	ListingURL         string           `yaml:"listing_url" mapstructure:"listing_url" json:"listing_url"`                               // URL of the vulnerability database.
	MaxAllowedBuiltAge time.Duration    `yaml:"max_allowed_built_age" mapstructure:"max_allowed_built_age" json:"max_allowed_built_age"` // Period of time after which the database is considered stale.
	ListingFileTimeout time.Duration    `yaml:"listing_file_timeout" mapstructure:"listing_file_timeout" json:"listing_file_timeout"`    // Timeout of grype's HTTP client used for downloading the listing file.
	UpdateTimeout      time.Duration    `yaml:"update_timeout" mapstructure:"update_timeout" json:"update_timeout"`                      // Timeout of grype's HTTP client used for downloading the database.
	Scope              string           `yaml:"scope" mapstructure:"scope" json:"scope"`                                                 // indicates "how" or from "which perspectives" the source object should be cataloged from.
	Registry           *common.Registry `yaml:"registry" mapstructure:"registry" json:"registry"`
	LocalImageScan     bool             `yaml:"local_image_scan" mapstructure:"local_image_scan" json:"local_image_scan"`
}

func (c *LocalGrypeConfig) SetRegistry(registry *common.Registry) {
	c.Registry = registry
}

func (c *LocalGrypeConfig) SetLocalImageScan(localScan bool) {
	c.LocalImageScan = localScan
}

func (c *LocalGrypeConfig) GetScope() source.Scope {
	return source.ParseScope(c.Scope)
}

func (c *LocalGrypeConfig) GetListingURL() string {
	if c.ListingURL != "" {
		return c.ListingURL
	}

	return DefaultListingURL
}

func (c *LocalGrypeConfig) GetListingFileTimeout() time.Duration {
	if c.ListingFileTimeout > 0 {
		return c.ListingFileTimeout
	}

	return DefaultListingFileTimeout
}

func (c *LocalGrypeConfig) GetUpdateTimeout() time.Duration {
	if c.UpdateTimeout > 0 {
		return c.UpdateTimeout
	}

	return DefaultUpdateTimeout
}

func (c *LocalGrypeConfig) GetMaxDatabaseAge() time.Duration {
	if c.MaxAllowedBuiltAge > 0 {
		return c.MaxAllowedBuiltAge
	}

	return DefaultMaxDatabaseAge
}

func (c *LocalGrypeConfig) GetRegistryOptions() *image.RegistryOptions {
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
