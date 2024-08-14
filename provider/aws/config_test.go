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

package aws

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	awstypes "github.com/openclarity/openclarity/provider/aws/types"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars map[string]string

		ExpectedNewErrorMatcher      types.GomegaMatcher
		ExpectedConfig               *Config
		ExpectedValidateErrorMatcher types.GomegaMatcher
	}{
		{
			Name: "New config from environment",
			EnvVars: map[string]string{
				"OPENCLARITY_AWS_SCANNER_REGION":                "eu-west-1",
				"OPENCLARITY_AWS_SUBNET_ID":                     "subnet-038f85dc621fd5b5d",
				"OPENCLARITY_AWS_SECURITY_GROUP_ID":             "sg-02cfdc854e18664d4",
				"OPENCLARITY_AWS_KEYPAIR_NAME":                  "vmclarity-ssh-key",
				"OPENCLARITY_AWS_SCANNER_IMAGE_NAME_FILTER":     "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-*",
				"OPENCLARITY_AWS_SCANNER_IMAGE_OWNERS":          "099720109477,amazon",
				"OPENCLARITY_AWS_SCANNER_INSTANCE_TYPE_MAPPING": "x86_64:t3.large,arm64:t4g.large",
				"OPENCLARITY_AWS_SCANNER_INSTANCE_ARCH_TO_USE":  "x86_64",
				"OPENCLARITY_AWS_BLOCK_DEVICE_NAME":             "xvdh",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ScannerRegion:          "eu-west-1",
				SubnetID:               "subnet-038f85dc621fd5b5d",
				SecurityGroupID:        "sg-02cfdc854e18664d4",
				KeyPairName:            "vmclarity-ssh-key",
				ScannerImageNameFilter: "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-*",
				ScannerImageOwners: []string{
					"099720109477",
					"amazon",
				},
				ScannerInstanceTypeMapping: awstypes.InstanceTypeMapping{
					"x86_64": "t3.large",
					"arm64":  "t4g.large",
				},
				ScannerInstanceArchitecture: "x86_64",
				BlockDeviceName:             "xvdh",
			},
			ExpectedValidateErrorMatcher: Not(HaveOccurred()),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			os.Clearenv()
			for k, v := range test.EnvVars {
				err := os.Setenv(k, v)
				g.Expect(err).Should(Not(HaveOccurred()))
			}

			config, err := NewConfig()

			g.Expect(err).Should(test.ExpectedNewErrorMatcher)
			g.Expect(config).Should(BeEquivalentTo(test.ExpectedConfig))

			err = config.Validate()
			g.Expect(err).Should(test.ExpectedValidateErrorMatcher)
		})
	}
}

func TestInstanceTypeMapping(t *testing.T) {
	tests := []struct {
		Name        string
		MappingText []byte

		ExpectedErrorMatcher        types.GomegaMatcher
		ExpectedInstanceTypeMapping *awstypes.InstanceTypeMapping
	}{
		{
			Name:        "Valid instance type mapping",
			MappingText: []byte("x86_64:t3.large,arm64:t4g.large"),

			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedInstanceTypeMapping: &awstypes.InstanceTypeMapping{
				"x86_64": "t3.large",
				"arm64":  "t4g.large",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			mapping := &awstypes.InstanceTypeMapping{}
			err := mapping.UnmarshalText(test.MappingText)

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(mapping).Should(BeEquivalentTo(test.ExpectedInstanceTypeMapping))
		})
	}
}
