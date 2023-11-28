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

// MisconfigurationKey One test can report multiple misconfigurations so we need to include the
// message in the unique key.
type MisconfigurationKey struct {
	ScannerName string
	TestID      string
	Message     string
}

// String returns an unique string representation of the misconfiguration finding.
func (k MisconfigurationKey) String() string {
	return fmt.Sprintf("%s.%s.%s", k.ScannerName, k.TestID, k.Message)
}

// MisconfigurationString returns an unique string representation of the misconfiguration independent of
// where the misconfiguration finding was found by the scanner.
func (k MisconfigurationKey) MisconfigurationString() string {
	return k.String()
}

func GenerateMisconfigurationKey(info models.MisconfigurationFindingInfo) MisconfigurationKey {
	return MisconfigurationKey{
		ScannerName: *info.ScannerName,
		TestID:      *info.TestID,
		Message:     *info.Message,
	}
}
