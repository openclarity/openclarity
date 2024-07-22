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

import (
	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/openclarity/vmclarity/scanner/common"
)

type AppInfo struct {
	SourceMetadata map[string]string
	SourceType     common.InputType
	SourcePath     string
	SourceHash     string
}

type ScannerResult struct {
	Sbom         *cdx.BOM
	AnalyzerInfo string
	AppInfo      AppInfo
}

func (scan *ScannerResult) GetTotalFindings() int {
	if scan.Sbom != nil && scan.Sbom.Components != nil {
		return len(*scan.Sbom.Components)
	}

	return 0
}

func CreateScannerResult(sbom *cdx.BOM, analyzerName, userInput string, srcType common.InputType) *ScannerResult {
	return &ScannerResult{
		Sbom:         sbom,
		AnalyzerInfo: analyzerName,
		AppInfo: AppInfo{
			SourceMetadata: map[string]string{},
			SourceType:     srcType,
			SourcePath:     userInput,
		},
	}
}
