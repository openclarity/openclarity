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

package scanconfig

import (
	"fmt"
	"time"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func newScanFromScanConfig(scanConfig *apitypes.ScanConfig) *apitypes.Scan {
	return &apitypes.Scan{
		Name: to.Ptr(fmt.Sprintf("%s-%s", *scanConfig.Name, scanConfig.Scheduled.OperationTime.Format(time.RFC3339))),
		ScanConfig: &apitypes.ScanConfigRelationship{
			Id: *scanConfig.Id,
		},
		AssetScanTemplate:   scanConfig.ScanTemplate.AssetScanTemplate,
		Scope:               scanConfig.ScanTemplate.Scope,
		MaxParallelScanners: scanConfig.ScanTemplate.MaxParallelScanners,
		TimeoutSeconds:      scanConfig.ScanTemplate.TimeoutSeconds,
		Status: apitypes.NewScanStatus(
			apitypes.ScanStatusStatePending,
			apitypes.ScanStatusReasonCreated,
			nil,
		),
		Summary: &apitypes.ScanSummary{
			JobsCompleted:          to.Ptr(0),
			JobsLeftToRun:          to.Ptr(0),
			TotalExploits:          to.Ptr(0),
			TotalMalware:           to.Ptr(0),
			TotalMisconfigurations: to.Ptr(0),
			TotalPackages:          to.Ptr(0),
			TotalRootkits:          to.Ptr(0),
			TotalSecrets:           to.Ptr(0),
			TotalInfoFinder:        to.Ptr(0),
			TotalVulnerabilities: &apitypes.VulnerabilitySeveritySummary{
				TotalCriticalVulnerabilities:   to.Ptr(0),
				TotalHighVulnerabilities:       to.Ptr(0),
				TotalLowVulnerabilities:        to.Ptr(0),
				TotalMediumVulnerabilities:     to.Ptr(0),
				TotalNegligibleVulnerabilities: to.Ptr(0),
			},
			TotalPlugins: to.Ptr(0),
		},
	}
}
