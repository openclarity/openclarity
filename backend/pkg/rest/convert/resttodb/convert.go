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

package resttodb

import (
	"encoding/json"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/database"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func ConvertScanConfig(config *models.ScanConfig, id string) (*database.ScanConfig, error) {
	var ret database.ScanConfig
	var err error
	var scanConfigUUID uuid.UUID

	if id != "" {
		scanConfigUUID, err = uuid.FromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scanConfigID %v to uuid: %v", id, err)
		}
	}

	if config.ScanFamiliesConfig != nil {
		ret.ScanFamiliesConfig, err = json.Marshal(config.ScanFamiliesConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if config.Scope != nil {
		ret.Scope, err = config.Scope.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if config.Scheduled != nil {
		ret.Scheduled, err = config.Scheduled.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	ret.Name = config.Name

	ret.Base = database.Base{ID: scanConfigUUID}

	return &ret, nil
}

func ConvertTarget(target *models.Target, id string) (*database.Target, error) {
	var targetUUID uuid.UUID
	var err error
	if id != "" {
		targetUUID, err = uuid.FromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert targetID %v to uuid: %v", id, err)
		}
	}
	disc, err := target.TargetInfo.Discriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get discriminator: %w", err)
	}
	switch disc {
	case "VMInfo":
		vminfo, err := target.TargetInfo.AsVMInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to convert target to vm info: %w", err)
		}
		var provider *string
		if vminfo.InstanceProvider != nil {
			provider = utils.StringPtr(string(*vminfo.InstanceProvider))
		}
		return &database.Target{
			Base: database.Base{
				ID: targetUUID,
			},
			Type:             vminfo.ObjectType,
			Location:         &vminfo.Location,
			InstanceID:       utils.StringPtr(vminfo.InstanceID),
			InstanceProvider: provider,
		}, nil
	default:
		return nil, fmt.Errorf("unknown target type: %v", disc)
	}
}

// nolint:cyclop
func ConvertScanResult(result *models.TargetScanResult, id string) (*database.ScanResult, error) {
	var ret database.ScanResult
	var err error
	var scanResultUUID uuid.UUID

	if id != "" {
		scanResultUUID, err = uuid.FromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scanResultID %v to uuid: %v", id, err)
		}
	}
	ret.ScanID = result.ScanId
	ret.TargetID = result.TargetId

	if result.Exploits != nil {
		ret.Exploits, err = json.Marshal(result.Exploits)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}
	if result.Malware != nil {
		ret.Malware, err = json.Marshal(result.Malware)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}
	if result.Misconfigurations != nil {
		ret.Misconfigurations, err = json.Marshal(result.Misconfigurations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}
	if result.Rootkits != nil {
		ret.Rootkits, err = json.Marshal(result.Rootkits)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}
	if result.Sboms != nil {
		ret.Sboms, err = json.Marshal(result.Sboms)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if result.Secrets != nil {
		ret.Secrets, err = json.Marshal(result.Secrets)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}
	if result.Status != nil {
		ret.Status, err = json.Marshal(result.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}
	if result.Vulnerabilities != nil {
		ret.Vulnerabilities, err = json.Marshal(result.Vulnerabilities)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	ret.Base = database.Base{ID: scanResultUUID}

	return &ret, nil
}

func ConvertScan(scan *models.Scan, id string) (*database.Scan, error) {
	var ret database.Scan
	var err error
	var scanUUID uuid.UUID

	if id != "" {
		scanUUID, err = uuid.FromString(id)
		if err != nil {
			return nil, fmt.Errorf("failed to convert scanID %v to uuid: %v", id, err)
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
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	if scan.TargetIDs != nil {
		ret.TargetIDs, err = json.Marshal(scan.TargetIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
	}

	ret.Base = database.Base{ID: scanUUID}

	return &ret, nil
}
