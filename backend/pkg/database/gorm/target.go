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
	targetsTableName = "targets"

	targetTypeVM  = "VMInfo"
	targetTypeDir = "Dir"
	targetTypePod = "Pod"
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

type TargetsTableHandler struct {
	targetsTable *gorm.DB
}

func (db *Handler) TargetsTable() types.TargetsTable {
	return &TargetsTableHandler{
		targetsTable: db.DB.Table(targetsTableName),
	}
}

func (t *TargetsTableHandler) GetTargets(params models.GetTargetsParams) (models.Targets, error) {
	var targets []Target
	tx := t.targetsTable
	if err := tx.Find(&targets).Error; err != nil {
		return models.Targets{}, fmt.Errorf("failed to find targets: %w", err)
	}

	converted, err := ConvertToRestTargets(targets)
	if err != nil {
		return models.Targets{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (t *TargetsTableHandler) GetTarget(targetID models.TargetID) (models.Target, error) {
	var target Target
	if err := t.targetsTable.Where("id = ?", targetID).First(&target).Error; err != nil {
		return models.Target{}, fmt.Errorf("failed to get target by id %q: %w", targetID, err)
	}

	converted, err := ConvertToRestTarget(target)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (t *TargetsTableHandler) CreateTarget(target models.Target) (models.Target, error) {
	dbTarget, err := ConvertToDBTarget(target)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	existingTarget, exist, err := t.checkExist(dbTarget)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to check existing target: %w", err)
	}
	if exist {
		converted, err := ConvertToRestTarget(existingTarget)
		if err != nil {
			return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return converted, &common.ConflictError{
			Reason: fmt.Sprintf("Target exists with the unique constraint combination: %s", getUniqueConstraintsOfTarget(existingTarget)),
		}
	}

	if err := t.targetsTable.Create(&dbTarget).Error; err != nil {
		return models.Target{}, fmt.Errorf("failed to create target in db: %w", err)
	}

	converted, err := ConvertToRestTarget(dbTarget)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (t *TargetsTableHandler) SaveTarget(target models.Target) (models.Target, error) {
	dbTarget, err := ConvertToDBTarget(target)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	if err := t.targetsTable.Save(&dbTarget).Error; err != nil {
		return models.Target{}, fmt.Errorf("failed to save target in db: %w", err)
	}

	converted, err := ConvertToRestTarget(dbTarget)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return converted, nil
}

func (t *TargetsTableHandler) DeleteTarget(targetID models.TargetID) error {
	if err := t.targetsTable.Delete(&Target{}, targetID).Error; err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}
	return nil
}

func (t *TargetsTableHandler) checkExist(target Target) (Target, bool, error) {
	var targetFromDB Target

	tx := t.targetsTable.WithContext(context.Background())

	switch target.Type {
	case targetTypeVM:
		if err := tx.Where("instance_id = ? AND location = ?", *target.InstanceID, *target.Location).First(&targetFromDB).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return Target{}, false, nil
			}
			return Target{}, false, fmt.Errorf("failed to query: %w", err)
		}
	}

	return targetFromDB, true, nil
}

func getUniqueConstraintsOfTarget(target Target) string {
	switch target.Type {
	case targetTypeVM:
		return fmt.Sprintf("instanceID=%s, region=%s", *target.InstanceID, *target.Location)
	case targetTypeDir:
		return "unsupported target type Dir"
	case targetTypePod:
		return "unsupported target type Pod"
	default:
		return fmt.Sprintf("unknown target type: %v", target.Type)
	}
}
