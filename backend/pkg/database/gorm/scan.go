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
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
)

const (
	scansTableName = "scans"
)

type Scan struct {
	Base

	ScanStartTime *time.Time `json:"scan_start_time,omitempty" gorm:"column:scan_start_time"`
	ScanEndTime   *time.Time `json:"scan_end_time,omitempty" gorm:"column:scan_end_time"`

	// ScanConfigID The ID of the config that this scan was initiated from (optionanl)
	ScanConfigID *string `json:"scan_config_id,omitempty" gorm:"column:scan_config_id"`
	// ScanFamiliesConfig The configuration of the scanner families within a scan config
	ScanConfigSnapshot []byte `json:"scan_families_config,omitempty" gorm:"column:scan_families_config"`

	// State The lifecycle state of this scan.
	State string `json:"state,omitempty" gorm:"column:state"`

	// StateMessage Human-readable message indicating details about the last state transition.
	StateMessage string `json:"state_message,omitempty" gorm:"column:state_message"`

	// StateReason Machine-readable, UpperCamelCase text indicating the reason for the condition's last transition.
	StateReason string `json:"state_reason,omitempty" gorm:"column:state_reason"`

	// Summary A summary of the progress of a scan for informational purposes.
	Summary []byte `json:"summary,omitempty" gorm:"column:summary"`

	// TargetIDs List of target IDs that are targeted for scanning as part of this scan
	TargetIDs []byte `json:"target_ids,omitempty" gorm:"column:target_ids"`
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
	scansTable *gorm.DB
}

func (db *Handler) ScansTable() types.ScansTable {
	return &ScansTableHandler{
		scansTable: db.DB.Table(scansTableName),
	}
}

func (s *ScansTableHandler) GetScans(params models.GetScansParams) (models.Scans, error) {
	tx := s.scansTable

	var dbScans []Scan
	if err := tx.Find(&dbScans).Error; err != nil {
		return models.Scans{}, fmt.Errorf("failed to find scans: %w", err)
	}

	scans, err := ConvertToRestScans(dbScans)
	if err != nil {
		return models.Scans{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return scans, nil
}

func (s *ScansTableHandler) GetScan(scanID models.ScanID) (models.Scan, error) {
	var dbScan Scan
	if err := s.scansTable.First(&dbScan, "id = ?", scanID).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to get scan by id %q: %w", scanID, err)
	}

	scan, err := ConvertToRestScan(dbScan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return scan, nil
}

func (s *ScansTableHandler) CreateScan(scan models.Scan) (models.Scan, error) {
	// check if there is already a running scan for that scan config id.
	existingSR, exist, err := s.checkExist(scan.ScanConfig.Id)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to check existing scan: %w", err)
	}
	if exist {
		converted, err := ConvertToRestScan(existingSR)
		if err != nil {
			return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return converted, &common.ConflictError{
			Reason: fmt.Sprintf("There is a running scan with scanConfigID=%s", *existingSR.ScanConfigID),
		}
	}

	dbScan, err := ConvertToDBScan(scan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to create scan in db: %w", err)
	}

	if err := s.scansTable.Create(&dbScan).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to create scan in db: %w", err)
	}

	converted, err := ConvertToRestScan(dbScan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScansTableHandler) SaveScan(scan models.Scan) (models.Scan, error) {
	if scan.Id == nil || *scan.Id == "" {
		return models.Scan{}, fmt.Errorf("ID is required to update scan in DB")
	}

	dbScan, err := ConvertToDBScan(scan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to create scan in db: %w", err)
	}

	if err := s.scansTable.Save(&dbScan).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to save scan in db: %w", err)
	}

	converted, err := ConvertToRestScan(dbScan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScansTableHandler) UpdateScan(scan models.Scan) (models.Scan, error) {
	if scan.Id == nil || *scan.Id == "" {
		return models.Scan{}, fmt.Errorf("ID is required to update scan in DB")
	}

	dbScan, err := ConvertToDBScan(scan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to create scan in db: %w", err)
	}

	if err := s.scansTable.Model(dbScan).Updates(&dbScan).Error; err != nil {
		return models.Scan{}, fmt.Errorf("failed to update scan in db: %w", err)
	}

	converted, err := ConvertToRestScan(dbScan)
	if err != nil {
		return models.Scan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScansTableHandler) DeleteScan(scanID models.ScanID) error {
	if err := s.scansTable.Delete(&Scan{}, scanID).Error; err != nil {
		return fmt.Errorf("failed to delete scan: %w", err)
	}
	return nil
}

func (s *ScansTableHandler) checkExist(scanConfigID string) (Scan, bool, error) {
	var scans []Scan

	tx := s.scansTable.WithContext(context.Background())

	if err := tx.Where("scan_config_id = ?", scanConfigID).Find(&scans).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Scan{}, false, nil
		}
		return Scan{}, false, fmt.Errorf("failed to query: %w", err)
	}

	// check if there is a running scan (end time not set)
	for i, scan := range scans {
		if scan.ScanEndTime == nil {
			return scans[i], true, nil
		}
	}

	return Scan{}, false, nil
}
