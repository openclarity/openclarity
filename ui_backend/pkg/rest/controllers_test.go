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

package rest

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	backendmodels "github.com/openclarity/vmclarity/api/models"

	"github.com/openclarity/vmclarity/shared/pkg/utils"
	"github.com/openclarity/vmclarity/ui_backend/api/models"

	"gotest.tools/v3/assert"
)

func Test_getTargetLocation(t *testing.T) {
	targetInfo := backendmodels.TargetType{}
	err := targetInfo.FromVMInfo(backendmodels.VMInfo{
		Location: "us-east-1",
	})
	assert.NilError(t, err)
	nonSupportedTargetInfo := backendmodels.TargetType{}
	err = nonSupportedTargetInfo.FromDirInfo(backendmodels.DirInfo{})
	assert.NilError(t, err)

	type args struct {
		target backendmodels.Target
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "sanity",
			args: args{
				target: backendmodels.Target{
					TargetInfo: &targetInfo,
				},
			},
			want:    "us-east-1",
			wantErr: false,
		},
		{
			name: "non supported target",
			args: args{
				target: backendmodels.Target{
					TargetInfo: &nonSupportedTargetInfo,
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTargetLocation(tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTargetLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getTargetLocation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_addTargetSummaryToFindingsCount(t *testing.T) {
	type args struct {
		findingsCount *models.FindingsCount
		summary       *backendmodels.ScanFindingsSummary
	}
	tests := []struct {
		name string
		args args
		want *models.FindingsCount
	}{
		{
			name: "nil",
			args: args{
				findingsCount: &models.FindingsCount{
					Exploits:          utils.PointerTo(1),
					Malware:           utils.PointerTo(2),
					Misconfigurations: utils.PointerTo(3),
					Rootkits:          utils.PointerTo(4),
					Secrets:           utils.PointerTo(5),
					Vulnerabilities:   utils.PointerTo(6),
				},
				summary: nil,
			},
			want: &models.FindingsCount{
				Exploits:          utils.PointerTo(1),
				Malware:           utils.PointerTo(2),
				Misconfigurations: utils.PointerTo(3),
				Rootkits:          utils.PointerTo(4),
				Secrets:           utils.PointerTo(5),
				Vulnerabilities:   utils.PointerTo(6),
			},
		},
		{
			name: "sanity",
			args: args{
				findingsCount: &models.FindingsCount{
					Exploits:          utils.PointerTo(1),
					Malware:           utils.PointerTo(2),
					Misconfigurations: utils.PointerTo(3),
					Rootkits:          utils.PointerTo(4),
					Secrets:           utils.PointerTo(5),
					Vulnerabilities:   utils.PointerTo(6),
				},
				summary: &backendmodels.ScanFindingsSummary{
					TotalExploits:          utils.PointerTo(2),
					TotalMalware:           utils.PointerTo(3),
					TotalMisconfigurations: utils.PointerTo(4),
					TotalPackages:          utils.PointerTo(5),
					TotalRootkits:          utils.PointerTo(6),
					TotalSecrets:           utils.PointerTo(7),
					TotalVulnerabilities: &backendmodels.VulnerabilityScanSummary{
						TotalCriticalVulnerabilities:   utils.PointerTo(10),
						TotalHighVulnerabilities:       utils.PointerTo(11),
						TotalLowVulnerabilities:        utils.PointerTo(12),
						TotalMediumVulnerabilities:     utils.PointerTo(13),
						TotalNegligibleVulnerabilities: utils.PointerTo(14),
					},
				},
			},
			want: &models.FindingsCount{
				Exploits:          utils.PointerTo(3),
				Malware:           utils.PointerTo(5),
				Misconfigurations: utils.PointerTo(7),
				Rootkits:          utils.PointerTo(10),
				Secrets:           utils.PointerTo(12),
				Vulnerabilities:   utils.PointerTo(66),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addTargetSummaryToFindingsCount(tt.args.findingsCount, tt.args.summary); !reflect.DeepEqual(got, tt.want) {
				gotB, _ := json.Marshal(got)
				wantB, _ := json.Marshal(tt.want)
				t.Errorf("addTargetSummaryToFindingsCount() = %v, want %v", string(gotB), string(wantB))
			}
		})
	}
}

func Test_getTotalVulnerabilities(t *testing.T) {
	type args struct {
		summary *backendmodels.VulnerabilityScanSummary
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "sanity",
			args: args{
				summary: &backendmodels.VulnerabilityScanSummary{
					TotalCriticalVulnerabilities:   utils.PointerTo(1),
					TotalHighVulnerabilities:       utils.PointerTo(2),
					TotalLowVulnerabilities:        utils.PointerTo(3),
					TotalMediumVulnerabilities:     utils.PointerTo(4),
					TotalNegligibleVulnerabilities: utils.PointerTo(5),
				},
			},
			want: 15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTotalVulnerabilities(tt.args.summary); got != tt.want {
				t.Errorf("getTotalVulnerabilities() = %v, want %v", got, tt.want)
			}
		})
	}
}

// nolint:errcheck
func Test_createRegionFindingsFromTargets(t *testing.T) {
	dirTarget := backendmodels.TargetType{}
	dirTarget.FromDirInfo(backendmodels.DirInfo{
		DirName:  utils.PointerTo("test-name"),
		Location: utils.PointerTo("location-test"),
	})

	vmFromRegion1 := backendmodels.TargetType{}
	vmFromRegion1.FromVMInfo(backendmodels.VMInfo{
		Location: "region1",
	})
	vm1FromRegion2 := backendmodels.TargetType{}
	vm1FromRegion2.FromVMInfo(backendmodels.VMInfo{
		Location: "region2",
	})
	type args struct {
		targets *backendmodels.Targets
	}
	tests := []struct {
		name string
		args args
		want []models.RegionFindings
	}{
		{
			name: "Unsupported target is skipped",
			args: args{
				targets: &backendmodels.Targets{
					Count: utils.PointerTo(1),
					Items: utils.PointerTo([]backendmodels.Target{
						{
							Summary: &backendmodels.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(1),
								TotalMalware:           utils.PointerTo(1),
								TotalMisconfigurations: utils.PointerTo(1),
								TotalPackages:          utils.PointerTo(1),
								TotalRootkits:          utils.PointerTo(1),
								TotalSecrets:           utils.PointerTo(1),
								TotalVulnerabilities: &backendmodels.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(1),
									TotalHighVulnerabilities:       utils.PointerTo(1),
									TotalLowVulnerabilities:        utils.PointerTo(1),
									TotalMediumVulnerabilities:     utils.PointerTo(1),
									TotalNegligibleVulnerabilities: utils.PointerTo(1),
								},
							},
							TargetInfo: &dirTarget,
						},
					}),
				},
			},
			want: []models.RegionFindings{},
		},
		{
			name: "sanity",
			args: args{
				targets: &backendmodels.Targets{
					Count: utils.PointerTo(3),
					Items: &[]backendmodels.Target{
						{
							Summary: &backendmodels.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(1),
								TotalMalware:           utils.PointerTo(2),
								TotalMisconfigurations: utils.PointerTo(3),
								TotalPackages:          utils.PointerTo(4),
								TotalRootkits:          utils.PointerTo(5),
								TotalSecrets:           utils.PointerTo(6),
								TotalVulnerabilities: &backendmodels.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(7),
									TotalHighVulnerabilities:       utils.PointerTo(8),
									TotalLowVulnerabilities:        utils.PointerTo(9),
									TotalMediumVulnerabilities:     utils.PointerTo(10),
									TotalNegligibleVulnerabilities: utils.PointerTo(11),
								},
							},
							TargetInfo: &vmFromRegion1,
						},
						{
							Summary: &backendmodels.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(2),
								TotalMalware:           utils.PointerTo(3),
								TotalMisconfigurations: utils.PointerTo(4),
								TotalPackages:          utils.PointerTo(5),
								TotalRootkits:          utils.PointerTo(6),
								TotalSecrets:           utils.PointerTo(7),
								TotalVulnerabilities: &backendmodels.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(8),
									TotalHighVulnerabilities:       utils.PointerTo(9),
									TotalLowVulnerabilities:        utils.PointerTo(10),
									TotalMediumVulnerabilities:     utils.PointerTo(11),
									TotalNegligibleVulnerabilities: utils.PointerTo(12),
								},
							},
							TargetInfo: &vmFromRegion1,
						},
						{
							Summary: &backendmodels.ScanFindingsSummary{
								TotalExploits:          utils.PointerTo(3),
								TotalMalware:           utils.PointerTo(4),
								TotalMisconfigurations: utils.PointerTo(5),
								TotalPackages:          utils.PointerTo(6),
								TotalRootkits:          utils.PointerTo(7),
								TotalSecrets:           utils.PointerTo(8),
								TotalVulnerabilities: &backendmodels.VulnerabilityScanSummary{
									TotalCriticalVulnerabilities:   utils.PointerTo(9),
									TotalHighVulnerabilities:       utils.PointerTo(10),
									TotalLowVulnerabilities:        utils.PointerTo(11),
									TotalMediumVulnerabilities:     utils.PointerTo(12),
									TotalNegligibleVulnerabilities: utils.PointerTo(13),
								},
							},
							TargetInfo: &vm1FromRegion2,
						},
					},
				},
			},
			want: []models.RegionFindings{
				{
					FindingsCount: &models.FindingsCount{
						Exploits:          utils.PointerTo(3),
						Malware:           utils.PointerTo(5),
						Misconfigurations: utils.PointerTo(7),
						Rootkits:          utils.PointerTo(11),
						Secrets:           utils.PointerTo(13),
						Vulnerabilities:   utils.PointerTo(95),
					},
					RegionName: utils.PointerTo("region1"),
				},
				{
					FindingsCount: &models.FindingsCount{
						Exploits:          utils.PointerTo(3),
						Malware:           utils.PointerTo(4),
						Misconfigurations: utils.PointerTo(5),
						Rootkits:          utils.PointerTo(7),
						Secrets:           utils.PointerTo(8),
						Vulnerabilities:   utils.PointerTo(55),
					},
					RegionName: utils.PointerTo("region2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createRegionFindingsFromTargets(tt.args.targets)
			if diff := cmp.Diff(tt.want, got, cmpopts.SortSlices(func(a, b models.RegionFindings) bool { return *a.RegionName < *b.RegionName })); diff != "" {
				t.Errorf("createRegionFindingsFromTargets() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
