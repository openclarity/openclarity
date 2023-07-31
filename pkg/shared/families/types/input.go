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

package types

import "time"

type Metadata struct {
	Timestamp  time.Time `json:"Timestamp"`
	Scanners   []string  `json:"Scanners"`
	InputScans []InputScanMetadata
}

type Input struct {
	// StripPathFromResult overrides global StripInputPaths value
	StripPathFromResult *bool  `yaml:"strip_path_from_result" mapstructure:"strip_path_from_result"`
	Input               string `yaml:"input" mapstructure:"input"`
	InputType           string `yaml:"input_type" mapstructure:"input_type"`
}

type InputScanMetadata struct {
	InputType     string
	InputPath     string
	InputSize     int64
	ScanStartTime time.Time
	ScanEndTime   time.Time
}

func CreateInputScanMetadata(startTime, endTime time.Time, inputSize int64, input Input) InputScanMetadata {
	return InputScanMetadata{
		InputType:     input.InputType,
		InputPath:     input.Input,
		InputSize:     inputSize,
		ScanStartTime: startTime,
		ScanEndTime:   endTime,
	}
}
