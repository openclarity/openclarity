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
	"time"

	"github.com/aptible/supercronic/cronexpr"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type ScanConfig struct {
	ODataObject
}

type ScanConfigsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) ScanConfigsTable() types.ScanConfigsTable {
	return &ScanConfigsTableHandler{
		DB: db.DB,
	}
}

func (s *ScanConfigsTableHandler) GetScanConfigs(params models.GetScanConfigsParams) (models.ScanConfigs, error) {
	var scanConfigs []ScanConfig
	err := ODataQuery(s.DB, "ScanConfig", params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &scanConfigs)
	if err != nil {
		return models.ScanConfigs{}, err
	}

	items := []models.ScanConfig{}
	for _, scanConfig := range scanConfigs {
		var sc models.ScanConfig
		err := json.Unmarshal(scanConfig.Data, &sc)
		if err != nil {
			return models.ScanConfigs{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items = append(items, sc)
	}

	output := models.ScanConfigs{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, "ScanConfig", params.Filter)
		if err != nil {
			return models.ScanConfigs{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *ScanConfigsTableHandler) GetScanConfig(scanConfigID models.ScanConfigID, params models.GetScanConfigsScanConfigIDParams) (models.ScanConfig, error) {
	var dbScanConfig ScanConfig
	filter := fmt.Sprintf("id eq '%s'", scanConfigID)
	err := ODataQuery(s.DB, "ScanConfig", &filter, params.Select, params.Expand, nil, nil, nil, false, &dbScanConfig)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ScanConfig{}, types.ErrNotFound
		}
		return models.ScanConfig{}, err
	}

	var sc models.ScanConfig
	err = json.Unmarshal(dbScanConfig.Data, &sc)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return sc, nil
}

// nolint:cyclop
func (s *ScanConfigsTableHandler) CreateScanConfig(scanConfig models.ScanConfig) (models.ScanConfig, error) {
	// Check the user provided the name field
	if scanConfig.Name != nil && *scanConfig.Name == "" {
		return models.ScanConfig{}, fmt.Errorf("name must be provided and can not be empty")
	}

	// Check the user didn't provide an ID
	if scanConfig.Id != nil {
		return models.ScanConfig{}, fmt.Errorf("can not specify Id field when creating a new ScanConfig")
	}

	if err := validateRuntimeScheduleScanConfig(scanConfig.Scheduled); err != nil {
		// Should we return a BadRequest error here?
		return models.ScanConfig{}, fmt.Errorf("failed to validate runtime schedule scan config: %v", err)
	}

	// Generate a new UUID
	scanConfig.Id = utils.PointerTo(uuid.New().String())

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

	// Check the existing DB entries to ensure that the name field is unique
	var scanConfigs []ScanConfig
	filter := fmt.Sprintf("name eq '%s'", *scanConfig.Name)
	err := ODataQuery(s.DB, "ScanConfig", &filter, nil, nil, nil, nil, nil, true, &scanConfigs)
	if err != nil {
		return models.ScanConfig{}, err
	}

	if len(scanConfigs) > 0 {
		var sc models.ScanConfig
		err := json.Unmarshal(scanConfigs[0].Data, &sc)
		if err != nil {
			return models.ScanConfig{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return sc, &common.ConflictError{
			Reason: fmt.Sprintf("Scan config exists with name=%s", *sc.Name),
		}
	}

	marshaled, err := json.Marshal(scanConfig)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newScanConfig := ScanConfig{}
	newScanConfig.Data = marshaled

	if err := s.DB.Create(&newScanConfig).Error; err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to create scan config in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// scanConfig pre-marshal above.
	var sc models.ScanConfig
	err = json.Unmarshal(newScanConfig.Data, &sc)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return sc, nil
}

func validateRuntimeScheduleScanConfig(scheduled *models.RuntimeScheduleScanConfig) error {
	if scheduled == nil {
		return fmt.Errorf("scheduled must be configured")
	}

	if scheduled.CronLine == nil && isEmptyOperationTime(scheduled.OperationTime) {
		return fmt.Errorf("both operationTime and cronLine are not set, " +
			"at least one should be set")
	}

	if scheduled.CronLine != nil {
		// validate cron expression
		expr, err := cronexpr.Parse(*scheduled.CronLine)
		if err != nil {
			return fmt.Errorf("malformed cron expression: %v", err)
		}

		// set operation time if missing
		if isEmptyOperationTime(scheduled.OperationTime) {
			operationTime := expr.Next(time.Now())
			scheduled.OperationTime = &operationTime
		}
	}

	return nil
}

func isEmptyOperationTime(operationTime *time.Time) bool {
	return operationTime == nil || (*operationTime).IsZero()
}

func (s *ScanConfigsTableHandler) SaveScanConfig(scanConfig models.ScanConfig) (models.ScanConfig, error) {
	if scanConfig.Id == nil || *scanConfig.Id == "" {
		return models.ScanConfig{}, fmt.Errorf("ID is required to update scan config in DB")
	}

	// Check the user provided the name field
	if scanConfig.Name != nil && *scanConfig.Name == "" {
		return models.ScanConfig{}, fmt.Errorf("name must be provided and can not be empty")
	}

	if err := validateRuntimeScheduleScanConfig(scanConfig.Scheduled); err != nil {
		// Should we return a BadRequest error here?
		return models.ScanConfig{}, fmt.Errorf("failed to validate runtime schedule scan config: %v", err)
	}

	var dbScanConfig ScanConfig
	if err := getExistingObjByID(s.DB, "ScanConfig", *scanConfig.Id, &dbScanConfig); err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to get scan config from db: %w", err)
	}

	marshaled, err := json.Marshal(scanConfig)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbScanConfig.Data = marshaled

	if err := s.DB.Save(&dbScanConfig).Error; err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to save scan config in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// scanConfig pre-marshal above.
	var sc models.ScanConfig
	err = json.Unmarshal(dbScanConfig.Data, &sc)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return sc, nil
}

func (s *ScanConfigsTableHandler) UpdateScanConfig(scanConfig models.ScanConfig) (models.ScanConfig, error) {
	if scanConfig.Id == nil || *scanConfig.Id == "" {
		return models.ScanConfig{}, fmt.Errorf("ID is required to update scan config in DB")
	}

	// we will want to validate Scheduled upon update only if exists.
	if scanConfig.Scheduled != nil {
		if err := validateRuntimeScheduleScanConfig(scanConfig.Scheduled); err != nil {
			// Should we return a BadRequest error here?
			return models.ScanConfig{}, fmt.Errorf("failed to validate runtime schedule scan config: %v", err)
		}
	}

	var dbScanConfig ScanConfig
	if err := getExistingObjByID(s.DB, "ScanConfig", *scanConfig.Id, &dbScanConfig); err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to get scan config from db: %w", err)
	}

	var err error
	dbScanConfig.Data, err = patchObject(dbScanConfig.Data, scanConfig)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	if err := s.DB.Save(&dbScanConfig).Error; err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to save scan config in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// scanConfig pre-marshal above.
	var sc models.ScanConfig
	err = json.Unmarshal(dbScanConfig.Data, &sc)
	if err != nil {
		return models.ScanConfig{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return sc, nil
}

func (s *ScanConfigsTableHandler) DeleteScanConfig(scanConfigID models.ScanConfigID) error {
	if err := deleteObjByID(s.DB, scanConfigID, &ScanConfig{}); err != nil {
		return fmt.Errorf("failed to delete scan config: %w", err)
	}
	return nil
}
