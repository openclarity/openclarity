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
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
)

type SecretKey struct {
	Fingerprint string
	StartColumn int
	EndColumn   int
}

// String returns an unique string representation of the secret finding.
func (k SecretKey) String() string {
	return fmt.Sprintf("%s.%d.%d", k.Fingerprint, k.StartColumn, k.EndColumn)
}

// SecretString returns an unique string representation of the secret independent of
// where the secret finding was found by the scanner.
func (k SecretKey) SecretString() string {
	return k.String()
}

func GenerateSecretKey(secret models.SecretFindingInfo) SecretKey {
	return SecretKey{
		Fingerprint: *secret.Fingerprint,
		StartColumn: *secret.StartColumn,
		EndColumn:   *secret.EndColumn,
	}
}
