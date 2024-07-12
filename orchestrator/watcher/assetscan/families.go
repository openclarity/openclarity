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
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner"
	scannercommon "github.com/openclarity/vmclarity/scanner/common"
	exploitdbconfig "github.com/openclarity/vmclarity/scanner/families/exploits/exploitdb/config"
	exploits "github.com/openclarity/vmclarity/scanner/families/exploits/types"
	sshtopologyconfig "github.com/openclarity/vmclarity/scanner/families/infofinder/sshtopology/config"
	infofinder "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	clamconfig "github.com/openclarity/vmclarity/scanner/families/malware/clam/config"
	malware "github.com/openclarity/vmclarity/scanner/families/malware/types"
	yaraconfig "github.com/openclarity/vmclarity/scanner/families/malware/yara/config"
	cisdockerconfig "github.com/openclarity/vmclarity/scanner/families/misconfiguration/cisdocker/config"
	lynisconfig "github.com/openclarity/vmclarity/scanner/families/misconfiguration/lynis/config"
	misconfigurations "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	pluginsrunnerconfig "github.com/openclarity/vmclarity/scanner/families/plugins/runner/config"
	plugins "github.com/openclarity/vmclarity/scanner/families/plugins/types"
	chrootkitconfig "github.com/openclarity/vmclarity/scanner/families/rootkits/chkrootkit/config"
	rootkits "github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	syftconfig "github.com/openclarity/vmclarity/scanner/families/sbom/syft/config"
	trivyconfigsbom "github.com/openclarity/vmclarity/scanner/families/sbom/trivy/config"
	sbom "github.com/openclarity/vmclarity/scanner/families/sbom/types"
	gitleaksconfig "github.com/openclarity/vmclarity/scanner/families/secrets/gitleaks/config"
	secrets "github.com/openclarity/vmclarity/scanner/families/secrets/types"
	grypeconfig "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/grype/config"
	trivyconfig "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/trivy/config"
	vulnerabilities "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
)

type ScannerConfigOption func(*scanner.Config)

func withSBOM(config *apitypes.SBOMConfig, opts *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.SBOM = sbom.Config{
			Enabled:       true,
			AnalyzersList: config.GetAnalyzersList(),
			Inputs:        nil, // rootfs directory will be determined by the CLI after mount.
			Registry:      &scannercommon.Registry{},
			OutputFormat:  sbom.DefaultOutputFormat,
			AnalyzersConfig: sbom.AnalyzersConfig{
				Trivy: trivyconfigsbom.Config{
					Timeout: int(opts.TrivyScanTimeout / time.Second),
				},
				Syft: syftconfig.Config{
					Scope: "squashed",
				},
			},
		}
	}
}

func withVulnerabilities(config *apitypes.VulnerabilitiesConfig, opts *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		var grypeConfig grypeconfig.Config
		if opts.GrypeServerAddress != "" {
			grypeConfig = grypeconfig.Config{
				Mode: grypeconfig.ModeRemote,
				Remote: grypeconfig.RemoteGrypeConfig{
					GrypeServerAddress: opts.GrypeServerAddress,
					GrypeServerTimeout: opts.GrypeServerTimeout,
				},
			}
		} else {
			grypeConfig = grypeconfig.Config{
				Mode: grypeconfig.ModeLocal,
				Local: grypeconfig.LocalGrypeConfig{
					UpdateDB:           true,
					DBRootDir:          "/tmp/",
					ListingURL:         grypeconfig.DefaultListingURL,
					MaxAllowedBuiltAge: grypeconfig.DefaultMaxDatabaseAge,
					ListingFileTimeout: grypeconfig.DefaultListingFileTimeout,
					UpdateTimeout:      grypeconfig.DefaultUpdateTimeout,
					Scope:              string(source.SquashedScope),
				},
			}
		}

		c.Vulnerabilities = vulnerabilities.Config{
			Enabled:       true,
			ScannersList:  config.GetScannersList(),
			InputFromSbom: false, // will be determined by the CLI.
			ScannersConfig: vulnerabilities.ScannersConfig{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Grype: grypeConfig,
				Trivy: trivyconfig.Config{
					Timeout:    int(opts.TrivyScanTimeout / time.Second), // NOTE(chrisgacsal): Timeout is expected to be in seconds.
					ServerAddr: opts.TrivyServerAddress,
				},
			},
		}
	}
}

func withSecretsConfig(config *apitypes.SecretsConfig, _ *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Secrets = secrets.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: secrets.ScannersConfig{
				Gitleaks: gitleaksconfig.Config{},
			},
		}
	}
}

func withExploitsConfig(config *apitypes.ExploitsConfig, opts *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Exploits = exploits.Config{
			Enabled:       true,
			ScannersList:  config.GetScannersList(),
			InputFromVuln: true,
			ScannersConfig: exploits.ScannersConfig{
				ExploitDB: exploitdbconfig.Config{
					BaseURL: opts.ExploitsDBAddress,
				},
			},
		}
	}
}

func withMalwareConfig(config *apitypes.MalwareConfig, opts *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Malware = malware.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: malware.ScannersConfig{
				Clam: clamconfig.Config{
					// NOTE(ramizpolic): We disable scanning with daemon as we don't have proper
					// default configuration in place. Once we have defined valid default
					// configuration to use with clam daemon scan, we should re-enable this.
					// https://github.com/openclarity/vmclarity/issues/1870
					UseClamDaemon:                 false,
					AlternativeFreshclamMirrorURL: opts.AlternativeFreshclamMirrorURL,
				},
				Yara: yaraconfig.Config{
					CompiledRuleURL:   opts.YaraRuleServerAddress,
					DirectoriesToScan: config.GetYaraDirectoriesToScan(),
				},
			},
		}
	}
}

func withMisconfigurationConfig(config *apitypes.MisconfigurationsConfig, _ *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Misconfiguration = misconfigurations.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: misconfigurations.ScannersConfig{
				Lynis:     lynisconfig.Config{},
				CISDocker: cisdockerconfig.Config{},
			},
		}
	}
}

func withInfoFinderConfig(config *apitypes.InfoFinderConfig, _ *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.InfoFinder = infofinder.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: infofinder.ScannersConfig{
				SSHTopology: sshtopologyconfig.Config{},
			},
		}
	}
}

func withRootkitsConfig(config *apitypes.RootkitsConfig, _ *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Rootkits = rootkits.Config{
			Enabled:      true,
			ScannersList: config.GetScannersList(),
			Inputs:       nil,
			ScannersConfig: rootkits.ScannersConfig{
				Chkrootkit: chrootkitconfig.Config{},
			},
		}
	}
}

func withPluginsConfig(config *apitypes.PluginsConfig, _ *ScannerConfig) ScannerConfigOption {
	return func(c *scanner.Config) {
		if !config.IsEnabled() {
			return
		}

		scannersConfig := plugins.ScannersConfig{}
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
			BinaryMode:     to.ValueOrZero(config.BinaryMode),
			ScannersConfig: scannersConfig,
		}
	}
}

func NewScannerConfigFrom(config *ScannerConfig, sfc *apitypes.ScanFamiliesConfig) *scanner.Config {
	c := scanner.NewConfig()

	opts := []ScannerConfigOption{
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
