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
	targetsTableName = "targets"
)

type Target struct {
	Base

	Type     string  `json:"type,omitempty" gorm:"column:type"`
	Location *string `json:"location,omitempty" gorm:"column:location"`

	// VMInfo
	InstanceID       *string `json:"instance_id,omitempty" gorm:"column:instance_id"`
	InstanceProvider *string `json:"instance_provider,omitempty" gorm:"column:instance_provider"`

	// PodInfo
	PodName *string `json:"pod_name,omitempty" gorm:"column:pod_name"`

	// DirInfo
	DirName *string `json:"dir_name,omitempty" gorm:"column:dir_name"`
}

type GetTargetsParams struct {
	// Filter Odata filter
	Filter *string
	// Page Page number of the query
	Page int
	// PageSize Maximum items to return
	PageSize int
}

type TargetsTable interface {
	GetTargetsAndTotal(params GetTargetsParams) ([]*Target, int64, error)
	GetTarget(targetID uuid.UUID) (*Target, error)
	CreateTarget(target *Target) (*Target, error)
	SaveTarget(target *Target) (*Target, error)
	DeleteTarget(targetID uuid.UUID) error
}

type TargetsTableHandler struct {
	targetsTable *gorm.DB
}

func (db *Handler) TargetsTable() TargetsTable {
	return &TargetsTableHandler{
		targetsTable: db.DB.Table(targetsTableName),
	}
}

func (t *TargetsTableHandler) GetTargetsAndTotal(params GetTargetsParams) ([]*Target, int64, error) {
	var count int64
	var targets []*Target

	tx := t.targetsTable

	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total: %w", err)
	}

	if err := tx.Find(&targets).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find targets: %w", err)
	}

	return targets, count, nil
}

func (t *TargetsTableHandler) GetTarget(targetID uuid.UUID) (*Target, error) {
	var target *Target

	if err := t.targetsTable.Where("id = ?", targetID).First(&target).Error; err != nil {
		return nil, fmt.Errorf("failed to get target by id %q: %w", targetID, err)
	}

	return target, nil
}

func (t *TargetsTableHandler) CreateTarget(target *Target) (*Target, error) {
	existingTarget, exist, err := t.checkExist(target)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing target: %w", err)
	}
	if exist {
		return existingTarget, fmt.Errorf("a target alredy exists in db: %w", common.ErrConflict)
	}

	if err := t.targetsTable.Create(target).Error; err != nil {
		return nil, fmt.Errorf("failed to create target in db: %w", err)
	}
	return target, nil
}

func (t *TargetsTableHandler) SaveTarget(target *Target) (*Target, error) {
	if err := t.targetsTable.Save(target).Error; err != nil {
		return nil, fmt.Errorf("failed to save target in db: %w", err)
	}

	return target, nil
}

func (t *TargetsTableHandler) DeleteTarget(targetID uuid.UUID) error {
	if err := t.targetsTable.Delete(&Scan{}, targetID).Error; err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}
	return nil
}

func (t *TargetsTableHandler) checkExist(target *Target) (*Target, bool, error) {
	var targetFromDB Target

	tx := t.targetsTable.WithContext(context.Background())

	switch target.Type {
	case "VMInfo":
		if err := tx.Where("instance_id = ? AND location = ?", *target.InstanceID, *target.Location).First(&targetFromDB).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, false, nil
			}
			return nil, false, fmt.Errorf("failed to query: %w", err)
		}
	}

	return &targetFromDB, true, nil
}
