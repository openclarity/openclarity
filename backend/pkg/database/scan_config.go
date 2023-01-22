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

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/backend/pkg/common"
)

const (
	scanConfigsTableName = "scan_configs"
)

type ScanConfig struct {
	Base

	Name *string `json:"name,omitempty" gorm:"column:name"`

	// ScanFamiliesConfig The configuration of the scanner families within a scan config
	ScanFamiliesConfig []byte `json:"scan_families_config,omitempty" gorm:"column:scan_families_config"`
	Scheduled          []byte `json:"scheduled,omitempty" gorm:"column:scheduled"`
	Scope              []byte `json:"scope,omitempty" gorm:"column:scope"`
}

type GetScanConfigsParams struct {
	// Filter Odata filter
	Filter *string
	// Page Page number of the query
	Page *int
	// PageSize Maximum items to return
	PageSize *int
}

type ScanConfigsTable interface {
	GetScanConfigsAndTotal(params GetScanConfigsParams) ([]*ScanConfig, int64, error)
	GetScanConfig(scanConfigID uuid.UUID) (*ScanConfig, error)
	UpdateScanConfig(scanConfig *ScanConfig) (*ScanConfig, error)
	SaveScanConfig(scanConfig *ScanConfig) (*ScanConfig, error)
	DeleteScanConfig(scanConfigID uuid.UUID) error
	CreateScanConfig(scanConfig *ScanConfig) (*ScanConfig, error)
}

type ScanConfigsTableHandler struct {
	scanConfigsTable *gorm.DB
}

func (db *Handler) ScanConfigsTable() ScanConfigsTable {
	return &ScanConfigsTableHandler{
		scanConfigsTable: db.DB.Table(scanConfigsTableName),
	}
}

func (s *ScanConfigsTableHandler) GetScanConfigsAndTotal(params GetScanConfigsParams) ([]*ScanConfig, int64, error) {
	var count int64
	var scanConfigs []*ScanConfig

	tx := s.scanConfigsTable

	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %w", err)
	}

	if err := tx.Find(&scanConfigs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find scan configs: %w", err)
	}

	return scanConfigs, count, nil
}

func (s *ScanConfigsTableHandler) CreateScanConfig(scanConfig *ScanConfig) (*ScanConfig, error) {
	// check if there is already a scan config with that name.
	existingSR, exist, err := s.checkExist(*scanConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing scan config: %w", err)
	}
	if exist {
		return existingSR, fmt.Errorf("a scan config alredy exists with the name %v: %w", *scanConfig.Name, common.ErrConflict)
	}

	if err := s.scanConfigsTable.Create(scanConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create scan config in db: %w", err)
	}
	return scanConfig, nil
}

func (s *ScanConfigsTableHandler) SaveScanConfig(scanConfig *ScanConfig) (*ScanConfig, error) {
	if err := s.scanConfigsTable.Save(scanConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to save scan config in db: %w", err)
	}

	return scanConfig, nil
}

func (s *ScanConfigsTableHandler) UpdateScanConfig(scanConfig *ScanConfig) (*ScanConfig, error) {
	if err := s.scanConfigsTable.Model(scanConfig).Updates(scanConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to update scan config in db: %w", err)
	}

	return scanConfig, nil
}

func (s *ScanConfigsTableHandler) GetScanConfig(scanConfigID uuid.UUID) (*ScanConfig, error) {
	var scanConfig *ScanConfig

	if err := s.scanConfigsTable.Where("id = ?", scanConfigID).First(&scanConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to get scan config by id %q: %w", scanConfigID, err)
	}

	return scanConfig, nil
}

func (s *ScanConfigsTableHandler) DeleteScanConfig(scanConfigID uuid.UUID) error {
	if err := s.scanConfigsTable.Delete(&Scan{}, scanConfigID).Error; err != nil {
		return fmt.Errorf("failed to delete scan config: %w", err)
	}
	return nil
}

func (s *ScanConfigsTableHandler) checkExist(name string) (*ScanConfig, bool, error) {
	var scanConfig *ScanConfig

	tx := s.scanConfigsTable.WithContext(context.Background())

	if err := tx.Where("name = ?", name).First(&scanConfig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to query: %w", err)
	}

	return scanConfig, true, nil
}
