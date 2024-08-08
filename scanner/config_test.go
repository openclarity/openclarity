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
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	exploits "github.com/openclarity/vmclarity/scanner/families/exploits/types"
	infofinder "github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	malware "github.com/openclarity/vmclarity/scanner/families/malware/types"
	misconfigurations "github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	plugins "github.com/openclarity/vmclarity/scanner/families/plugins/types"
	rootkits "github.com/openclarity/vmclarity/scanner/families/rootkits/types"
	sbom "github.com/openclarity/vmclarity/scanner/families/sbom/types"
	secrets "github.com/openclarity/vmclarity/scanner/families/secrets/types"
	vulnerabilities "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
)

func Test_ConfigAddInputs(t *testing.T) {
	type args struct {
		inputs []string
		config *Config
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "sbom, vuls, secrets and malware are enabled",
			args: args{
				inputs: []string{"/mnt/snapshot1"},
				config: &Config{
					SBOM: sbom.Config{
						Enabled: true,
						Inputs:  nil,
					},
					Vulnerabilities: vulnerabilities.Config{
						Enabled:       true,
						Inputs:        nil,
						InputFromSbom: false,
					},
					Secrets: secrets.Config{
						Enabled: true,
						Inputs:  nil,
					},
					Malware: malware.Config{
						Enabled: true,
						Inputs:  nil,
					},
				},
			},
			want: &Config{
				SBOM: sbom.Config{
					Enabled: true,
					Inputs: []common.ScanInput{
						{
							Input:     "/mnt/snapshot1",
							InputType: common.ROOTFS,
						},
					},
				},
				Vulnerabilities: vulnerabilities.Config{
					Enabled:       true,
					InputFromSbom: true,
				},
				Secrets: secrets.Config{
					Enabled: true,
					Inputs: []common.ScanInput{
						{
							StripPathFromResult: to.Ptr(true),
							Input:               "/mnt/snapshot1",
							InputType:           common.ROOTFS,
						},
					},
				},
				Malware: malware.Config{
					Enabled: true,
					Inputs: []common.ScanInput{
						{
							StripPathFromResult: to.Ptr(true),
							Input:               "/mnt/snapshot1",
							InputType:           common.ROOTFS,
						},
					},
				},
			},
		},
		{
			name: "only vuls enabled",
			args: args{
				inputs: []string{"/mnt/snapshot1"},
				config: &Config{
					Vulnerabilities: vulnerabilities.Config{
						Enabled:       true,
						Inputs:        nil,
						InputFromSbom: false,
					},
				},
			},
			want: &Config{
				Vulnerabilities: vulnerabilities.Config{
					Enabled: true,
					Inputs: []common.ScanInput{
						{
							Input:     "/mnt/snapshot1",
							InputType: common.ROOTFS,
						},
					},
					InputFromSbom: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.config.AddInputs(common.ROOTFS, tt.args.inputs)
			got := tt.args.config
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddInputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NewConfig(t *testing.T) {
	// Define shared config for mountpoints replacer option
	mountpoints := []string{"/mount-a", "/mount-b"}
	mountpointsInputsHave := []common.ScanInput{
		{
			Input:     "/dir-to-scan",
			InputType: common.ROOTFS,
		},
		{
			Input:     "nginx:latest",
			InputType: common.IMAGE,
		},
	}
	mountpointsInputsWant := []common.ScanInput{
		{
			Input:     "/mount-a/dir-to-scan",
			InputType: common.ROOTFS,
		},
		{
			Input:     "/mount-b/dir-to-scan",
			InputType: common.ROOTFS,
		},
		{
			Input:     "nginx:latest",
			InputType: common.IMAGE,
		},
	}

	tests := []struct {
		name    string
		options []ConfigOption
		want    *Config
	}{
		{
			name:    "no options",
			options: nil,
			want:    &Config{},
		},
		{
			name: "base config",
			options: []ConfigOption{
				WithBaseConfig(Config{
					SBOM: sbom.Config{
						Enabled: true,
					},
				}),
			},
			want: &Config{
				SBOM: sbom.Config{
					Enabled: true,
				},
			},
		},
		{
			name: "base config with sbom family replacer function",
			options: []ConfigOption{
				WithBaseConfig(Config{
					SBOM: sbom.Config{
						Enabled: true,
						Inputs: []common.ScanInput{
							{
								Input: "",
							},
						},
					},
				}),
				WithFamilyInputsReplacer(families.SBOM, func(input common.ScanInput) []common.ScanInput {
					input.Input = "replaced"
					return []common.ScanInput{input}
				}),
			},
			want: &Config{
				SBOM: sbom.Config{
					Enabled: true,
					Inputs: []common.ScanInput{{
						Input: "replaced",
					}},
				},
			},
		},
		{
			name: "base config with mount points override",
			options: []ConfigOption{
				WithBaseConfig(Config{
					SBOM:             sbom.Config{Inputs: mountpointsInputsHave},
					Vulnerabilities:  vulnerabilities.Config{Inputs: mountpointsInputsHave},
					Secrets:          secrets.Config{Inputs: mountpointsInputsHave},
					Rootkits:         rootkits.Config{Inputs: mountpointsInputsHave},
					Malware:          malware.Config{Inputs: mountpointsInputsHave},
					Misconfiguration: misconfigurations.Config{Inputs: mountpointsInputsHave},
					InfoFinder:       infofinder.Config{Inputs: mountpointsInputsHave},
					Exploits:         exploits.Config{Inputs: mountpointsInputsHave},
					Plugins:          plugins.Config{Inputs: mountpointsInputsHave},
				}),
				WithInputsMountOverride(mountpoints...),
			},
			want: &Config{
				SBOM:             sbom.Config{Inputs: mountpointsInputsWant},
				Vulnerabilities:  vulnerabilities.Config{Inputs: mountpointsInputsWant},
				Secrets:          secrets.Config{Inputs: mountpointsInputsWant},
				Rootkits:         rootkits.Config{Inputs: mountpointsInputsWant},
				Malware:          malware.Config{Inputs: mountpointsInputsWant},
				Misconfiguration: misconfigurations.Config{Inputs: mountpointsInputsWant},
				InfoFinder:       infofinder.Config{Inputs: mountpointsInputsWant},
				Exploits:         exploits.Config{Inputs: mountpointsInputsWant},
				Plugins:          plugins.Config{Inputs: mountpointsInputsWant},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConfig(tt.options...)
			if !cmp.Equal(got, tt.want, cmpopts.EquateEmpty()) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
