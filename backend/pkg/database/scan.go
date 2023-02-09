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

package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/backend/pkg/common"
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
	ScanFamiliesConfig []byte `json:"scan_families_config,omitempty" gorm:"column:scan_families_config"`

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

type ScansTable interface {
	GetScansAndTotal(params GetScansParams) ([]*Scan, int64, error)
	GetScan(scanID uuid.UUID) (*Scan, error)
	UpdateScan(scan *Scan) (*Scan, error)
	SaveScan(scan *Scan) (*Scan, error)
	DeleteScan(scanID uuid.UUID) error
	CreateScan(scan *Scan) (*Scan, error)
}

type ScansTableHandler struct {
	scansTable *gorm.DB
}

func (db *Handler) ScansTable() ScansTable {
	return &ScansTableHandler{
		scansTable: db.DB.Table(scansTableName),
	}
}

func (s *ScansTableHandler) GetScansAndTotal(params GetScansParams) ([]*Scan, int64, error) {
	var count int64
	var scans []*Scan

	tx := s.scansTable

	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %w", err)
	}

	if err := tx.Find(&scans).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find scans: %w", err)
	}

	return scans, count, nil
}

func (s *ScansTableHandler) CreateScan(scan *Scan) (*Scan, error) {
	// check if there is already a running scan for that scan config id.
	existingSR, exist, err := s.checkExist(*scan.ScanConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing scan: %w", err)
	}
	if exist {
		return existingSR, &common.ConflictError{
			Reason: fmt.Sprintf("There is a running scan with scanConfigID=%s", *existingSR.ScanConfigID),
		}
	}

	if err := s.scansTable.Create(scan).Error; err != nil {
		return nil, fmt.Errorf("failed to create scan in db: %w", err)
	}
	return scan, nil
}

func (s *ScansTableHandler) SaveScan(scan *Scan) (*Scan, error) {
	if err := s.scansTable.Save(scan).Error; err != nil {
		return nil, fmt.Errorf("failed to save scan in db: %w", err)
	}

	return scan, nil
}

func (s *ScansTableHandler) UpdateScan(scan *Scan) (*Scan, error) {
	if err := s.scansTable.Model(scan).Updates(scan).Error; err != nil {
		return nil, fmt.Errorf("failed to update scan in db: %w", err)
	}
	return scan, nil
}

func (s *ScansTableHandler) GetScan(scanID uuid.UUID) (*Scan, error) {
	var scan *Scan

	if err := s.scansTable.Where("id = ?", scanID).First(&scan).Error; err != nil {
		return nil, fmt.Errorf("failed to get scan by id %q: %w", scanID, err)
	}

	return scan, nil
}

func (s *ScansTableHandler) DeleteScan(scanID uuid.UUID) error {
	if err := s.scansTable.Delete(&Scan{}, scanID).Error; err != nil {
		return fmt.Errorf("failed to delete scan: %w", err)
	}
	return nil
}

func (s *ScansTableHandler) checkExist(scanConfigID string) (*Scan, bool, error) {
	var scans []Scan

	tx := s.scansTable.WithContext(context.Background())

	if err := tx.Where("scan_config_id = ?", scanConfigID).Find(&scans).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to query: %w", err)
	}

	// check if there is a running scan (end time not set)
	for i, scan := range scans {
		if scan.ScanEndTime == nil {
			return &scans[i], true, nil
		}
	}

	return nil, false, nil
}
