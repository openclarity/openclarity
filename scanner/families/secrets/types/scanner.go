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
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
)

type ScannerResult struct {
	Metadata families.ScannerMetadata
	Findings []Finding
}

func NewScannerResult(findings []Finding) *ScannerResult {
	return &ScannerResult{
		Findings: findings,
	}
}

func (s *ScannerResult) PatchMetadata(scan common.ScanMetadata) {
	s.Metadata.Scan = &scan
	s.Metadata.Summary = &families.ScannerSummary{
		FindingsCount: len(s.Findings),
	}
}