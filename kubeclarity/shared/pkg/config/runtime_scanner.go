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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/openclarity/kubeclarity/shared/pkg/utils/creds"
)

const (
	ResultServiceAddress  = "RESULT_SERVICE_ADDR"
	ImageIDToScan         = "IMAGE_ID_TO_SCAN"
	ImageHashToScan       = "IMAGE_HASH_TO_SCAN"
	ImageNameToScan       = "IMAGE_NAME_TO_SCAN"
	ScanUUID              = "SCAN_UUID"
	RegistrySkipVerifyTlS = "REGISTRY_SKIP_VERIFY_TLS"
	RegistryUseHTTP       = "REGISTRY_USE_HTTP"
	ImagePullSecretPath   = "IMAGE_PULL_SECRET_PATH" // nolint:gosec
)

func setRegistryConfigDefaults() {
	viper.SetDefault(RegistrySkipVerifyTlS, false)
	viper.SetDefault(RegistryUseHTTP, false)
}

func LoadRuntimeScannerRegistryConfig(imageID string) *Registry {
	setRegistryConfigDefaults()

	var auths []Auth
	authConfig, err := creds.GetAuthConfig(viper.GetString(ImagePullSecretPath), imageID)
	if err != nil {
		// Just log the issue don't fail, getting the credentials is best effort
		log.Warnf("Failed to resolve auth config from pull secrets: %v", err)
	} else {
		var auth Auth
		if authConfig.Username != "" {
			auth.Username = authConfig.Username
		}

		if authConfig.Password != "" {
			auth.Password = authConfig.Password
		}

		if authConfig.RegistryToken != "" {
			auth.Token = authConfig.RegistryToken
		}
		auths = append(auths, auth)
	}

	return &Registry{
		SkipVerifyTLS: viper.GetBool(RegistrySkipVerifyTlS),
		UseHTTP:       viper.GetBool(RegistryUseHTTP),
		Auths:         auths,
	}
}
