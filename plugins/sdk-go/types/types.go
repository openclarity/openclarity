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

func Ptr[T any](value T) *T {
	return &value
}

func NewScannerStatus(s State, m *string) *Status {
	return &Status{
		State:              s,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}

// Export saves the data as JSON to the provided file.
func (r *Result) Export(outputFile string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	err = os.WriteFile(outputFile, data, 0o440 /* read only, owner & group */) //nolint:gosec,mnd
	if err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	return nil
}

// LoadFrom imports data from a JSON into the object.
func (r *Result) LoadFrom(outputFile string) error {
	contents, err := os.ReadFile(outputFile)
	if err != nil {
		return fmt.Errorf("failed to read result: %w", err)
	}

	var result Result
	if err := json.Unmarshal(contents, &result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	*r = result
	return nil
}
