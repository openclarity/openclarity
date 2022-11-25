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
)

// TODO: maybe we need to extend syft confg.
type SyftConfig struct {
	Scope           source.Scope
	RegistryOptions *image.RegistryOptions
}

func CreateSyftConfig(analyzer *Analyzer, registry *Registry) SyftConfig {
	return SyftConfig{
		Scope:           source.ParseScope(analyzer.Scope),
		RegistryOptions: CreateRegistryOptions(registry),
	}
}

func CreateRegistryOptions(registry *Registry) *image.RegistryOptions {
	credentials := make([]image.RegistryCredentials, len(registry.Auths))

	for i, cred := range registry.Auths {
		credentials[i] = image.RegistryCredentials{
			Authority: cred.Authority,
			Username:  cred.Username,
			Password:  cred.Password,
			Token:     cred.Token,
		}
	}

	return &image.RegistryOptions{
		InsecureSkipTLSVerify: registry.SkipVerifyTLS,
		InsecureUseHTTP:       registry.UseHTTP,
		Credentials:           credentials,
	}
}
