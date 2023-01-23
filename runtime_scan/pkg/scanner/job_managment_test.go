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

package scanner

import (
	"testing"

	"github.com/anchore/syft/syft/source"
	"github.com/google/go-cmp/cmp"
	kubeclarityConfig "github.com/openclarity/kubeclarity/shared/pkg/config"
	kubeclarityUtils "github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/api/models"
	familiesSbom "github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	familiesVulnerabilities "github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func Test_userSBOMConfigToFamiliesSbomConfig(t *testing.T) {
	type args struct {
		sbomConfig        *models.SBOMConfig
		scanRootDirectory string
	}
	type returns struct {
		config familiesSbom.Config
	}
	tests := []struct {
		name string
		args args
		want returns
	}{
		{
			name: "No SBOM Config",
			args: args{
				sbomConfig:        nil,
				scanRootDirectory: "/test",
			},
			want: returns{
				config: familiesSbom.Config{},
			},
		},
		{
			name: "Missing Enabled",
			args: args{
				sbomConfig:        &models.SBOMConfig{},
				scanRootDirectory: "/test",
			},
			want: returns{
				config: familiesSbom.Config{},
			},
		},
		{
			name: "Disabled",
			args: args{
				sbomConfig: &models.SBOMConfig{
					Enabled: utils.BoolPtr(false),
				},
				scanRootDirectory: "/test",
			},
			want: returns{
				config: familiesSbom.Config{},
			},
		},
		{
			name: "Enabled",
			args: args{
				sbomConfig: &models.SBOMConfig{
					Enabled: utils.BoolPtr(true),
				},
				scanRootDirectory: "/test",
			},
			want: returns{
				config: familiesSbom.Config{
					Enabled:       true,
					AnalyzersList: []string{"syft", "trivy"},
					Inputs: []familiesSbom.Input{
						{
							Input:     "/test",
							InputType: string(kubeclarityUtils.ROOTFS),
						},
					},
					AnalyzersConfig: &kubeclarityConfig.Config{
						Registry: &kubeclarityConfig.Registry{},
						Analyzer: &kubeclarityConfig.Analyzer{
							OutputFormat: "cyclonedx",
							TrivyConfig: kubeclarityConfig.AnalyzerTrivyConfig{
								Timeout: TrivyTimeout,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := userSBOMConfigToFamiliesSbomConfig(tt.args.sbomConfig, tt.args.scanRootDirectory)
			if diff := cmp.Diff(tt.want.config, got); diff != "" {
				t.Errorf("userSBOMConfigToFamiliesSbomConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_userVulnConfigToFamiliesVulnConfig(t *testing.T) {
	type args struct {
		vulnerabilitiesConfig *models.VulnerabilitiesConfig
	}
	type returns struct {
		config familiesVulnerabilities.Config
	}
	tests := []struct {
		name string
		args args
		want returns
	}{
		{
			name: "No Vulnerability Config",
			args: args{
				vulnerabilitiesConfig: nil,
			},
			want: returns{
				config: familiesVulnerabilities.Config{},
			},
		},
		{
			name: "Missing Enabled",
			args: args{
				vulnerabilitiesConfig: &models.VulnerabilitiesConfig{},
			},
			want: returns{
				config: familiesVulnerabilities.Config{},
			},
		},
		{
			name: "Disabled",
			args: args{
				vulnerabilitiesConfig: &models.VulnerabilitiesConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: returns{
				config: familiesVulnerabilities.Config{},
			},
		},
		{
			name: "Enabled",
			args: args{
				vulnerabilitiesConfig: &models.VulnerabilitiesConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: returns{
				config: familiesVulnerabilities.Config{
					Enabled: true,
					// TODO(sambetts) This choice should come from the user's configuration
					ScannersList:  []string{"grype", "trivy"},
					InputFromSbom: true,
					ScannersConfig: &kubeclarityConfig.Config{
						// TODO(sambetts) The user needs to be able to provide this configuration
						Registry: &kubeclarityConfig.Registry{},
						Scanner: &kubeclarityConfig.Scanner{
							GrypeConfig: kubeclarityConfig.GrypeConfig{
								// TODO(sambetts) Should run grype in remote mode eventually
								Mode: kubeclarityConfig.ModeLocal,
								LocalGrypeConfig: kubeclarityConfig.LocalGrypeConfig{
									UpdateDB:   true,
									DBRootDir:  "/tmp/",
									ListingURL: "https://toolbox-data.anchore.io/grype/databases/listing.json",
									Scope:      source.SquashedScope,
								},
							},
							TrivyConfig: kubeclarityConfig.ScannerTrivyConfig{
								Timeout: TrivyTimeout,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := userVulnConfigToFamiliesVulnConfig(tt.args.vulnerabilitiesConfig)
			if diff := cmp.Diff(tt.want.config, got); diff != "" {
				t.Errorf("userVulnConfigToFamiliesVulnConfig() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusVulnerabilitiesStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.VulnerabilitiesConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.VulnerabilitiesConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.VulnerabilitiesConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.VulnerabilitiesConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusVulnerabilitiesStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusVulnerabilitiesStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusSecretsStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.SecretsConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.SecretsConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.SecretsConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.SecretsConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusSecretsStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusSecretsStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusSbomStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.SBOMConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.SBOMConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.SBOMConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.SBOMConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusSbomStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusSbomStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusRootkitsStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.RootkitsConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.RootkitsConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.RootkitsConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.RootkitsConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusRootkitsStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusRootkitsStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusMisconfigurationsStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.MisconfigurationsConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.MisconfigurationsConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.MisconfigurationsConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.MisconfigurationsConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusMisconfigurationsStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusMisconfigurationsStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusMalwareStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.MalwareConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.MalwareConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.MalwareConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.MalwareConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusMalwareStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusMalwareStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getInitScanStatusExploitsStateFromEnabled(t *testing.T) {
	type args struct {
		config *models.ExploitsConfig
	}
	tests := []struct {
		name string
		args args
		want *models.TargetScanStateState
	}{
		{
			name: "enabled",
			args: args{
				config: &models.ExploitsConfig{
					Enabled: utils.BoolPtr(true),
				},
			},
			want: stateToPointer(models.INIT),
		},
		{
			name: "disabled",
			args: args{
				config: &models.ExploitsConfig{
					Enabled: utils.BoolPtr(false),
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil enabled",
			args: args{
				config: &models.ExploitsConfig{
					Enabled: nil,
				},
			},
			want: stateToPointer(models.NOTSCANNED),
		},
		{
			name: "nil config",
			args: args{
				config: nil,
			},
			want: stateToPointer(models.NOTSCANNED),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInitScanStatusExploitsStateFromEnabled(tt.args.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getInitScanStatusExploitsStateFromEnabled() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
