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

package trivy

import (
	"os"

	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"

	"github.com/openclarity/vmclarity/scanner/config"
)

func SetTrivyRegistryConfigs(registry *config.Registry, trivyOptions trivyFlag.Options) trivyFlag.Options {
	if registry.UseHTTP {
		os.Setenv("TRIVY_NON_SSL", "true")
	}
	if registry.SkipVerifyTLS {
		trivyOptions.GlobalOptions.Insecure = true
	}
	if len(registry.Auths) >= 1 {
		// Trivy only supports one auth right now so use the
		// first entry
		auth := registry.Auths[0]
		if auth.Username != "" {
			os.Setenv("TRIVY_USERNAME", auth.Username)
		}
		if auth.Password != "" {
			os.Setenv("TRIVY_PASSWORD", auth.Password)
		}
		if auth.Token != "" {
			os.Setenv("TRIVY_REGISTRY_TOKEN", auth.Token)
		}
	}
	return trivyOptions
}
