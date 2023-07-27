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

package families

import (
	kubeclarityutils "github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	misconfigurationTypes "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type Config struct {
	// Analyzers
	SBOM sbom.Config `json:"sbom" yaml:"sbom" mapstructure:"sbom"`

	// Scanners
	Vulnerabilities  vulnerabilities.Config       `json:"vulnerabilities" yaml:"vulnerabilities" mapstructure:"vulnerabilities"`
	Secrets          secrets.Config               `json:"secrets" yaml:"secrets" mapstructure:"secrets"`
	Rootkits         rootkits.Config              `json:"rootkits" yaml:"rootkits" mapstructure:"rootkits"`
	Malware          malware.Config               `json:"malware" yaml:"malware" mapstructure:"malware"`
	Misconfiguration misconfigurationTypes.Config `json:"misconfiguration" yaml:"misconfiguration" mapstructure:"misconfiguration"`

	// Enrichers
	Exploits exploits.Config `json:"exploits" yaml:"exploits" mapstructure:"exploits"`
}

func NewConfig() *Config {
	return &Config{
		SBOM:             sbom.Config{},
		Vulnerabilities:  vulnerabilities.Config{},
		Secrets:          secrets.Config{},
		Rootkits:         rootkits.Config{},
		Malware:          malware.Config{},
		Misconfiguration: misconfigurationTypes.Config{},
		Exploits:         exploits.Config{},
	}
}

func SetMountPointsForFamiliesInput(mountPoints []string, familiesConfig *Config) *Config {
	// update families inputs with the mount point as rootfs
	for _, mountDir := range mountPoints {
		if familiesConfig.SBOM.Enabled {
			familiesConfig.SBOM.Inputs = append(familiesConfig.SBOM.Inputs, sbom.Input{
				Input:     mountDir,
				InputType: string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Vulnerabilities.Enabled {
			if familiesConfig.SBOM.Enabled {
				familiesConfig.Vulnerabilities.InputFromSbom = true
			} else {
				familiesConfig.Vulnerabilities.Inputs = append(familiesConfig.Vulnerabilities.Inputs, vulnerabilities.Input{
					Input:     mountDir,
					InputType: string(kubeclarityutils.ROOTFS),
				})
			}
		}

		if familiesConfig.Secrets.Enabled {
			familiesConfig.Secrets.Inputs = append(familiesConfig.Secrets.Inputs, secrets.Input{
				StripPathFromResult: utils.PointerTo(true),
				Input:               mountDir,
				InputType:           string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Malware.Enabled {
			familiesConfig.Malware.Inputs = append(familiesConfig.Malware.Inputs, malware.Input{
				StripPathFromResult: utils.PointerTo(true),
				Input:               mountDir,
				InputType:           string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Rootkits.Enabled {
			familiesConfig.Rootkits.Inputs = append(familiesConfig.Rootkits.Inputs, rootkits.Input{
				StripPathFromResult: utils.PointerTo(true),
				Input:               mountDir,
				InputType:           string(kubeclarityutils.ROOTFS),
			})
		}

		if familiesConfig.Misconfiguration.Enabled {
			familiesConfig.Misconfiguration.Inputs = append(
				familiesConfig.Misconfiguration.Inputs,
				misconfigurationTypes.Input{
					StripPathFromResult: utils.PointerTo(true),
					Input:               mountDir,
					InputType:           string(kubeclarityutils.ROOTFS),
				},
			)
		}
	}
	return familiesConfig
}
