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

package lsblk

import (
	_ "embed"
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

type Numbers struct {
	IntegerField Bytes `json:"integer_field"`
	StringField  Bytes `json:"string_field"`
}

// nolint:maintidx
func TestJSONUnmarshal(t *testing.T) {
	tests := []struct {
		Name string
		JSON []byte
		Err  error

		ExpectedErrorMatcher types.GomegaMatcher
		ExpectedNumbers      Numbers
	}{
		{
			Name:                 "JSON",
			JSON:                 []byte("{\"integer_field\": 100, \"string_field\": \"100\"}"),
			Err:                  nil,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedNumbers: Numbers{
				IntegerField: 100,
				StringField:  100,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			var numbers Numbers
			err := json.Unmarshal(test.JSON, &numbers)

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(numbers).Should(BeEquivalentTo(test.ExpectedNumbers))
		})
	}
}
