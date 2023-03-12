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

	uuid "github.com/satori/go.uuid"

	"github.com/openclarity/vmclarity/api/models"
)

// nolint:cyclop
func ConvertToDBScan(scan models.Scan) (Scan, error) {
	var ret Scan
	var err error
	var scanUUID uuid.UUID

	if scan.Id != nil {
		scanUUID, err = uuid.FromString(*scan.Id)
		if err != nil {
			return ret, fmt.Errorf("failed to convert scanID %v to uuid: %v", *scan.Id, err)
		}
	}

	if scan.ScanConfig != nil {
		ret.ScanConfigID = &scan.ScanConfig.Id
	}

	ret.ScanEndTime = scan.EndTime

	ret.ScanStartTime = scan.StartTime

	if scan.ScanConfigSnapshot != nil {
		ret.ScanConfigSnapshot, err = json.Marshal(scan.ScanConfigSnapshot)
		if err != nil {
			return ret, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if scan.State != nil {
		ret.State = string(*scan.State)
	}
	if scan.StateMessage != nil {
		ret.StateMessage = *scan.StateMessage
	}
	if scan.StateReason != nil {
		ret.StateReason = string(*scan.StateReason)
	}

	if scan.Summary != nil {
		ret.Summary, err = json.Marshal(scan.Summary)
		if err != nil {
			return ret, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if scan.TargetIDs != nil {
		ret.TargetIDs, err = json.Marshal(scan.TargetIDs)
		if err != nil {
			return ret, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if scan.Summary != nil {
		ret.Summary, err = json.Marshal(scan.Summary)
		if err != nil {
			return ret, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	ret.Base = Base{ID: scanUUID}

	return ret, nil
}
