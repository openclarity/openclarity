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

	"github.com/openclarity/vmclarity/api/server/common"
	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"
	"github.com/openclarity/vmclarity/api/types"
)

type Finding struct {
	ODataObject
}

type FindingsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) FindingsTable() dbtypes.FindingsTable {
	return &FindingsTableHandler{
		DB: db.DB,
	}
}

func (s *FindingsTableHandler) GetFindings(params types.GetFindingsParams) (types.Findings, error) {
	var findings []Finding
	err := ODataQuery(s.DB, "Finding", params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &findings)
	if err != nil {
		return types.Findings{}, err
	}

	items := []types.Finding{}
	for _, finding := range findings {
		var sc types.Finding
		err := json.Unmarshal(finding.Data, &sc)
		if err != nil {
			return types.Findings{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items = append(items, sc)
	}

	output := types.Findings{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, "Finding", params.Filter)
		if err != nil {
			return types.Findings{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *FindingsTableHandler) GetFinding(findingID types.FindingID, params types.GetFindingsFindingIDParams) (types.Finding, error) {
	var dbFinding Finding
	filter := fmt.Sprintf("id eq '%s'", findingID)
	err := ODataQuery(s.DB, "Finding", &filter, params.Select, params.Expand, nil, nil, nil, false, &dbFinding)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.Finding{}, dbtypes.ErrNotFound
		}
		return types.Finding{}, err
	}

	var sc types.Finding
	err = json.Unmarshal(dbFinding.Data, &sc)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return sc, nil
}

func (s *FindingsTableHandler) CreateFinding(finding types.Finding) (types.Finding, error) {
	// Check the user didn't provide an ID
	if finding.Id != nil {
		return types.Finding{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new Finding",
		}
	}

	// Generate a new UUID
	newID := uuid.New().String()
	finding.Id = &newID

	marshaled, err := json.Marshal(finding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newFinding := Finding{}
	newFinding.Data = marshaled

	if err := s.DB.Create(&newFinding).Error; err != nil {
		return types.Finding{}, fmt.Errorf("failed to create finding in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// finding pre-marshal above.
	var sc types.Finding
	err = json.Unmarshal(newFinding.Data, &sc)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return sc, nil
}

func (s *FindingsTableHandler) SaveFinding(finding types.Finding) (types.Finding, error) {
	if finding.Id == nil || *finding.Id == "" {
		return types.Finding{}, &common.BadRequestError{
			Reason: "id is required to save finding",
		}
	}

	var dbFinding Finding
	err := getExistingObjByID(s.DB, "Finding", *finding.Id, &dbFinding)
	if err != nil {
		return types.Finding{}, err
	}

	marshaled, err := json.Marshal(finding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbFinding.Data = marshaled

	if err := s.DB.Save(&dbFinding).Error; err != nil {
		return types.Finding{}, fmt.Errorf("failed to save finding in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// finding pre-marshal above.
	var sc types.Finding
	err = json.Unmarshal(dbFinding.Data, &sc)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return sc, nil
}

func (s *FindingsTableHandler) UpdateFinding(finding types.Finding) (types.Finding, error) {
	if finding.Id == nil || *finding.Id == "" {
		return types.Finding{}, &common.BadRequestError{
			Reason: "id is required to update finding",
		}
	}

	var dbFinding Finding
	err := getExistingObjByID(s.DB, "Finding", *finding.Id, &dbFinding)
	if err != nil {
		return types.Finding{}, err
	}

	dbFinding.Data, err = patchObject(dbFinding.Data, finding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	if err := s.DB.Save(&dbFinding).Error; err != nil {
		return types.Finding{}, fmt.Errorf("failed to save finding in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// finding pre-marshal above.
	var sc types.Finding
	err = json.Unmarshal(dbFinding.Data, &sc)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}
	return sc, nil
}

func (s *FindingsTableHandler) DeleteFinding(findingID types.FindingID) error {
	if err := deleteObjByID(s.DB, findingID, &Finding{}); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}
