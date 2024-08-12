// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package analyzer

import (
	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/openclarity/vmclarity/scanner/utils"
)

type Results struct {
	Sbom         *cdx.BOM
	AnalyzerInfo string
	AppInfo      AppInfo
	Error        error
}

func (r *Results) GetError() error {
	return r.Error
}

type AppInfo struct {
	SourceMetadata map[string]string
	SourceType     utils.SourceType
	SourcePath     string
	SourceHash     string
}

func CreateResults(sbomBytes *cdx.BOM, analyzerName, userInput string, srcType utils.SourceType) *Results {
	return &Results{
		Sbom:         sbomBytes,
		AnalyzerInfo: analyzerName,
		AppInfo: AppInfo{
			SourceMetadata: map[string]string{},
			SourceType:     srcType,
			SourcePath:     userInput,
		},
	}
}
