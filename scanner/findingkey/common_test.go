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

package findingkey

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gotest.tools/v3/assert"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func TestGenerateFindingKey(t *testing.T) {
	rootkitFindingInfo := apitypes.RootkitFindingInfo{
		Message:     to.Ptr("Message"),
		RootkitName: to.Ptr("RootkitName"),
		RootkitType: to.Ptr(apitypes.RootkitType("RootkitType")),
	}
	exploitFindingInfo := apitypes.ExploitFindingInfo{
		CveID:       to.Ptr("CveID"),
		Description: to.Ptr("Description"),
		Name:        to.Ptr("Name"),
		SourceDB:    to.Ptr("SourceDB"),
		Title:       to.Ptr("Title"),
		Urls:        to.Ptr([]string{"url1", "url2"}),
	}
	vulFindingInfo := apitypes.VulnerabilityFindingInfo{
		Package: &apitypes.Package{
			Name:    to.Ptr("Package.Name"),
			Version: to.Ptr("Package.Version"),
		},
		VulnerabilityName: to.Ptr("VulnerabilityName"),
	}
	malwareFindingInfo := apitypes.MalwareFindingInfo{
		MalwareName: to.Ptr("MalwareName"),
		MalwareType: to.Ptr("MalwareType"),
		Path:        to.Ptr("Path"),
		RuleName:    to.Ptr("RuleName"),
	}
	miscFindingInfo := apitypes.MisconfigurationFindingInfo{
		Message:     to.Ptr("Message"),
		ScannerName: to.Ptr("ScannerName"),
		Id:          to.Ptr("Id"),
	}
	secretFindingInfo := apitypes.SecretFindingInfo{
		EndColumn:   to.Ptr(1),
		Fingerprint: to.Ptr("Fingerprint"),
		StartColumn: to.Ptr(2),
	}
	pkgFindingInfo := apitypes.PackageFindingInfo{
		Name:    to.Ptr("Name"),
		Version: to.Ptr("Version"),
	}

	type args struct {
		findingInfo *apitypes.FindingInfo
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "exploit",
			args: args{
				findingInfo: createFindingInfo(t, exploitFindingInfo),
			},
			want:    GenerateExploitKey(exploitFindingInfo).ExploitString(),
			wantErr: false,
		},
		{
			name: "vul",
			args: args{
				findingInfo: createFindingInfo(t, vulFindingInfo),
			},
			want:    GenerateVulnerabilityKey(vulFindingInfo).VulnerabilityString(),
			wantErr: false,
		},
		{
			name: "malware",
			args: args{
				findingInfo: createFindingInfo(t, malwareFindingInfo),
			},
			want:    GenerateMalwareKey(malwareFindingInfo).MalwareString(),
			wantErr: false,
		},
		{
			name: "misc",
			args: args{
				findingInfo: createFindingInfo(t, miscFindingInfo),
			},
			want:    GenerateMisconfigurationKey(miscFindingInfo).MisconfigurationString(),
			wantErr: false,
		},
		{
			name: "rootkit",
			args: args{
				findingInfo: createFindingInfo(t, rootkitFindingInfo),
			},
			want:    GenerateRootkitKey(rootkitFindingInfo).RootkitString(),
			wantErr: false,
		},
		{
			name: "secret",
			args: args{
				findingInfo: createFindingInfo(t, secretFindingInfo),
			},
			want:    GenerateSecretKey(secretFindingInfo).SecretString(),
			wantErr: false,
		},
		{
			name: "pkg",
			args: args{
				findingInfo: createFindingInfo(t, pkgFindingInfo),
			},
			want:    GeneratePackageKey(pkgFindingInfo).PackageString(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFindingKey(tt.args.findingInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFindingKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GenerateFindingKey() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func createFindingInfo(t *testing.T, info interface{}) *apitypes.FindingInfo {
	t.Helper()
	var err error
	findingInfoB := apitypes.FindingInfo{}
	switch fInfo := info.(type) {
	case apitypes.RootkitFindingInfo:
		err = findingInfoB.FromRootkitFindingInfo(fInfo)
	case apitypes.ExploitFindingInfo:
		err = findingInfoB.FromExploitFindingInfo(fInfo)
	case apitypes.SecretFindingInfo:
		err = findingInfoB.FromSecretFindingInfo(fInfo)
	case apitypes.MisconfigurationFindingInfo:
		err = findingInfoB.FromMisconfigurationFindingInfo(fInfo)
	case apitypes.MalwareFindingInfo:
		err = findingInfoB.FromMalwareFindingInfo(fInfo)
	case apitypes.VulnerabilityFindingInfo:
		err = findingInfoB.FromVulnerabilityFindingInfo(fInfo)
	case apitypes.PackageFindingInfo:
		err = findingInfoB.FromPackageFindingInfo(fInfo)
	}
	assert.NilError(t, err)
	return &findingInfoB
}
