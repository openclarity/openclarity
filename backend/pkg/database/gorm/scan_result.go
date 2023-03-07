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

package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/backend/pkg/database/types"
)

const (
	scanResultsTableName = "scan_results"
)

type ScanResult struct {
	Base

	ScanID   string `json:"scan_id,omitempty" gorm:"column:scan_id"`
	TargetID string `json:"target_id,omitempty" gorm:"column:target_id"`

	Exploits          []byte `json:"exploits,omitempty" gorm:"column:exploits"`
	Malware           []byte `json:"malware,omitempty" gorm:"column:malware"`
	Misconfigurations []byte `json:"misconfigurations,omitempty" gorm:"column:misconfigurations"`
	Rootkits          []byte `json:"rootkits,omitempty" gorm:"column:rootkits"`
	Sboms             []byte `json:"sboms,omitempty" gorm:"column:sboms"`
	Secrets           []byte `json:"secrets,omitempty" gorm:"column:secrets"`
	Status            []byte `json:"status,omitempty" gorm:"column:status"`
	Summary           []byte `json:"summary,omitempty" gorm:"column:summary"`
	Vulnerabilities   []byte `json:"vulnerabilities,omitempty" gorm:"column:vulnerabilities"`
}

type ScanResultsTableHandler struct {
	scanResultsTable *gorm.DB
}

func (db *Handler) ScanResultsTable() types.ScanResultsTable {
	return &ScanResultsTableHandler{
		scanResultsTable: db.DB.Table(scanResultsTableName),
	}
}

func (s *ScanResultsTableHandler) GetScanResults(params models.GetScanResultsParams) (models.TargetScanResults, error) {
	var scanResults []ScanResult
	tx := s.scanResultsTable
	if err := tx.Find(&scanResults).Error; err != nil {
		return models.TargetScanResults{}, fmt.Errorf("failed to find scan results: %w", err)
	}

	converted, err := ConvertToRestScanResults(scanResults)
	if err != nil {
		return models.TargetScanResults{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScanResultsTableHandler) GetScanResult(scanResultID models.ScanResultID, params models.GetScanResultsScanResultIDParams) (models.TargetScanResult, error) {
	var scanResult ScanResult
	if err := s.scanResultsTable.Where("id = ?", scanResultID).First(&scanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to get scan result by id %q: %w", scanResultID, err)
	}

	converted, err := ConvertToRestScanResult(scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScanResultsTableHandler) CreateScanResult(scanResult models.TargetScanResult) (models.TargetScanResult, error) {
	// check if there is already a scanResult for that scan id and target id.
	existingSR, exist, err := s.checkExist(scanResult.ScanId, scanResult.TargetId)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to check existing scan result: %w", err)
	}
	if exist {
		converted, err := ConvertToRestScanResult(existingSR)
		if err != nil {
			return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return converted, &common.ConflictError{
			Reason: fmt.Sprintf("Target scan result exists with scanID=%s, targetID=%s", existingSR.ScanID, existingSR.TargetID),
		}
	}

	dbScanResult, err := ConvertToDBScanResult(scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	if err := s.scanResultsTable.Create(&dbScanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to create scan result in db: %w", err)
	}

	converted, err := ConvertToRestScanResult(dbScanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScanResultsTableHandler) SaveScanResult(scanResult models.TargetScanResult) (models.TargetScanResult, error) {
	dbScanResult, err := ConvertToDBScanResult(scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	if err := s.scanResultsTable.Save(&dbScanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to save scan result in db: %w", err)
	}

	converted, err := ConvertToRestScanResult(dbScanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScanResultsTableHandler) UpdateScanResult(scanResult models.TargetScanResult) (models.TargetScanResult, error) {
	dbScanResult, err := ConvertToDBScanResult(scanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	if err := s.scanResultsTable.Model(dbScanResult).Updates(&dbScanResult).Error; err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to update scan result in db: %w", err)
	}

	converted, err := ConvertToRestScanResult(dbScanResult)
	if err != nil {
		return models.TargetScanResult{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (s *ScanResultsTableHandler) checkExist(scanID string, targetID string) (ScanResult, bool, error) {
	var scanResult ScanResult

	tx := s.scanResultsTable.WithContext(context.Background())

	if err := tx.Where("scan_id = ? AND target_id = ?", scanID, targetID).First(&scanResult).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ScanResult{}, false, nil
		}
		return ScanResult{}, false, fmt.Errorf("failed to query: %w", err)
	}

	return scanResult, true, nil
}
