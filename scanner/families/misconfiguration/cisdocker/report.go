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

package cisdocker

import (
	dockle_types "github.com/Portshift/dockle/pkg/types"

	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	"github.com/openclarity/vmclarity/scanner/utils"
)

const (
	CISDockerImpactCategory = "best-practice"
)

var scanDirIgnoreIDs = map[string]struct{}{
	dockle_types.UseContentTrust: {},
}

func parseDockleReport(sourceType utils.SourceType, imageName string, assessmentMap dockle_types.AssessmentMap) []types.Misconfiguration {
	ret := make([]types.Misconfiguration, 0, len(assessmentMap))

	for _, codeInfo := range assessmentMap {
		if sourceType == utils.ROOTFS || sourceType == utils.DIR {
			if _, ok := scanDirIgnoreIDs[codeInfo.Code]; ok {
				continue
			}
		}

		severity := convertDockleSeverity(codeInfo.Level)
		if severity == "" {
			// skip when no severity
			continue
		}

		description := ""
		for _, assessment := range codeInfo.Assessments {
			description += assessment.Desc + "\n"
		}

		ret = append(ret, types.Misconfiguration{
			Location:    imageName,
			Category:    CISDockerImpactCategory,
			ID:          codeInfo.Code,
			Description: description,
			Severity:    severity,
			Message:     dockle_types.TitleMap[codeInfo.Code],
		})
	}

	return ret
}

func convertDockleSeverity(level int) types.Severity {
	switch level {
	case dockle_types.FatalLevel:
		return types.HighSeverity
	case dockle_types.WarnLevel:
		return types.MediumSeverity
	case dockle_types.InfoLevel:
		return types.LowSeverity
	default: // ignore PassLevel, IgnoreLevel, SkipLevel
		return ""
	}
}
