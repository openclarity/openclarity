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

package scan_manager // nolint:revive,stylecheck

import (
	"context"
	"fmt"
	"github.com/openclarity/vmclarity/scanner/common"
	"time"

	"github.com/openclarity/vmclarity/scanner/families"
)

type (
	// ConfigType define families.Scanner configuration type
	ConfigType any

	// ResultType define families.Scanner scan result type
	ResultType any

	// NewScannerFunc defines a function that creates a new families.Scanner
	NewScannerFunc[CT ConfigType, RT ResultType] func(context.Context, string, CT) (families.Scanner[RT], error)
)

// ScanResult is result of a successfully scanned input by a specific scanner.
type ScanResult[RT ResultType] struct {
	ScanMetadata // embed and inherit methods
	Result       RT
}

// ScanMetadata is metadata of a successfully scanned input by a specific scanner.
type ScanMetadata struct {
	// Input meta
	InputPath      string
	InputType      common.InputType
	InputSize      int64
	StripInputPath *bool

	// Scanner meta
	ScannerName string
	StartTime   time.Time
	EndTime     time.Time
}

func (m ScanMetadata) String() string {
	return fmt.Sprintf("Scanner=%s Input=%s:%s InputSize=%d MB",
		m.ScannerName, m.InputType, m.InputPath, m.InputSize,
	)
}

func (m ScanMetadata) GetScanInputMetadata(totalFindings int) families.ScanInputMetadata {
	return families.ScanInputMetadata{
		ScannerName:   m.ScannerName,
		InputType:     m.InputType,
		InputPath:     m.InputPath,
		InputSize:     m.InputSize,
		StartTime:     m.StartTime,
		EndTime:       m.EndTime,
		TotalFindings: totalFindings,
	}
}
