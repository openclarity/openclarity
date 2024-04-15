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
	"reflect"
	"testing"

	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/families/malware"
	"github.com/openclarity/vmclarity/scanner/families/sbom"
	"github.com/openclarity/vmclarity/scanner/families/secrets"
	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities"
	"github.com/openclarity/vmclarity/scanner/utils"
)

func Test_SetMountPointsForFamiliesInput(t *testing.T) {
	type args struct {
		mountPoints    []string
		familiesConfig *Config
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "sbom, vuls, secrets and malware are enabled",
			args: args{
				mountPoints: []string{"/mnt/snapshot1"},
				familiesConfig: &Config{
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
					Inputs: []types.Input{
						{
							Input:     "/mnt/snapshot1",
							InputType: string(utils.ROOTFS),
						},
					},
				},
				Vulnerabilities: vulnerabilities.Config{
					Enabled:       true,
					InputFromSbom: true,
				},
				Secrets: secrets.Config{
					Enabled: true,
					Inputs: []types.Input{
						{
							StripPathFromResult: to.Ptr(true),
							Input:               "/mnt/snapshot1",
							InputType:           string(utils.ROOTFS),
						},
					},
				},
				Malware: malware.Config{
					Enabled: true,
					Inputs: []types.Input{
						{
							StripPathFromResult: to.Ptr(true),
							Input:               "/mnt/snapshot1",
							InputType:           string(utils.ROOTFS),
						},
					},
				},
			},
		},
		{
			name: "only vuls enabled",
			args: args{
				mountPoints: []string{"/mnt/snapshot1"},
				familiesConfig: &Config{
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
					Inputs: []types.Input{
						{
							Input:     "/mnt/snapshot1",
							InputType: string(utils.ROOTFS),
						},
					},
					InputFromSbom: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetMountPointsForFamiliesInput(tt.args.mountPoints, tt.args.familiesConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetMountPointsForFamiliesInput() = %v, want %v", got, tt.want)
			}
		})
	}
}
