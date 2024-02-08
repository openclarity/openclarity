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
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/server/common"
	"github.com/openclarity/vmclarity/api/server/database/types"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

const (
	scanSchemaName = "Scan"
)

type Scan struct {
	ODataObject
}

type ScansTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ScansTable() types.ScansTable {
	return &ScansTableHandler{
		DB: db.DB,
	}
}

func (s *ScansTableHandler) GetScans(params apitypes.GetScansParams) (apitypes.Scans, error) {
	var scans []Scan
	err := ODataQuery(s.DB, scanSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &scans)
	if err != nil {
		return apitypes.Scans{}, err
	}

	items := make([]apitypes.Scan, len(scans))
	for i, sc := range scans {
		var scan apitypes.Scan
		err = json.Unmarshal(sc.Data, &scan)
		if err != nil {
			return apitypes.Scans{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = scan
	}

	output := apitypes.Scans{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, scanSchemaName, params.Filter)
		if err != nil {
			return apitypes.Scans{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *ScansTableHandler) GetScan(scanID apitypes.ScanID, params apitypes.GetScansScanIDParams) (apitypes.Scan, error) {
	var dbScan Scan
	filter := fmt.Sprintf("id eq '%s'", scanID)
	err := ODataQuery(s.DB, scanSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbScan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apitypes.Scan{}, types.ErrNotFound
		}
		return apitypes.Scan{}, err
	}

	var apiScan apitypes.Scan
	err = json.Unmarshal(dbScan.Data, &apiScan)
	if err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScan, nil
}

func (s *ScansTableHandler) CreateScan(scan apitypes.Scan) (apitypes.Scan, error) {
	// Check the user didn't provide an ID
	if scan.Id != nil {
		return apitypes.Scan{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new Scan",
		}
	}

	// Generate a new UUID
	scan.Id = to.Ptr(uuid.New().String())

	// Initialise revision
	scan.Revision = to.Ptr(1)

	// TODO do we want ScanConfig to be required in the api?
	if scan.ScanConfig != nil {
		existingScan, err := s.checkUniqueness(scan)
		if err != nil {
			var conflictErr *common.ConflictError
			if errors.As(err, &conflictErr) {
				return existingScan, err
			}
			return apitypes.Scan{}, fmt.Errorf("failed to check existing scan: %w", err)
		}
	}

	marshaled, err := json.Marshal(scan)
	if err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newScan := Scan{}
	newScan.Data = marshaled

	if err = s.DB.Create(&newScan).Error; err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to create scan in db: %w", err)
	}

	var apiScan apitypes.Scan
	err = json.Unmarshal(newScan.Data, &apiScan)
	if err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScan, nil
}

// nolint:cyclop
func (s *ScansTableHandler) SaveScan(scan apitypes.Scan, params apitypes.PutScansScanIDParams) (apitypes.Scan, error) {
	if scan.Id == nil || *scan.Id == "" {
		return apitypes.Scan{}, &common.BadRequestError{
			Reason: "id is required to save scan",
		}
	}

	var dbObj Scan
	if err := getExistingObjByID(s.DB, scanSchemaName, *scan.Id, &dbObj); err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to get scan from db: %w", err)
	}

	var dbScan apitypes.Scan
	if err := json.Unmarshal(dbObj.Data, &dbScan); err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert DB object to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbScan.Revision); err != nil {
		return apitypes.Scan{}, err
	}

	if err := validateScanConfigID(scan, dbScan); err != nil {
		var badRequestErr *common.BadRequestError
		if errors.As(err, &badRequestErr) {
			return apitypes.Scan{}, err
		}
		return apitypes.Scan{}, fmt.Errorf("scan config id validation failed: %w", err)
	}

	if scan.ScanConfig != nil {
		existingScan, err := s.checkUniqueness(scan)
		if err != nil {
			var conflictErr *common.ConflictError
			if errors.As(err, &conflictErr) {
				return existingScan, err
			}
			return apitypes.Scan{}, fmt.Errorf("failed to check existing scan: %w", err)
		}
	}

	scan.Revision = bumpRevision(dbScan.Revision)

	marshaled, err := json.Marshal(scan)
	if err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err = s.DB.Save(&dbObj).Error; err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to save scan in db: %w", err)
	}

	var apiScan apitypes.Scan
	if err = json.Unmarshal(dbObj.Data, &apiScan); err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScan, nil
}

// nolint:cyclop
func (s *ScansTableHandler) UpdateScan(scan apitypes.Scan, params apitypes.PatchScansScanIDParams) (apitypes.Scan, error) {
	if scan.Id == nil || *scan.Id == "" {
		return apitypes.Scan{}, &common.BadRequestError{
			Reason: "id is required to update scan",
		}
	}

	var dbObj Scan
	if err := getExistingObjByID(s.DB, scanSchemaName, *scan.Id, &dbObj); err != nil {
		return apitypes.Scan{}, err
	}

	var dbScan apitypes.Scan
	if err := json.Unmarshal(dbObj.Data, &dbScan); err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert DB object to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbScan.Revision); err != nil {
		return apitypes.Scan{}, err
	}

	if err := validateScanConfigID(scan, dbScan); err != nil {
		var badRequestErr *common.BadRequestError
		if errors.As(err, &badRequestErr) {
			return apitypes.Scan{}, err
		}
		return apitypes.Scan{}, fmt.Errorf("scan config id validation failed: %w", err)
	}

	scan.Revision = bumpRevision(dbScan.Revision)

	var err error
	dbObj.Data, err = patchObject(dbObj.Data, scan)
	if err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret apitypes.Scan
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if ret.ScanConfig != nil {
		existingScan, err := s.checkUniqueness(ret)
		if err != nil {
			var conflictErr *common.ConflictError
			if errors.As(err, &conflictErr) {
				return existingScan, err
			}
			return apitypes.Scan{}, fmt.Errorf("failed to check existing scan: %w", err)
		}
	}

	if err := s.DB.Save(&dbObj).Error; err != nil {
		return apitypes.Scan{}, fmt.Errorf("failed to save scan in db: %w", err)
	}

	return ret, nil
}

func (s *ScansTableHandler) DeleteScan(scanID apitypes.ScanID) error {
	if err := deleteObjByID(s.DB, scanID, &Scan{}); err != nil {
		return fmt.Errorf("failed to delete scan: %w", err)
	}

	return nil
}

func (s *ScansTableHandler) checkUniqueness(scan apitypes.Scan) (apitypes.Scan, error) {
	var scans []Scan
	// In the case of creating or updating a scan, needs to be checked whether other running scan exists with same scan config id.
	filter := fmt.Sprintf("id ne '%s' and scanConfig/id eq '%s' and endTime eq null", *scan.Id, scan.ScanConfig.Id)
	err := ODataQuery(s.DB, scanSchemaName, &filter, nil, nil, nil, nil, nil, true, &scans)
	if err != nil {
		return apitypes.Scan{}, err
	}

	if len(scans) > 0 {
		var apiScan apitypes.Scan
		if err := json.Unmarshal(scans[0].Data, &apiScan); err != nil {
			return apitypes.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		// If the scan that we want to modify is already finished it can be changed.
		// In the case of creating a new scan the end time will be nil.
		if scan.EndTime == nil {
			return apiScan, &common.ConflictError{
				Reason: fmt.Sprintf("Runnig scan exists with same scanConfigID=%q", scan.ScanConfig.Id),
			}
		}
	}
	return apitypes.Scan{}, nil
}

// In the case of updating a scan, not allowed to change the scan config ID.
func validateScanConfigID(scan apitypes.Scan, dbScan apitypes.Scan) error {
	if scan.ScanConfig == nil {
		return nil
	}
	if scan.ScanConfig.Id == "" {
		return &common.BadRequestError{
			Reason: "scan config id is required when scan config is defined",
		}
	}
	if scan.ScanConfig.Id != dbScan.ScanConfig.Id {
		return &common.BadRequestError{
			Reason: fmt.Sprintf("not allowed to change scan config id from=%s to=%s", dbScan.ScanConfig.Id, scan.ScanConfig.Id),
		}
	}
	return nil
}
