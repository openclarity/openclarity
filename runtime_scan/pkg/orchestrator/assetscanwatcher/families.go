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
	"github.com/anchore/syft/syft/source"
	kubeclarityConfig "github.com/openclarity/kubeclarity/shared/pkg/config"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	exploitsCommon "github.com/openclarity/vmclarity/shared/pkg/families/exploits/common"
	exploitdbConfig "github.com/openclarity/vmclarity/shared/pkg/families/exploits/exploitdb/config"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	clamconfig "github.com/openclarity/vmclarity/shared/pkg/families/malware/clam/config"
	malwarecommon "github.com/openclarity/vmclarity/shared/pkg/families/malware/common"
	misconfiguration "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/shared/pkg/families/rootkits"
	chkrootkitConfig "github.com/openclarity/vmclarity/shared/pkg/families/rootkits/chkrootkit/config"
	rootkitsCommon "github.com/openclarity/vmclarity/shared/pkg/families/rootkits/common"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	secretscommon "github.com/openclarity/vmclarity/shared/pkg/families/secrets/common"
	gitleaksconfig "github.com/openclarity/vmclarity/shared/pkg/families/secrets/gitleaks/config"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
)

type FamiliesConfigOption func(*families.Config)

func withSBOM(config *models.SBOMConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.SBOM = sbom.Config{
			Enabled: true,
			// TODO(sambetts) This choice should come from the user's configuration
			AnalyzersList: []string{"syft", "trivy"},
			Inputs:        nil, // rootfs directory will be determined by the CLI after mount.
			AnalyzersConfig: &kubeclarityConfig.Config{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Registry: &kubeclarityConfig.Registry{},
				Analyzer: &kubeclarityConfig.Analyzer{
					OutputFormat: "cyclonedx",
					TrivyConfig: kubeclarityConfig.AnalyzerTrivyConfig{
						Timeout: int(opts.TrivyServerTimeout),
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
			Enabled: true,
			// TODO(sambetts) This choice should come from the user's configuration
			ScannersList:  []string{"grype", "trivy"},
			InputFromSbom: false, // will be determined by the CLI.
			ScannersConfig: &kubeclarityConfig.Config{
				// TODO(sambetts) The user needs to be able to provide this configuration
				Registry: &kubeclarityConfig.Registry{},
				Scanner: &kubeclarityConfig.Scanner{
					GrypeConfig: grypeConfig,
					TrivyConfig: kubeclarityConfig.ScannerTrivyConfig{
						Timeout:    int(opts.TrivyServerTimeout),
						ServerAddr: opts.TrivyServerAddress,
					},
				},
			},
		}
	}
}

func withSecretsConfig(config *models.SecretsConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Secrets = secrets.Config{
			Enabled: true,
			// TODO(idanf) This choice should come from the user's configuration
			ScannersList: []string{"gitleaks"},
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: &secretscommon.ScannersConfig{
				Gitleaks: gitleaksconfig.Config{
					BinaryPath: opts.GitleaksBinaryPath,
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
			ScannersList:  []string{"exploitdb"},
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
			ScannersList: []string{"clam"},
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: &malwarecommon.ScannersConfig{
				Clam: clamconfig.Config{
					ClamScanBinaryPath:            opts.ClamBinaryPath,
					FreshclamBinaryPath:           opts.FreshclamBinaryPath,
					AlternativeFreshclamMirrorURL: opts.AlternativeFreshclamMirrorURL,
				},
			},
		}
	}
}

func withMisconfigurationConfig(config *models.MisconfigurationsConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Misconfiguration = misconfiguration.Config{
			Enabled: true,
			// TODO(sambetts) This choice should come from the user's configuration
			ScannersList: []string{"lynis"},
			Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
			ScannersConfig: misconfiguration.ScannersConfig{
				// TODO(sambetts) Add scanner configurations here as we add them like Lynis
				Lynis: misconfiguration.LynisConfig{
					InstallPath: opts.LynisInstallPath,
				},
			},
		}
	}
}

func withRootkitsConfig(config *models.RootkitsConfig, opts *ScannerConfig) FamiliesConfigOption {
	return func(c *families.Config) {
		if !config.IsEnabled() {
			return
		}

		c.Rootkits = rootkits.Config{
			Enabled:      true,
			ScannersList: []string{"chkrootkit"},
			Inputs:       nil,
			ScannersConfig: &rootkitsCommon.ScannersConfig{
				Chkrootkit: chkrootkitConfig.Config{
					BinaryPath: opts.ChkrootkitBinaryPath,
				},
			},
		}
	}
}

func NewFamiliesConfigFrom(config *ScannerConfig, scanConfig *models.ScanConfigSnapshot) *families.Config {
	c := families.NewConfig()

	opts := []FamiliesConfigOption{
		withSBOM(scanConfig.ScanFamiliesConfig.Sbom, config),
		withVulnerabilities(scanConfig.ScanFamiliesConfig.Vulnerabilities, config),
		withSecretsConfig(scanConfig.ScanFamiliesConfig.Secrets, config),
		withExploitsConfig(scanConfig.ScanFamiliesConfig.Exploits, config),
		withMalwareConfig(scanConfig.ScanFamiliesConfig.Malware, config),
		withMisconfigurationConfig(scanConfig.ScanFamiliesConfig.Misconfigurations, config),
		withRootkitsConfig(scanConfig.ScanFamiliesConfig.Rootkits, config),
	}

	for _, o := range opts {
		o(c)
	}

	return c
}
