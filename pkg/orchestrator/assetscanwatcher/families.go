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

package assetscanwatcher

import (
	"time"

	"github.com/anchore/syft/syft/source"
	kubeclarityConfig "github.com/openclarity/kubeclarity/shared/pkg/config"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/families"
	"github.com/openclarity/vmclarity/pkg/shared/families/exploits"
	exploitsCommon "github.com/openclarity/vmclarity/pkg/shared/families/exploits/common"
	exploitdbConfig "github.com/openclarity/vmclarity/pkg/shared/families/exploits/exploitdb/config"
	infofinderTypes "github.com/openclarity/vmclarity/pkg/shared/families/infofinder/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/malware"
	clamconfig "github.com/openclarity/vmclarity/pkg/shared/families/malware/clam/config"
	malwarecommon "github.com/openclarity/vmclarity/pkg/shared/families/malware/common"
	yaraconfig "github.com/openclarity/vmclarity/pkg/shared/families/malware/yara/config"
	misconfiguration "github.com/openclarity/vmclarity/pkg/shared/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits"
	chkrootkitConfig "github.com/openclarity/vmclarity/pkg/shared/families/rootkits/chkrootkit/config"
	rootkitsCommon "github.com/openclarity/vmclarity/pkg/shared/families/rootkits/common"
	"github.com/openclarity/vmclarity/pkg/shared/families/sbom"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets"
	secretscommon "github.com/openclarity/vmclarity/pkg/shared/families/secrets/common"
	gitleaksconfig "github.com/openclarity/vmclarity/pkg/shared/families/secrets/gitleaks/config"
	"github.com/openclarity/vmclarity/pkg/shared/families/vulnerabilities"
)

type FamiliesConfigOption func(*families.Config)

func withSBOM(config *models.SBOMConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.SBOM = sbom.Config{
			Enabled:       true,
			AnalyzersList: config.GetAnalyzersList(),
			Inputs:        nil, // rootfs directory will be determined by the CLI after mount.
			AnalyzersConfig: &kubeclarityConfig.Config{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Registry: &kubeclarityConfig.Registry{},
				Analyzer: &kubeclarityConfig.Analyzer{
					OutputFormat: "cyclonedx",
					TrivyConfig: kubeclarityConfig.AnalyzerTrivyConfig{
						Timeout: int(opts.TrivyScanTimeout / time.Second), // NOTE(chrisgacsal): Timeout is expected to be in seconds.
					},
				},
			},
		}
	}
}

func withVulnerabilities(config *models.VulnerabilitiesConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		var grypeConfig kubeclarityConfig.GrypeConfig
		if opts.GrypeServerAddress != "" {
			grypeConfig = kubeclarityConfig.GrypeConfig{
				Mode: kubeclarityConfig.ModeRemote,
				RemoteGrypeConfig: kubeclarityConfig.RemoteGrypeConfig{
					GrypeServerAddress: opts.GrypeServerAddress,
					GrypeServerTimeout: opts.GrypeServerTimeout,
				},
			}
		} else {
			grypeConfig = kubeclarityConfig.GrypeConfig{
				Mode: kubeclarityConfig.ModeLocal,
				LocalGrypeConfig: kubeclarityConfig.LocalGrypeConfig{
					UpdateDB:   true,
					DBRootDir:  "/tmp/",
					ListingURL: "https://toolbox-data.anchore.io/grype/databases/listing.json",
					Scope:      source.SquashedScope,
				},
			}
		}

		c.Vulnerabilities = vulnerabilities.Config{
			Enabled:       true,
			ScannersList:  config.GetScannersList(),
			InputFromSbom: false, // will be determined by the CLI.
			ScannersConfig: &kubeclarityConfig.Config{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Registry: &kubeclarityConfig.Registry{},
				Scanner: &kubeclarityConfig.Scanner{
					GrypeConfig: grypeConfig,
					TrivyConfig: kubeclarityConfig.ScannerTrivyConfig{
						Timeout:    int(opts.TrivyScanTimeout / time.Second), // NOTE(chrisgacsal): Timeout is expected to be in seconds.
						ServerAddr: opts.TrivyServerAddress,
					},
				},
			},
		}
	}
}

func withSecretsConfig(config *models.SecretsConfig, _ *ScannerConfig) FamiliesConfigOption {
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

func withExploitsConfig(config *models.ExploitsConfig, opts *ScannerConfig) FamiliesConfigOption {
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

func withMalwareConfig(config *models.MalwareConfig, opts *ScannerConfig) FamiliesConfigOption {
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
					YaraBinaryPath:  "",
					CompiledRuleURL: opts.YaraRuleServerAddress,
				},
			},
		}
	}
}

func withMisconfigurationConfig(config *models.MisconfigurationsConfig, _ *ScannerConfig) FamiliesConfigOption {
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
			},
		}
	}
}

func withInfoFinderConfig(config *models.InfoFinderConfig, _ *ScannerConfig) FamiliesConfigOption {
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

func withRootkitsConfig(config *models.RootkitsConfig, _ *ScannerConfig) FamiliesConfigOption {
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

func NewFamiliesConfigFrom(config *ScannerConfig, sfc *models.ScanFamiliesConfig) *families.Config {
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
	}

	for _, o := range opts {
		o(c)
	}

	return c
}
