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

package common

import (
	"fmt"
	"time"
)

type ScanInput struct {
	// StripPathFromResult overrides global StripInputPaths value
	StripPathFromResult *bool     `json:"strip_path_from_result" yaml:"strip_path_from_result" mapstructure:"strip_path_from_result"`
	Input               string    `json:"input" yaml:"input" mapstructure:"input"`
	InputType           InputType `json:"input_type" yaml:"input_type" mapstructure:"input_type"`
}

func (s ScanInput) String() string {
	return fmt.Sprintf("%s:%s", s.InputType, s.Input)
}

// ScanMetadata defines metadata for a successfully processed ScanInput.
type ScanMetadata struct {
	ScannerName string    `json:"scanner_name" yaml:"scanner_name" mapstructure:"scanner_name"`
	InputType   InputType `json:"input_type" yaml:"input_type" mapstructure:"input_type"`
	InputPath   string    `json:"input_path" yaml:"input_path" mapstructure:"input_path"`
	InputSize   int64     `json:"input_size" yaml:"input_size" mapstructure:"input_size"`
	StartTime   time.Time `json:"start_time" yaml:"start_time" mapstructure:"start_time"`
	EndTime     time.Time `json:"end_time" yaml:"end_time" mapstructure:"end_time"`
}

func (m *ScanMetadata) String() string {
	return fmt.Sprintf("Scanner=%s Input=%s:%s Size=%d MB",
		m.ScannerName, m.InputType, m.InputPath, m.InputSize,
	)
}
