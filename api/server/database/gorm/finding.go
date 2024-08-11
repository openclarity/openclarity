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
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/findingkey"
)

const (
	findingSchemaName = "Finding"
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
	err := ODataQuery(s.DB, findingSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &findings)
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
		count, err := ODataCount(s.DB, findingSchemaName, params.Filter)
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
	err := ODataQuery(s.DB, findingSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbFinding)
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

	// Initialise revision
	finding.Revision = to.Ptr(1)

	// Check uniqueness
	existingFinding, err := s.checkUniqueness(finding)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingFinding, err
		}
		return types.Finding{}, fmt.Errorf("failed to check existing finding: %w", err)
	}

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

	var dbObj Finding
	err := getExistingObjByID(s.DB, findingSchemaName, *finding.Id, &dbObj)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to get finding from db: %w", err)
	}

	var dbFinding types.Finding
	err = json.Unmarshal(dbObj.Data, &dbFinding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	finding.Revision = bumpRevision(dbFinding.Revision)

	// Check uniqueness
	existingFinding, err := s.checkUniqueness(finding)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingFinding, err
		}
		return types.Finding{}, fmt.Errorf("failed to check existing finding: %w", err)
	}

	marshaled, err := json.Marshal(finding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err := s.DB.Save(&dbObj).Error; err != nil {
		return types.Finding{}, fmt.Errorf("failed to save finding in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// finding pre-marshal above.
	var sc types.Finding
	err = json.Unmarshal(dbObj.Data, &sc)
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

	var dbObj Finding
	err := getExistingObjByID(s.DB, findingSchemaName, *finding.Id, &dbObj)
	if err != nil {
		return types.Finding{}, err
	}

	var dbFinding types.Finding
	err = json.Unmarshal(dbObj.Data, &dbFinding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	finding.Revision = bumpRevision(dbFinding.Revision)

	dbObj.Data, err = patchObject(dbObj.Data, finding)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret types.Finding
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return types.Finding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	// Check uniqueness
	existingFinding, err := s.checkUniqueness(ret)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingFinding, err
		}
		return types.Finding{}, fmt.Errorf("failed to check existing finding: %w", err)
	}

	if err := s.DB.Save(&dbObj).Error; err != nil {
		return types.Finding{}, fmt.Errorf("failed to save finding in db: %w", err)
	}

	return ret, nil
}

func (s *FindingsTableHandler) DeleteFinding(findingID types.FindingID) error {
	if err := deleteObjByID(s.DB, findingID, &Finding{}); err != nil {
		return fmt.Errorf("failed to delete finding: %w", err)
	}

	return nil
}

//nolint:cyclop
func (s *FindingsTableHandler) checkUniqueness(finding types.Finding) (*types.Finding, error) {
	if finding.Id == nil || *finding.Id == "" {
		return &types.Finding{}, &common.BadRequestError{
			Reason: "finding ID is required",
		}
	}

	if finding.FindingInfo == nil {
		return &types.Finding{}, &common.BadRequestError{
			Reason: "finding FindingInfo is required",
		}
	}

	discriminator, err := finding.FindingInfo.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	// Construct filter based on discriminator type
	// Use info properties that make the finding unique, check package scanner/findingkey.
	var key string
	switch info := discriminator.(type) {
	case types.PackageFindingInfo:
		key = findingkey.GeneratePackageKey(info).Filter()

	case types.VulnerabilityFindingInfo:
		key = findingkey.GenerateVulnerabilityKey(info).Filter()

	case types.MalwareFindingInfo:
		key = findingkey.GenerateMalwareKey(info).Filter()

	case types.SecretFindingInfo:
		key = findingkey.GenerateSecretKey(info).Filter()

	case types.MisconfigurationFindingInfo:
		key = findingkey.GenerateMisconfigurationKey(info).Filter()

	case types.RootkitFindingInfo:
		key = findingkey.GenerateRootkitKey(info).Filter()

	case types.ExploitFindingInfo:
		key = findingkey.GenerateExploitKey(info).Filter()

	case types.InfoFinderFindingInfo:
		key = findingkey.GenerateInfoFinderKey(info).Filter()

	default:
		return nil, fmt.Errorf("finding type is not supported (%T): %w", discriminator, err)
	}

	filter := fmt.Sprintf("id ne '%s' and ", *finding.Id) + key

	// In the case of creating or updating a finding, needs to be checked whether other finding exists with same properties.
	var findings []Finding
	err = ODataQuery(s.DB, findingSchemaName, &filter, nil, nil, nil, nil, nil, true, &findings)
	if err != nil {
		return nil, err
	}
	if len(findings) > 0 {
		var apiFinding types.Finding
		if err := json.Unmarshal(findings[0].Data, &apiFinding); err != nil {
			return nil, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return &apiFinding, &common.ConflictError{
			Reason: fmt.Sprintf("Finding exists with same properties ($filter=%s)", filter),
		}
	}

	return nil, nil //nolint:nilnil
}
