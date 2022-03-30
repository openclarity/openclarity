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

import "github.com/spf13/viper"

const (
	BackendAPIKey             = "BACKEND_API_KEY"
	BackendHost               = "BACKEND_HOST"
	BackendDisableTLS         = "BACKEND_DISABLE_TLS"
	BackendInsecureSkipVerify = "BACKEND_INSECURE_SKIP_VERIFY"
)

type Backend struct {
	APIKey             string `json:"-"`
	Host               string `json:"host"`
	DisableTLS         bool   `json:"disable-tls"`
	InsecureSkipVerify bool   `json:"insecure-skip-verify"`
}

func setBackendConfigDefaults() {
	viper.SetDefault(BackendHost, "localhost:8080")
}

func loadBackendConfig() *Backend {
	setBackendConfigDefaults()
	return &Backend{
		APIKey:             viper.GetString(BackendAPIKey),
		Host:               viper.GetString(BackendHost),
		DisableTLS:         viper.GetBool(BackendDisableTLS),
		InsecureSkipVerify: viper.GetBool(BackendInsecureSkipVerify),
	}
}
