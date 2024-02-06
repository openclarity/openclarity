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

package assetscan

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestUnmarshalDeleteJobPolicy(t *testing.T) {
	tests := []struct {
		Name             string
		DeletePolicyText string

		ExpectedNewErrorMatcher types.GomegaMatcher
		ExpectedDeleteJobPolicy DeleteJobPolicyType
	}{
		{
			Name:                    "Always",
			DeletePolicyText:        "alWaYs",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedDeleteJobPolicy: DeleteJobPolicyAlways,
		},
		{
			Name:                    "Never",
			DeletePolicyText:        "nEvEr",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedDeleteJobPolicy: DeleteJobPolicyNever,
		},
		{
			Name:                    "OnSuccess",
			DeletePolicyText:        "onsUCCESS",
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedDeleteJobPolicy: DeleteJobPolicyOnSuccess,
		},
		{
			Name:                    "Invalid",
			DeletePolicyText:        "super awesome delete policy",
			ExpectedNewErrorMatcher: HaveOccurred(),
			ExpectedDeleteJobPolicy: DeleteJobPolicyType("invalid"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			policy := DeleteJobPolicyType("invalid")
			err := policy.UnmarshalText([]byte(test.DeletePolicyText))

			g.Expect(err).Should(test.ExpectedNewErrorMatcher)
			g.Expect(policy).Should(BeEquivalentTo(test.ExpectedDeleteJobPolicy))
		})
	}
}
