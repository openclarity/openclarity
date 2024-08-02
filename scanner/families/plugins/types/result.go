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
	apitypes "github.com/openclarity/vmclarity/api/types"
	plugintypes "github.com/openclarity/vmclarity/plugins/sdk-go/types"
	"github.com/openclarity/vmclarity/scanner/families"
)

type Result struct {
	Metadata      families.FamilyMetadata `json:"Metadata"`
	Findings      []apitypes.FindingInfo  `json:"Findings"`
	PluginOutputs []plugintypes.Result    `json:"PluginOutputs"`
}

func NewResult() *Result {
	return &Result{}
}

func (r *Result) Merge(scan *ScannerResult) {
	if scan == nil {
		return
	}

	// Sync result data
	defer r.patchResult(scan)

	// Merge all findings
	r.Findings = append(r.Findings, scan.Findings...)
}

func (r *Result) patchResult(scan *ScannerResult) {
	// Update metadata
	r.Metadata.Scans = append(r.Metadata.Scans, scan.Metadata)
	r.Metadata.Summary = &families.FamilySummary{
		FindingsCount: len(r.Findings),
	}

	// Add scanner outputs
	r.PluginOutputs = append(r.PluginOutputs, scan.Output)
}
