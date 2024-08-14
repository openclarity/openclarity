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
	apitypes "github.com/openclarity/openclarity/api/types"
	plugintypes "github.com/openclarity/openclarity/plugins/sdk-go/types"
	"github.com/openclarity/openclarity/scanner/common"
	"github.com/openclarity/openclarity/scanner/families"
)

type Result struct {
	Metadata      families.FamilyMetadata       `json:"Metadata"`
	Findings      []apitypes.FindingInfo        `json:"Findings"`
	PluginOutputs map[string]plugintypes.Result `json:"PluginOutputs"`
}

func NewResult() *Result {
	return &Result{
		PluginOutputs: make(map[string]plugintypes.Result),
	}
}

func (r *Result) Merge(scan common.ScanInfo, result *ScannerResult) {
	if result == nil {
		return
	}

	r.Metadata.Merge(families.ScannerMetadata{
		ScanInfo:      scan,
		TotalFindings: len(result.Findings),
	})

	r.Findings = append(r.Findings, result.Findings...)
	r.PluginOutputs[scan.ScannerName] = result.Output
}
