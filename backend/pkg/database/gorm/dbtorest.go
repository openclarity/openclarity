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

package gorm

import (
	"encoding/json"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func ConvertToRestScan(scan Scan) (models.Scan, error) {
	var ret models.Scan

	if scan.ScanConfigSnapshot != nil {
		ret.ScanConfigSnapshot = &models.ScanConfigData{}
		if err := json.Unmarshal(scan.ScanConfigSnapshot, ret.ScanConfigSnapshot); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	if scan.Summary != nil {
		ret.Summary = &models.ScanSummary{}
		if err := json.Unmarshal(scan.Summary, ret.Summary); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	if scan.TargetIDs != nil {
		ret.TargetIDs = &[]string{}
		if err := json.Unmarshal(scan.TargetIDs, ret.TargetIDs); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}
	if scan.Summary != nil {
		ret.Summary = &models.ScanSummary{}
		if err := json.Unmarshal(scan.Summary, ret.Summary); err != nil {
			return ret, fmt.Errorf("failed to unmarshal json: %w", err)
		}
	}

	ret.Id = utils.StringPtr(scan.ID.String())
	ret.StartTime = scan.ScanStartTime
	ret.EndTime = scan.ScanEndTime

	if scan.ScanConfigID != nil {
		ret.ScanConfig = &models.ScanConfigRelationship{Id: *scan.ScanConfigID}
	}

	ret.State = utils.PointerTo[models.ScanState](models.ScanState(scan.State))
	ret.StateMessage = utils.PointerTo[string](scan.StateMessage)
	ret.StateReason = utils.PointerTo[models.ScanStateReason](models.ScanStateReason(scan.StateReason))

	return ret, nil
}

func ConvertToRestScans(scans []Scan) (models.Scans, error) {
	ret := models.Scans{
		Items: &[]models.Scan{},
	}

	for _, scan := range scans {
		sc, err := ConvertToRestScan(scan)
		if err != nil {
			return models.Scans{}, fmt.Errorf("failed to convert scan: %w", err)
		}
		*ret.Items = append(*ret.Items, sc)
	}

	ret.Total = utils.IntPtr(len(scans))

	return ret, nil
}
