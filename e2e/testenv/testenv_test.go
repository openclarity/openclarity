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

package testenv

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	envtypes "github.com/openclarity/vmclarity/e2e/testenv/types"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars map[string]string

		ExpectedNewErrorMatcher      types.GomegaMatcher
		ExpectedConfig               *envtypes.Config
		ExpectedValidateErrorMatcher types.GomegaMatcher
	}{
		{
			Name: "Valid config",
			EnvVars: map[string]string{
				"VMCLARITY_E2E_PLATFORM":     "kubernetes",
				"VMCLARITY_E2E_USE_EXISTING": "true",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &envtypes.Config{
				Platform: envtypes.Kubernetes,
				ReuseEnv: true,
			},
			ExpectedValidateErrorMatcher: Not(HaveOccurred()),
		},
		{
			Name:                    "Default config",
			EnvVars:                 map[string]string{},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &envtypes.Config{
				Platform: envtypes.Docker,
				ReuseEnv: false,
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
		})
	}
}
