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

package scanconfigwatcher

import (
	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

func newScanFromScanConfig(scanConfig *models.ScanConfig) *models.Scan {
	return &models.Scan{
		ScanConfig: &models.ScanConfigRelationship{
			Id: *scanConfig.Id,
		},
		ScanConfigSnapshot: &models.ScanConfigSnapshot{
			MaxParallelScanners: scanConfig.MaxParallelScanners,
			Name:                scanConfig.Name,
			ScanFamiliesConfig:  scanConfig.ScanFamiliesConfig,
			Scheduled:           scanConfig.Scheduled,
			Scope:               scanConfig.Scope,
			TimeoutSeconds:      scanConfig.TimeoutSeconds,
		},
		State: utils.PointerTo(models.ScanStatePending),
		Summary: &models.ScanSummary{
			JobsCompleted:          utils.PointerTo(0),
			JobsLeftToRun:          utils.PointerTo(0),
			TotalExploits:          utils.PointerTo(0),
			TotalMalware:           utils.PointerTo(0),
			TotalMisconfigurations: utils.PointerTo(0),
			TotalPackages:          utils.PointerTo(0),
			TotalRootkits:          utils.PointerTo(0),
			TotalSecrets:           utils.PointerTo(0),
			TotalVulnerabilities: &models.VulnerabilityScanSummary{
				TotalCriticalVulnerabilities:   utils.PointerTo(0),
				TotalHighVulnerabilities:       utils.PointerTo(0),
				TotalLowVulnerabilities:        utils.PointerTo(0),
				TotalMediumVulnerabilities:     utils.PointerTo(0),
				TotalNegligibleVulnerabilities: utils.PointerTo(0),
			},
		},
	}
}
