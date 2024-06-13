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

package assetscan

import (
	"time"

	"github.com/anchore/syft/syft/source"

	apitypes "github.com/openclarity/vmclarity/api/types"
	cliconfig "github.com/openclarity/vmclarity/scanner/config"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/exploits"
	exploitsCommon "github.com/openclarity/vmclarity/scanner/families/exploits/common"
	exploitdbConfig "github.com/openclarity/vmclarity/scanner/families/exploits/exploitdb/config"
	infofinderTypes "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	"github.com/openclarity/vmclarity/scanner/families/malware"
	clamconfig "github.com/openclarity/vmclarity/scanner/families/malware/clam/config"
	malwarecommon "github.com/openclarity/vmclarity/scanner/families/malware/common"
	yaraconfig "github.com/openclarity/vmclarity/scanner/families/malware/yara/config"
	misconfiguration "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/families/plugins"
	pluginscommon "github.com/openclarity/vmclarity/scanner/families/plugins/common"
	pluginsrunnerconfig "github.com/openclarity/vmclarity/scanner/families/plugins/runner/config"
	"github.com/openclarity/vmclarity/scanner/families/rootkits"
	chkrootkitConfig "github.com/openclarity/vmclarity/scanner/families/rootkits/chkrootkit/config"
	rootkitsCommon "github.com/openclarity/vmclarity/scanner/families/rootkits/common"
	"github.com/openclarity/vmclarity/scanner/families/sbom"
	"github.com/openclarity/vmclarity/scanner/families/secrets"
	secretscommon "github.com/openclarity/vmclarity/scanner/families/secrets/common"
	gitleaksconfig "github.com/openclarity/vmclarity/scanner/families/secrets/gitleaks/config"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities"
)

type FamiliesConfigOption func(*families.Config)

func withSBOM(config *apitypes.SBOMConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.SBOM = sbom.Config{
			Enabled:       true,
			AnalyzersList: config.GetAnalyzersList(),
			Inputs:        nil, // rootfs directory will be determined by the CLI after mount.
			AnalyzersConfig: &cliconfig.Config{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Registry: &cliconfig.Registry{},
				Analyzer: &cliconfig.Analyzer{
					Scope:        "squashed", // TODO(sambetts) This should be a default in the scanner/cli config
					OutputFormat: "cyclonedx",
					TrivyConfig: cliconfig.AnalyzerTrivyConfig{
						Timeout: int(opts.TrivyScanTimeout / time.Second), // NOTE(chrisgacsal): Timeout is expected to be in seconds.
					},
				},
			},
		}
	}
}

func withVulnerabilities(config *apitypes.VulnerabilitiesConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		var grypeConfig cliconfig.GrypeConfig
		if opts.GrypeServerAddress != "" {
			grypeConfig = cliconfig.GrypeConfig{
				Mode: cliconfig.ModeRemote,
				RemoteGrypeConfig: cliconfig.RemoteGrypeConfig{
					GrypeServerAddress: opts.GrypeServerAddress,
					GrypeServerTimeout: opts.GrypeServerTimeout,
				},
			}
		} else {
			grypeConfig = cliconfig.GrypeConfig{
				Mode: cliconfig.ModeLocal,
				LocalGrypeConfig: cliconfig.LocalGrypeConfig{
					UpdateDB:           true,
					DBRootDir:          "/tmp/",
					ListingURL:         cliconfig.DefaultGrypeListingURL,
					MaxAllowedBuiltAge: cliconfig.DefaultGrypeMaxDatabaseAge,
					ListingFileTimeout: cliconfig.DefaultGrypeListingFileTimeout,
					UpdateTimeout:      cliconfig.DefaultGrypeUpdateTimeout,
					Scope:              source.SquashedScope,
				},
			}
		}

		c.Vulnerabilities = vulnerabilities.Config{
			Enabled:       true,
			ScannersList:  config.GetScannersList(),
			InputFromSbom: false, // will be determined by the CLI.
			ScannersConfig: &cliconfig.Config{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Registry: &cliconfig.Registry{},
				Scanner: &cliconfig.Scanner{
					GrypeConfig: grypeConfig,
					TrivyConfig: cliconfig.ScannerTrivyConfig{
						Timeout:    int(opts.TrivyScanTimeout / time.Second), // NOTE(chrisgacsal): Timeout is expected to be in seconds.
						ServerAddr: opts.TrivyServerAddress,
					},
				},
			},
		}
	}
}

func withSecretsConfig(config *apitypes.SecretsConfig, _ *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Secrets = secrets.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: &secretscommon.ScannersConfig{
				Gitleaks: gitleaksconfig.Config{
					BinaryPath: "",
				},
			},
		}
	}
}

func withExploitsConfig(config *apitypes.ExploitsConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Exploits = exploits.Config{
			Enabled:       true,
			ScannersList:  config.GetScannersList(),
			InputFromVuln: true,
			ScannersConfig: &exploitsCommon.ScannersConfig{
				ExploitDB: exploitdbConfig.Config{
					BaseURL: opts.ExploitsDBAddress,
				},
			},
		}
	}
}

func withMalwareConfig(config *apitypes.MalwareConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Malware = malware.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: &malwarecommon.ScannersConfig{
				Clam: clamconfig.Config{
					ClamScanBinaryPath:            "",
					FreshclamBinaryPath:           "",
					AlternativeFreshclamMirrorURL: opts.AlternativeFreshclamMirrorURL,
				},
				Yara: yaraconfig.Config{
					YaraBinaryPath:    "",
					CompiledRuleURL:   opts.YaraRuleServerAddress,
					DirectoriesToScan: config.GetYaraDirectoriesToScan(),
				},
			},
		}
	}
}

func withMisconfigurationConfig(config *apitypes.MisconfigurationsConfig, _ *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Misconfiguration = misconfiguration.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: misconfiguration.ScannersConfig{
				// TODO(sambetts) Add scanner configurations here as we add them like Lynis
				Lynis: misconfiguration.LynisConfig{
					BinaryPath: "",
				},
				CISDocker: misconfiguration.CISDockerConfig{
					Timeout:  0,
					Registry: &cliconfig.Registry{},
				},
			},
		}
	}
}

func withInfoFinderConfig(config *apitypes.InfoFinderConfig, _ *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.InfoFinder = infofinderTypes.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: infofinderTypes.ScannersConfig{
				SSHTopology: infofinderTypes.SSHTopologyConfig{},
			},
		}
	}
}

func withRootkitsConfig(config *apitypes.RootkitsConfig, _ *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Rootkits = rootkits.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil,
			ScannersConfig: &rootkitsCommon.ScannersConfig{
				Chkrootkit: chkrootkitConfig.Config{
					BinaryPath: "",
				},
			},
		}
	}
}

func withPluginsConfig(config *apitypes.PluginsConfig, _ *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		scannersConfig := pluginscommon.ScannersConfig{}
		for k, v := range *config.ScannersConfig {
			scannersConfig[k] = pluginsrunnerconfig.Config{
				ImageName:     *v.ImageName,
				ScannerConfig: *v.Config,
			}
		}

		c.Plugins = plugins.Config{
			Enabled:        true,
			ScannersList:   *config.ScannersList,
			Inputs:         nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: &scannersConfig,
		}
	}
}

func NewFamiliesConfigFrom(config *ScannerConfig, sfc *apitypes.ScanFamiliesConfig) *families.Config {
	c := families.NewConfig()

	opts := []FamiliesConfigOption{
		withSBOM(sfc.Sbom, config),
		withVulnerabilities(sfc.Vulnerabilities, config),
		withSecretsConfig(sfc.Secrets, config),
		withExploitsConfig(sfc.Exploits, config),
		withMalwareConfig(sfc.Malware, config),
		withMisconfigurationConfig(sfc.Misconfigurations, config),
		withRootkitsConfig(sfc.Rootkits, config),
		withInfoFinderConfig(sfc.InfoFinder, config),
		withPluginsConfig(sfc.Plugins, config),
	}

	for _, o := range opts {
		o(c)
	}

	return c
}
