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
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/families/exploits"
	infofinderTypes "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	"github.com/openclarity/vmclarity/scanner/families/malware"
	misconfigurationTypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/families/plugins"
	"github.com/openclarity/vmclarity/scanner/families/rootkits"
	"github.com/openclarity/vmclarity/scanner/families/sbom"
	"github.com/openclarity/vmclarity/scanner/families/secrets"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities"
	"github.com/openclarity/vmclarity/scanner/utils"
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
	InfoFinder       infofinderTypes.Config       `json:"infofinder" yaml:"infofinder" mapstructure:"infofinder"`

	// Enrichers
	Exploits exploits.Config `json:"exploits" yaml:"exploits" mapstructure:"exploits"`

	// Plugins
	Plugins plugins.Config `json:"plugins" yaml:"plugins" mapstructure:"plugins"`
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
			familiesConfig.SBOM.Inputs = append(familiesConfig.SBOM.Inputs, types.Input{
				Input:     mountDir,
				InputType: string(utils.ROOTFS),
			})
		}

		if familiesConfig.Vulnerabilities.Enabled {
			if familiesConfig.SBOM.Enabled {
				familiesConfig.Vulnerabilities.InputFromSbom = true
			} else {
				familiesConfig.Vulnerabilities.Inputs = append(familiesConfig.Vulnerabilities.Inputs, types.Input{
					Input:     mountDir,
					InputType: string(utils.ROOTFS),
				})
			}
		}

		if familiesConfig.Secrets.Enabled {
			familiesConfig.Secrets.Inputs = append(familiesConfig.Secrets.Inputs, types.Input{
				StripPathFromResult: to.Ptr(true),
				Input:               mountDir,
				InputType:           string(utils.ROOTFS),
			})
		}

		if familiesConfig.Malware.Enabled {
			familiesConfig.Malware.Inputs = append(familiesConfig.Malware.Inputs, types.Input{
				StripPathFromResult: to.Ptr(true),
				Input:               mountDir,
				InputType:           string(utils.ROOTFS),
			})
		}

		if familiesConfig.Rootkits.Enabled {
			familiesConfig.Rootkits.Inputs = append(familiesConfig.Rootkits.Inputs, types.Input{
				StripPathFromResult: to.Ptr(true),
				Input:               mountDir,
				InputType:           string(utils.ROOTFS),
			})
		}

		if familiesConfig.Misconfiguration.Enabled {
			familiesConfig.Misconfiguration.Inputs = append(
				familiesConfig.Misconfiguration.Inputs,
				types.Input{
					StripPathFromResult: to.Ptr(true),
					Input:               mountDir,
					InputType:           string(utils.ROOTFS),
				},
			)
		}

		if familiesConfig.InfoFinder.Enabled {
			familiesConfig.InfoFinder.Inputs = append(
				familiesConfig.InfoFinder.Inputs,
				types.Input{
					StripPathFromResult: to.Ptr(true),
					Input:               mountDir,
					InputType:           string(utils.ROOTFS),
				},
			)
		}

		if familiesConfig.Plugins.Enabled {
			familiesConfig.Plugins.Inputs = append(familiesConfig.Plugins.Inputs, types.Input{
				Input:     mountDir,
				InputType: string(utils.ROOTFS),
			})
		}
	}
	return familiesConfig
}

// TODO(sambetts) Refactor this and the function above.
func SetOciArchiveForFamiliesInput(archives []string, familiesConfig *Config) *Config {
	// update families inputs with the oci archives
	for _, archive := range archives {
		if familiesConfig.SBOM.Enabled {
			familiesConfig.SBOM.Inputs = append(familiesConfig.SBOM.Inputs, types.Input{
				Input:     archive,
				InputType: string(utils.OCIARCHIVE),
			})
		}

		if familiesConfig.Vulnerabilities.Enabled {
			if familiesConfig.SBOM.Enabled {
				familiesConfig.Vulnerabilities.InputFromSbom = true
			} else {
				familiesConfig.Vulnerabilities.Inputs = append(familiesConfig.Vulnerabilities.Inputs, types.Input{
					Input:     archive,
					InputType: string(utils.OCIARCHIVE),
				})
			}
		}

		if familiesConfig.Secrets.Enabled {
			familiesConfig.Secrets.Inputs = append(familiesConfig.Secrets.Inputs, types.Input{
				StripPathFromResult: to.Ptr(true),
				Input:               archive,
				InputType:           string(utils.OCIARCHIVE),
			})
		}

		if familiesConfig.Malware.Enabled {
			familiesConfig.Malware.Inputs = append(familiesConfig.Malware.Inputs, types.Input{
				StripPathFromResult: to.Ptr(true),
				Input:               archive,
				InputType:           string(utils.OCIARCHIVE),
			})
		}

		if familiesConfig.Rootkits.Enabled {
			familiesConfig.Rootkits.Inputs = append(familiesConfig.Rootkits.Inputs, types.Input{
				StripPathFromResult: to.Ptr(true),
				Input:               archive,
				InputType:           string(utils.OCIARCHIVE),
			})
		}

		if familiesConfig.Misconfiguration.Enabled {
			familiesConfig.Misconfiguration.Inputs = append(
				familiesConfig.Misconfiguration.Inputs,
				types.Input{
					StripPathFromResult: to.Ptr(true),
					Input:               archive,
					InputType:           string(utils.OCIARCHIVE),
				},
			)
		}

		if familiesConfig.InfoFinder.Enabled {
			familiesConfig.InfoFinder.Inputs = append(
				familiesConfig.InfoFinder.Inputs,
				types.Input{
					StripPathFromResult: to.Ptr(true),
					Input:               archive,
					InputType:           string(utils.OCIARCHIVE),
				},
			)
		}
	}
	return familiesConfig
}
