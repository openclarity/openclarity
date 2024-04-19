// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package types

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// APIVersion defines the current version of the Scanner Plugin API.
const APIVersion = "1.0.0"

func NewScannerStatus(s StatusState, m *string) *Status {
	return &Status{
		State:              s,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}

func Ptr[T any](value T) *T {
	return &value
}

// Export saves the data as JSON to the provided file.
func (r *Result) Export(outputFile string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	err = os.WriteFile(outputFile, data, 0o660 /* read & write, owner & group */) //nolint:gosec,gomnd
	if err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	return nil
}
