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

package scanner

import (
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/common"
	exploittypes "github.com/openclarity/vmclarity/scanner/families/exploits/types"
	infofindertypes "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	malwaretypes "github.com/openclarity/vmclarity/scanner/families/malware/types"
	misconfigurationtypes "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	plugintypes "github.com/openclarity/vmclarity/scanner/families/plugins/types"
	rootkittypes "github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	sbomtypes "github.com/openclarity/vmclarity/scanner/families/sbom/types"
	secrettypes "github.com/openclarity/vmclarity/scanner/families/secrets/types"
	vulnerabilitytypes "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
)

type Config struct {
	// Analyzers
	SBOM sbomtypes.Config `json:"sbom" yaml:"sbom" mapstructure:"sbom"`

	// Scanners
	Vulnerabilities  vulnerabilitytypes.Config    `json:"vulnerabilities" yaml:"vulnerabilities" mapstructure:"vulnerabilities"`
	Secrets          secrettypes.Config           `json:"secrets" yaml:"secrets" mapstructure:"secrets"`
	Rootkits         rootkittypes.Config          `json:"rootkits" yaml:"rootkits" mapstructure:"rootkits"`
	Malware          malwaretypes.Config          `json:"malware" yaml:"malware" mapstructure:"malware"`
	Misconfiguration misconfigurationtypes.Config `json:"misconfiguration" yaml:"misconfiguration" mapstructure:"misconfiguration"`
	InfoFinder       infofindertypes.Config       `json:"infofinder" yaml:"infofinder" mapstructure:"infofinder"`

	// Enrichers
	Exploits exploittypes.Config `json:"exploits" yaml:"exploits" mapstructure:"exploits"`

	// Plugins
	Plugins plugintypes.Config `json:"plugins" yaml:"plugins" mapstructure:"plugins"`
}

func NewConfig() *Config {
	return &Config{
		SBOM:             sbomtypes.Config{},
		Vulnerabilities:  vulnerabilitytypes.Config{},
		Secrets:          secrettypes.Config{},
		Rootkits:         rootkittypes.Config{},
		Malware:          malwaretypes.Config{},
		Misconfiguration: misconfigurationtypes.Config{},
		InfoFinder:       infofindertypes.Config{},
		Exploits:         exploittypes.Config{},
		Plugins:          plugintypes.Config{},
	}
}

func (c *Config) AddInputs(inputType common.InputType, inputs []string) {
	for _, mountDir := range inputs {
		if c.SBOM.Enabled {
			c.SBOM.Inputs = append(c.SBOM.Inputs, common.ScanInput{
				Input:     mountDir,
				InputType: inputType,
			})
		}

		if c.Vulnerabilities.Enabled {
			if c.SBOM.Enabled {
				c.Vulnerabilities.InputFromSbom = true
			} else {
				c.Vulnerabilities.Inputs = append(c.Vulnerabilities.Inputs, common.ScanInput{
					Input:     mountDir,
					InputType: inputType,
				})
			}
		}

		if c.Secrets.Enabled {
			c.Secrets.Inputs = append(c.Secrets.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               mountDir,
				InputType:           inputType,
			})
		}

		if c.Malware.Enabled {
			c.Malware.Inputs = append(c.Malware.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               mountDir,
				InputType:           inputType,
			})
		}

		if c.Rootkits.Enabled {
			c.Rootkits.Inputs = append(c.Rootkits.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               mountDir,
				InputType:           inputType,
			})
		}

		if c.Misconfiguration.Enabled {
			c.Misconfiguration.Inputs = append(
				c.Misconfiguration.Inputs,
				common.ScanInput{
					StripPathFromResult: to.Ptr(true),
					Input:               mountDir,
					InputType:           inputType,
				},
			)
		}

		if c.InfoFinder.Enabled {
			c.InfoFinder.Inputs = append(
				c.InfoFinder.Inputs,
				common.ScanInput{
					StripPathFromResult: to.Ptr(true),
					Input:               mountDir,
					InputType:           inputType,
				},
			)
		}

		if c.Plugins.Enabled {
			c.Plugins.Inputs = append(c.Plugins.Inputs, common.ScanInput{
				Input:     mountDir,
				InputType: inputType,
			})
		}
	}
}
