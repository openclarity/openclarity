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
	targetSchemaName = "Target"
)

type Target struct {
	ODataObject
}

type TargetsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) TargetsTable() types.TargetsTable {
	return &TargetsTableHandler{
		DB: db.DB,
	}
}

func (t *TargetsTableHandler) GetTargets(params models.GetTargetsParams) (models.Targets, error) {
	var targets []Target
	err := ODataQuery(t.DB, targetSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &targets)
	if err != nil {
		return models.Targets{}, err
	}

	items := make([]models.Target, len(targets))
	for i, tr := range targets {
		var target models.Target
		err = json.Unmarshal(tr.Data, &target)
		if err != nil {
			return models.Targets{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = target
	}

	output := models.Targets{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(t.DB, targetSchemaName, params.Filter)
		if err != nil {
			return models.Targets{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (t *TargetsTableHandler) GetTarget(targetID models.TargetID, params models.GetTargetsTargetIDParams) (models.Target, error) {
	var dbTarget Target
	filter := fmt.Sprintf("id eq '%s'", targetID)
	err := ODataQuery(t.DB, targetSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbTarget)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Target{}, types.ErrNotFound
		}
		return models.Target{}, err
	}

	var apiTarget models.Target
	err = json.Unmarshal(dbTarget.Data, &apiTarget)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiTarget, nil
}

func (t *TargetsTableHandler) CreateTarget(target models.Target) (models.Target, error) {
	// Check the user didn't provide an ID
	if target.Id != nil {
		return models.Target{}, fmt.Errorf("can not specify Id field when creating a new Target")
	}

	// Generate a new UUID
	target.Id = utils.PointerTo(uuid.New().String())

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

	existingTarget, err := t.checkUniqueness(target)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingTarget, err
		}
		return models.Target{}, fmt.Errorf("failed to check existing target: %w", err)
	}

	marshaled, err := json.Marshal(target)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newTarget := Target{}
	newTarget.Data = marshaled

	if err = t.DB.Create(&newTarget).Error; err != nil {
		return models.Target{}, fmt.Errorf("failed to create target in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// target pre-marshal above.
	var apiTarget models.Target
	err = json.Unmarshal(newTarget.Data, &apiTarget)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiTarget, nil
}

func (t *TargetsTableHandler) SaveTarget(target models.Target) (models.Target, error) {
	if target.Id == nil || *target.Id == "" {
		return models.Target{}, fmt.Errorf("ID is required to update target in DB")
	}

	var dbTarget Target
	if err := getExistingObjByID(t.DB, targetSchemaName, *target.Id, &dbTarget); err != nil {
		return models.Target{}, fmt.Errorf("failed to get target from db: %w", err)
	}

	marshaled, err := json.Marshal(target)
	if err != nil {
		return models.Target{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbTarget.Data = marshaled

	if err = t.DB.Save(&dbTarget).Error; err != nil {
		return models.Target{}, fmt.Errorf("failed to save target in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// target pre-marshal above.
	var apiTarget models.Target
	if err = json.Unmarshal(dbTarget.Data, &apiTarget); err != nil {
		return models.Target{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiTarget, nil
}

func (t *TargetsTableHandler) DeleteTarget(targetID models.TargetID) error {
	if err := deleteObjByID(t.DB, targetID, &Target{}); err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}

	return nil
}

func (t *TargetsTableHandler) checkUniqueness(target models.Target) (*models.Target, error) {
	discriminator, err := target.TargetInfo.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	switch info := discriminator.(type) {
	case models.VMInfo:
		var targets []Target
		filter := fmt.Sprintf("targetInfo/instanceID eq '%s' and targetInfo/location eq '%s'", info.InstanceID, info.Location)
		err = ODataQuery(t.DB, targetSchemaName, &filter, nil, nil, nil, nil, nil, true, &targets)
		if err != nil {
			return nil, err
		}

		if len(targets) > 0 {
			var apiTarget models.Target
			if err = json.Unmarshal(targets[0].Data, &apiTarget); err != nil {
				return nil, fmt.Errorf("failed to convert DB model to API model: %w", err)
			}
			return &apiTarget, &common.ConflictError{
				Reason: fmt.Sprintf("Target VM exists with instanceID=%q and location=%q", info.InstanceID, info.Location),
			}
		}
	default:
		return nil, fmt.Errorf("target type is not supported (%T): %w", discriminator, err)
	}

	return nil, nil // nolint:nilnil
}
