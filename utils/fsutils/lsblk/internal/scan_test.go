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

package internal

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestScanPairs(t *testing.T) {
	tests := []struct {
		Name string
		Data []byte
		EOF  bool

		ExpectedAdvance      int
		ExpectedToken        []byte
		ExpectedErrorMatcher types.GomegaMatcher
	}{
		{
			Name:                 "Token with no space in value",
			Data:                 []byte("NAME=\"nvme0n1p15\" KNAME=\"nvme0n1p15\" PATH=\"/dev/nvme0n1p15\" MAJ:MIN=\"259:3\""),
			EOF:                  false,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedAdvance:      18,
			ExpectedToken:        []byte("NAME=\"nvme0n1p15\""),
		},
		{
			Name:                 "Token with space in value",
			Data:                 []byte("MODEL=\"Amazon Elastic Block Store\" SERIAL=\"vol071a923e1421f847f\""),
			EOF:                  false,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedAdvance:      35,
			ExpectedToken:        []byte("MODEL=\"Amazon Elastic Block Store\""),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			advance, token, err := ScanPairs(test.Data, test.EOF)

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(token).Should(Equal(test.ExpectedToken))
			g.Expect(advance).Should(Equal(test.ExpectedAdvance))
		})
	}
}
