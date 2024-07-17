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
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
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

type ConfigOption func(*Config)

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

func (c *Config) AddInputs(inputType common.InputType, inputs []string) {
	for _, input := range inputs {
		if c.SBOM.Enabled {
			c.SBOM.Inputs = append(c.SBOM.Inputs, common.ScanInput{
				Input:     input,
				InputType: inputType,
			})
		}

		if c.Vulnerabilities.Enabled {
			if c.SBOM.Enabled {
				c.Vulnerabilities.InputFromSbom = true
			} else {
				c.Vulnerabilities.Inputs = append(c.Vulnerabilities.Inputs, common.ScanInput{
					Input:     input,
					InputType: inputType,
				})
			}
		}

		if c.Secrets.Enabled {
			c.Secrets.Inputs = append(c.Secrets.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               input,
				InputType:           inputType,
			})
		}

		if c.Exploits.Enabled {
			c.Exploits.Inputs = append(c.Exploits.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               input,
				InputType:           inputType,
			})
		}

		if c.Misconfiguration.Enabled {
			c.Misconfiguration.Inputs = append(c.Misconfiguration.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               input,
				InputType:           inputType,
			})
		}

		if c.InfoFinder.Enabled {
			c.InfoFinder.Inputs = append(c.InfoFinder.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               input,
				InputType:           inputType,
			})
		}

		if c.Malware.Enabled {
			c.Malware.Inputs = append(c.Malware.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               input,
				InputType:           inputType,
			})
		}

		if c.Rootkits.Enabled {
			c.Rootkits.Inputs = append(c.Rootkits.Inputs, common.ScanInput{
				StripPathFromResult: to.Ptr(true),
				Input:               input,
				InputType:           inputType,
			})
		}

		if c.Plugins.Enabled {
			c.Plugins.Inputs = append(c.Plugins.Inputs, common.ScanInput{
				Input:     input,
				InputType: inputType,
			})
		}
	}
}

func (c *Config) GetFamilyInputs(family families.FamilyType) []common.ScanInput {
	switch family {
	case families.SBOM:
		return c.SBOM.Inputs
	case families.Vulnerabilities:
		return c.Vulnerabilities.Inputs
	case families.Secrets:
		return c.Secrets.Inputs
	case families.Rootkits:
		return c.Rootkits.Inputs
	case families.Malware:
		return c.Malware.Inputs
	case families.Misconfiguration:
		return c.Misconfiguration.Inputs
	case families.InfoFinder:
		return c.InfoFinder.Inputs
	case families.Exploits:
		return c.Exploits.Inputs
	case families.Plugins:
		return c.Plugins.Inputs
	default:
		return nil
	}
}

func (c *Config) SetFamilyInputs(family families.FamilyType, inputs []common.ScanInput) {
	switch family {
	case families.SBOM:
		c.SBOM.Inputs = inputs
	case families.Vulnerabilities:
		c.Vulnerabilities.Inputs = inputs
	case families.Secrets:
		c.Secrets.Inputs = inputs
	case families.Rootkits:
		c.Rootkits.Inputs = inputs
	case families.Malware:
		c.Malware.Inputs = inputs
	case families.Misconfiguration:
		c.Misconfiguration.Inputs = inputs
	case families.InfoFinder:
		c.InfoFinder.Inputs = inputs
	case families.Exploits:
		c.Exploits.Inputs = inputs
	case families.Plugins:
		c.Plugins.Inputs = inputs
	default:
	}
}

func NewConfig(opts ...ConfigOption) *Config {
	config := &Config{
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

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// WithBaseConfig uses provided config to construct a new Config. It is used when
// trying to apply custom options to an existing Config.
//
// WithBaseConfig should only be used as an argument to NewConfig.
func WithBaseConfig(base Config) ConfigOption {
	return func(config *Config) {
		data, _ := yaml.Marshal(base)

		var cloned Config
		_ = yaml.Unmarshal(data, &cloned)

		*config = cloned
	}
}

// WithInputsMountOverride replaces filesystem inputs for all families with the
// same inputs but for given mount points. Inputs will be duplicated for every
// mount point using the mount point as the input path prefix.
//
// WithInputsMountOverride can only be used after WithBaseConfig and should only
// be used as an argument to NewConfig.
func WithInputsMountOverride(mountpoints ...string) ConfigOption {
	// Define input replacer function
	replacerFn := func(input common.ScanInput) []common.ScanInput {
		// If the input type cannot be found on the filesystem, return it without changes
		if !input.InputType.IsOnFilesystem() {
			return []common.ScanInput{input}
		}

		// If the input type can be found on the filesystem, expand it to mounted inputs
		// for all mount points
		var replacedInputs []common.ScanInput
		for _, mountpoint := range mountpoints {
			replacedInputs = append(replacedInputs, common.ScanInput{
				StripPathFromResult: input.StripPathFromResult,
				Input:               filepath.Join(mountpoint, input.Input),
				InputType:           input.InputType,
			})
		}

		return replacedInputs
	}

	// Return function that applies replacer to all families
	return func(config *Config) {
		for _, opt := range []ConfigOption{
			WithFamilyInputsReplacer(families.SBOM, replacerFn),
			WithFamilyInputsReplacer(families.Vulnerabilities, replacerFn),
			WithFamilyInputsReplacer(families.Secrets, replacerFn),
			WithFamilyInputsReplacer(families.Rootkits, replacerFn),
			WithFamilyInputsReplacer(families.Malware, replacerFn),
			WithFamilyInputsReplacer(families.Misconfiguration, replacerFn),
			WithFamilyInputsReplacer(families.InfoFinder, replacerFn),
			WithFamilyInputsReplacer(families.Exploits, replacerFn),
			WithFamilyInputsReplacer(families.Plugins, replacerFn),
		} {
			opt(config)
		}
	}
}

// WithFamilyInputsReplacer replaces specific family inputs using custom replacer
// function.
//
// WithFamilyInputsReplacer can only be used after WithBaseConfig and should only
// be used as an argument to NewConfig.
func WithFamilyInputsReplacer(family families.FamilyType, replacerFn func(common.ScanInput) []common.ScanInput) ConfigOption {
	return func(config *Config) {
		// Get the current inputs for the specified family
		currentInputs := config.GetFamilyInputs(family)

		// Apply the replacer function to each input
		var updatedInputs []common.ScanInput
		for _, input := range currentInputs {
			updatedInputs = append(updatedInputs, replacerFn(input)...)
		}

		// Set the updated inputs for the specified family
		config.SetFamilyInputs(family, updatedInputs)
	}
}
