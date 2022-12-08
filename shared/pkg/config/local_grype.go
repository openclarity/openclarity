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
	"github.com/spf13/viper"
)

const (
	ScannerLocalGrypeScope      = "SCANNER_LOCAL_GRYPE_SCOPE"
	ScannerLocalGrypeDBRootDir  = "SCANNER_LOCAL_GRYPE_DB_ROOT_DIR"
	ScannerLocalGrypeListingURL = "SCANNER_LOCAL_GRYPE_LISTING_URL"
	ScannerLocalGrypeUpdateDB   = "SCANNER_LOCAL_GRYPE_UPDATE_DB"
)

type LocalGrypeConfigEx struct {
	LocalGrypeConfig
	RegistryOptions *image.RegistryOptions
}

type LocalGrypeConfig struct {
	UpdateDB   bool         `yaml:"update_db" mapstructure:"update_db"`
	DBRootDir  string       `yaml:"db_root_dir" mapstructure:"db_root_dir"` // Location to write the vulnerability database cache.
	ListingURL string       `yaml:"listing_url" mapstructure:"listing_url"` // URL of the vulnerability database.
	Scope      source.Scope `yaml:"scope" mapstructure:"scope"`             // indicates "how" or from "which perspectives" the source object should be cataloged from.
}

func ConvertToLocalGrypeConfig(scanner *Scanner, registry *Registry) LocalGrypeConfigEx {
	credentials := make([]image.RegistryCredentials, len(registry.Auths))

	for i, cred := range registry.Auths {
		credentials[i] = image.RegistryCredentials{
			Authority: cred.Authority,
			Username:  cred.Username,
			Password:  cred.Password,
			Token:     cred.Token,
		}
	}

	return LocalGrypeConfigEx{
		LocalGrypeConfig: LocalGrypeConfig{
			UpdateDB:   scanner.GrypeConfig.UpdateDB,
			DBRootDir:  scanner.GrypeConfig.DBRootDir,
			ListingURL: scanner.GrypeConfig.ListingURL,
			Scope:      source.ParseScope(string(scanner.GrypeConfig.Scope)),
		},
		RegistryOptions: &image.RegistryOptions{
			InsecureSkipTLSVerify: registry.SkipVerifyTLS,
			InsecureUseHTTP:       registry.UseHTTP,
			Credentials:           credentials,
		},
	}
}

func loadLocalGrypeConfig() LocalGrypeConfig {
	setLocalGrypeScannerConfigDefaults()
	return LocalGrypeConfig{
		DBRootDir:  viper.GetString(ScannerLocalGrypeDBRootDir),
		ListingURL: viper.GetString(ScannerLocalGrypeListingURL),
		Scope:      source.ParseScope(viper.GetString(ScannerLocalGrypeScope)),
		UpdateDB:   viper.GetBool(ScannerLocalGrypeUpdateDB),
	}
}

func setLocalGrypeScannerConfigDefaults() {
	viper.SetDefault(ScannerLocalGrypeScope, source.SquashedScope)
	viper.SetDefault(ScannerLocalGrypeDBRootDir, "/tmp/")
	viper.SetDefault(ScannerLocalGrypeListingURL, "https://toolbox-data.anchore.io/grype/databases/listing.json")
	viper.SetDefault(ScannerLocalGrypeUpdateDB, true)
}
