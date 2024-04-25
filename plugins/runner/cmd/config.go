// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package main

import (
	"os"

	"github.com/openclarity/vmclarity/plugins/runner/types"
)

const DefaultTimeoutSeconds = 15 * 60 // 15 minutes

func LoadConfig() types.PluginConfig {
	return types.PluginConfig{
		Name:           os.Getenv("PLUGIN_SCANNER_NAME"),
		ImageName:      os.Getenv("PLUGIN_SCANNER_IMAGE"),
		InputDir:       os.Getenv("PLUGIN_SCANNER_INPUT_DIR"),
		ScannerConfig:  os.Getenv("PLUGIN_SCANNER_CONFIG_JSON"),
		TimeoutSeconds: DefaultTimeoutSeconds,
	}
}
