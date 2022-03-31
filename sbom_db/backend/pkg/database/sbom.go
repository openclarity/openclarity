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
	"fmt"

	"gorm.io/gorm"
	_clause "gorm.io/gorm/clause"
)

const (
	sbomTableName = "sbom"

	// NOTE: when changing one of the column names change also the gorm label in SBOM.
	columnSBOMID           = "id"
	columnSBOMResourceHash = "resource_hash"
)

func (SBOM) TableName() string {
	return sbomTableName
}

type SBOM struct {
	ID string `gorm:"primarykey" faker:"-"` // consists of the resource hash

	ResourceHash string `json:"resource_hash,omitempty" gorm:"column:resource_hash" faker:"oneof: hash1, hash2, hash3"`
	SBOM         string `json:"sbom,omitempty" gorm:"column:sbom" faker:"oneof: sbom1, sbom2, sbom3"`
}

type SBOMTable interface {
	CreateOrUpdateSBOM(sbom *SBOM) error
	GetSBOM(resourceHash string) (*SBOM, error)
}

type SBOMTableHandler struct {
	sbomTable *gorm.DB
}

func (db *Handler) SBOMTable() SBOMTable {
	return &SBOMTableHandler{
		sbomTable: db.DB.Table(sbomTableName),
	}
}

func (s *SBOMTableHandler) CreateOrUpdateSBOM(sbom *SBOM) error {
	// On conflict, update record with the new one
	clause := _clause.OnConflict{
		Columns:   []_clause.Column{{Name: columnSBOMID}},
		UpdateAll: true,
	}

	if err := s.sbomTable.Clauses(clause).Create(sbom).Error; err != nil {
		return fmt.Errorf("failed to create sbom: %v", err)
	}
	return nil
}

func (s *SBOMTableHandler) GetSBOM(resourceHash string) (*SBOM, error) {
	var sbom SBOM

	if err := s.sbomTable.Where(columnSBOMResourceHash+" = ?", resourceHash).Scan(&sbom).Error; err != nil {
		return nil, fmt.Errorf("failed to get sbom from db by resource hash %v: %v", resourceHash, err)
	}

	return &sbom, nil
}
