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

package cloudinit

import (
	_ "embed"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
)

//go:embed testdata/cloud-init.yaml
var ExpectedCloudInit string

//go:embed testdata/scanner-cli-config.yaml
var ScannerCLIConfig string

func TestNewCloudInit(t *testing.T) {
	tests := []struct {
		Name          string
		CloudInitData any

		ExpectedCloudInit string
	}{
		{
			Name: "Cloud-init from ScanJobConfig passed by pointer",
			CloudInitData: &provider.ScanJobConfig{
				ScannerImage:     "ghcr.io/openclarity/vmclarity-cli:latest",
				ScannerCLIConfig: ScannerCLIConfig,
				VMClarityAddress: "10.1.1.1:8888",
				ScanMetadata: provider.ScanMetadata{
					AssetScanID: "d6ff6f55-5d53-4934-bef5-c3abb70a7f76",
				},
			},
			ExpectedCloudInit: ExpectedCloudInit,
		},
		{
			Name: "Cloud-init from ScanJobConfig passed by value",
			CloudInitData: provider.ScanJobConfig{
				ScannerImage:     "ghcr.io/openclarity/vmclarity-cli:latest",
				ScannerCLIConfig: ScannerCLIConfig,
				VMClarityAddress: "10.1.1.1:8888",
				ScanMetadata: provider.ScanMetadata{
					AssetScanID: "d6ff6f55-5d53-4934-bef5-c3abb70a7f76",
				},
			},
			ExpectedCloudInit: ExpectedCloudInit,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result, err := New(test.CloudInitData)

			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(result).Should(MatchYAML(test.ExpectedCloudInit))
		})
	}
}
