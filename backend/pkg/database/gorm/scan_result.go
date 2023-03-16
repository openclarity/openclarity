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

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

const (
	targetScanResultsSchemaName = "TargetScanResult"
)

type ScanResult struct {
	ODataObject
}

type ScanResultsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ScanResultsTable() types.ScanResultsTable {
	return &ScanResultsTableHandler{
		DB: db.DB,
	}
}

func (s *ScanResultsTableHandler) GetScanResults(params models.GetScanResultsParams) (models.TargetScanResults, error) {
	var scanResults []ScanResult
	err := ODataQuery(s.DB, targetScanResultsSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &scanResults)
	if err != nil {
		return models.TargetScanResults{}, err
	}

	items := make([]models.TargetScanResult, len(scanResults))
	for i, scanResult := range scanResults {
		var tsr models.TargetScanResult
		if err = json.Unmarshal(scanResult.Data, &tsr); err != nil {
			return models.TargetScanResults{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = tsr
	}

	output := models.TargetScanResults{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, targetScanResultsSchemaName, params.Filter)
		if err != nil {
			return models.TargetScanResults{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *ScanResultsTableHandler) GetScanResult(scanResultID models.ScanResultID, params models.GetScanResultsScanResultIDParams) (models.TargetScanResult, error) {
	var dbScanResult ScanResult
	filter := fmt.Sprintf("id eq '%s'", scanResultID)
	err := ODataQuery(s.DB, targetScanResultsSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbScanResult)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.TargetScanResult{}, types.ErrNotFound
		}
		return models.TargetScanResult{}, err
	}

	var tsr models.TargetScanResult
	err = json.Unmarshal(dbScanResult.Data, &tsr)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return tsr, nil
}

// nolint:cyclop
func (s *ScanResultsTableHandler) CreateScanResult(scanResult models.TargetScanResult) (models.TargetScanResult, error) {
	// Check the user provided scan id and target id fields
	if scanResult.Scan != nil && scanResult.Scan.Id == "" {
		return models.TargetScanResult{}, fmt.Errorf("scan.id is a required field")
	}
	if scanResult.Target != nil && scanResult.Target.Id == "" {
		return models.TargetScanResult{}, fmt.Errorf("target.id is a required field")
	}

	// Check the user didn't provide an ID
	if scanResult.Id != nil {
		return models.TargetScanResult{}, fmt.Errorf("can not specify ID field when creating a new ScanResult")
	}

	// Generate a new UUID
	scanResult.Id = utils.PointerTo(uuid.New().String())

	// TODO(sambetts) Lock the table here to prevent race conditions
	// checking the uniqueness.
	//
	// We might also be able to do this without locking the table by doing
	// a single query which includes the uniqueness check like:
	//
	// INSERT INTO scan_configs(data) SELECT * FROM (SELECT "<encoded json>") AS tmp WHERE NOT EXISTS (SELECT * FROM scan_configs WHERE JSON_EXTRACT(`Data`, '$.Name') = '<name from input>') LIMIT 1;
	//
	// This should return 0 affected fields if there is a conflicting
	// record in the DB, and should be treated safely by the DB without
	// locking the table.

	// Check the existing DB entries to ensure that the scan id and target id fields are unique
	var scanResults []ScanResult
	filter := fmt.Sprintf("target/id eq '%s' and scan/id eq '%s'", scanResult.Target.Id, scanResult.Scan.Id)
	err := ODataQuery(s.DB, targetScanResultsSchemaName, &filter, nil, nil, nil, nil, nil, true, &scanResults)
	if err != nil {
		return models.TargetScanResult{}, err
	}

	if len(scanResults) > 0 {
		var tsr models.TargetScanResult
		if err = json.Unmarshal(scanResults[0].Data, &tsr); err != nil {
			return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return tsr, &common.ConflictError{
			Reason: fmt.Sprintf("Scan results exists with scan id=%s and target id=%s", scanResult.Target.Id, scanResult.Scan.Id),
		}
	}

	marshaled, err := json.Marshal(scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newScanResult := ScanResult{}
	newScanResult.Data = marshaled

	if err := s.DB.Create(&newScanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to create scan result in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// scanResult pre-marshal above.
	var tsr models.TargetScanResult
	err = json.Unmarshal(newScanResult.Data, &tsr)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return tsr, nil
}

// nolint:cyclop
func (s *ScanResultsTableHandler) SaveScanResult(scanResult models.TargetScanResult) (models.TargetScanResult, error) {
	if scanResult.Id == nil || *scanResult.Id == "" {
		return models.TargetScanResult{}, fmt.Errorf("ID is required to update scan result in DB")
	}

	// Check the user provided scan id and target id fields
	if scanResult.Scan != nil && scanResult.Scan.Id == "" {
		return models.TargetScanResult{}, fmt.Errorf("scan.id is a required field")
	}
	if scanResult.Target != nil && scanResult.Target.Id == "" {
		return models.TargetScanResult{}, fmt.Errorf("target.id is a required field")
	}

	var dbScanResult ScanResult
	if err := getExistingObjByID(s.DB, targetScanResultsSchemaName, *scanResult.Id, &dbScanResult); err != nil {
		return models.TargetScanResult{}, err
	}

	marshaled, err := json.Marshal(scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbScanResult.Data = marshaled

	if err := s.DB.Save(&dbScanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to save scan result in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// scanResult pre-marshal above.
	var tsr models.TargetScanResult
	err = json.Unmarshal(dbScanResult.Data, &tsr)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return tsr, nil
}

func (s *ScanResultsTableHandler) UpdateScanResult(scanResult models.TargetScanResult) (models.TargetScanResult, error) {
	if scanResult.Id == nil || *scanResult.Id == "" {
		return models.TargetScanResult{}, fmt.Errorf("ID is required to update scan result in DB")
	}

	var dbScanResult ScanResult
	if err := getExistingObjByID(s.DB, targetScanResultsSchemaName, *scanResult.Id, &dbScanResult); err != nil {
		return models.TargetScanResult{}, err
	}

	var err error
	dbScanResult.Data, err = patchObject(dbScanResult.Data, scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	if err := s.DB.Save(&dbScanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to save scan result in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// scanResult pre-marshal above.
	var tsr models.TargetScanResult
	err = json.Unmarshal(dbScanResult.Data, &tsr)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return tsr, nil
}
