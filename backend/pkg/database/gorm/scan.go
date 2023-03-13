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

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

const (
	scanSchemaName = "Scan"
)

type Scan struct {
	ODataObject
}

type GetScansParams struct {
	// Filter Odata filter
	Filter *string
	// Page Page number of the query
	Page *int
	// PageSize Maximum items to return
	PageSize *int
}

type ScansTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ScansTable() types.ScansTable {
	return &ScansTableHandler{
		DB: db.DB,
	}
}

func (s *ScansTableHandler) GetScans(params models.GetScansParams) (models.Scans, error) {
	var scans []Scan
	err := ODataQuery(s.DB, scanSchemaName, params.Filter, params.Select, params.Expand, params.Top, params.Skip, true, &scans)
	if err != nil {
		return models.Scans{}, err
	}

	items := make([]models.Scan, len(scans))
	for i, sc := range scans {
		var scan models.Scan
		err = json.Unmarshal(sc.Data, &scan)
		if err != nil {
			return models.Scans{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = scan
	}

	output := models.Scans{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, scanSchemaName, params.Filter)
		if err != nil {
			return models.Scans{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *ScansTableHandler) GetScan(scanID models.ScanID, params models.GetScansScanIDParams) (models.Scan, error) {
	var dbScan Scan
	filter := fmt.Sprintf("id eq '%s'", scanID)
	err := ODataQuery(s.DB, scanSchemaName, &filter, params.Select, params.Expand, nil, nil, false, &dbScan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Scan{}, types.ErrNotFound
		}
		return models.Scan{}, err
	}

	var apiScan models.Scan
	err = json.Unmarshal(dbScan.Data, &apiScan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScan, nil
}

func (s *ScansTableHandler) CreateScan(scan models.Scan) (models.Scan, error) {
	// Check the user didn't provide an ID
	if scan.Id != nil {
		return models.Scan{}, fmt.Errorf("can not specify Id field when creating a new Scan")
	}

	// Generate a new UUID
	scan.Id = utils.PointerTo(uuid.New().String())

	// TODO do we want ScanConfig to be required in the api?
	if scan.ScanConfig != nil {
		existingScan, err := s.checkUniqueness(scan.ScanConfig.Id)
		if err != nil {
			var conflictErr *common.ConflictError
			if errors.As(err, &conflictErr) {
				return *existingScan, err
			}
			return models.Scan{}, fmt.Errorf("failed to check existing scan: %w", err)
		}
	}

	marshaled, err := json.Marshal(scan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newScan := Scan{}
	newScan.Data = marshaled

	if err = s.DB.Create(&newScan).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to create scan in db: %w", err)
	}

	var apiScan models.Scan
	err = json.Unmarshal(newScan.Data, &apiScan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScan, nil
}

func (s *ScansTableHandler) SaveScan(scan models.Scan) (models.Scan, error) {
	if scan.Id == nil || *scan.Id == "" {
		return models.Scan{}, fmt.Errorf("ID is required to update scan in DB")
	}

	var dbScan Scan
	if err := getExistingObjByID(s.DB, scanSchemaName, *scan.Id, &dbScan); err != nil {
		return models.Scan{}, fmt.Errorf("failed to get scan from db: %w", err)
	}

	marshaled, err := json.Marshal(scan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbScan.Data = marshaled

	if err = s.DB.Save(&dbScan).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to save scan in db: %w", err)
	}

	var apiScan models.Scan
	if err = json.Unmarshal(dbScan.Data, &apiScan); err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiScan, nil
}

func (s *ScansTableHandler) UpdateScan(scan models.Scan) (models.Scan, error) {
	if scan.Id == nil || *scan.Id == "" {
		return models.Scan{}, fmt.Errorf("ID is required to update scan in DB")
	}

	var dbScan Scan
	if err := getExistingObjByID(s.DB, scanSchemaName, *scan.Id, &dbScan); err != nil {
		return models.Scan{}, err
	}

	marshaled, err := json.Marshal(scan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	// Calculate the diffs between the current doc and the user doc
	patch, err := jsonpatch.CreateMergePatch(dbScan.Data, marshaled)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to calculate patch changes: %w", err)
	}

	// Apply the diff to the doc stored in the DB
	updated, err := jsonpatch.MergePatch(dbScan.Data, patch)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	dbScan.Data = updated

	if err := s.DB.Save(&dbScan).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to save scan in db: %w", err)
	}

	var ret models.Scan
	err = json.Unmarshal(dbScan.Data, &ret)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return ret, nil
}

func (s *ScansTableHandler) DeleteScan(scanID models.ScanID) error {
	if err := deleteObjByID(s.DB, scanID, &Scan{}); err != nil {
		return fmt.Errorf("failed to delete scan: %w", err)
	}

	return nil
}

func (s *ScansTableHandler) checkUniqueness(scanConfigID string) (*models.Scan, error) {
	var scans []Scan
	filter := fmt.Sprintf("scanConfig/id eq '%s' and endTime eq null", scanConfigID)
	err := ODataQuery(s.DB, scanSchemaName, &filter, nil, nil, nil, nil, true, &scans)
	if err != nil {
		return nil, err
	}
	if len(scans) > 0 {
		var apiScan models.Scan
		if err = json.Unmarshal(scans[0].Data, &apiScan); err != nil {
			return nil, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return &apiScan, &common.ConflictError{
			Reason: fmt.Sprintf("Scan with scanConfigID=%q exists and already running", scanConfigID),
		}
	}

	return nil, nil // nolint:nilnil
}
