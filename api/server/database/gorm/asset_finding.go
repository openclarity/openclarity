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
)

const (
	assetFindingSchemaName = "AssetFinding"
)

type AssetFinding struct {
	ODataObject
}

type AssetFindingsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) AssetFindingsTable() dbtypes.AssetFindingsTable {
	return &AssetFindingsTableHandler{
		DB: db.DB,
	}
}

func (t *AssetFindingsTableHandler) GetAssetFindings(params types.GetAssetFindingsParams) (types.AssetFindings, error) {
	var assetFindings []AssetFinding
	err := ODataQuery(t.DB, assetFindingSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &assetFindings)
	if err != nil {
		return types.AssetFindings{}, err
	}

	items := make([]types.AssetFinding, len(assetFindings))
	for i, pr := range assetFindings {
		var assetFinding types.AssetFinding
		err = json.Unmarshal(pr.Data, &assetFinding)
		if err != nil {
			return types.AssetFindings{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = assetFinding
	}

	output := types.AssetFindings{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(t.DB, assetFindingSchemaName, params.Filter)
		if err != nil {
			return types.AssetFindings{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (t *AssetFindingsTableHandler) GetAssetFinding(assetFindingID types.AssetFindingID, params types.GetAssetFindingsAssetFindingIDParams) (types.AssetFinding, error) {
	var dbAssetFinding AssetFinding
	filter := fmt.Sprintf("id eq '%s'", assetFindingID)
	err := ODataQuery(t.DB, assetFindingSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbAssetFinding)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.AssetFinding{}, dbtypes.ErrNotFound
		}
		return types.AssetFinding{}, err
	}

	var apiAssetFinding types.AssetFinding
	err = json.Unmarshal(dbAssetFinding.Data, &apiAssetFinding)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiAssetFinding, nil
}

func (t *AssetFindingsTableHandler) CreateAssetFinding(assetFinding types.AssetFinding) (types.AssetFinding, error) {
	// Check the user didn't provide an ID
	if assetFinding.Id != nil {
		return types.AssetFinding{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new AssetFinding",
		}
	}

	if assetFinding.Asset == nil || assetFinding.Asset.Id == "" || assetFinding.Finding == nil || assetFinding.Finding.Id == "" {
		return types.AssetFinding{}, &common.BadRequestError{
			Reason: "asset and finding are required fields for creating a new AssetFinding",
		}
	}

	// Generate a new UUID
	assetFinding.Id = to.Ptr(uuid.New().String())

	// Initialise revision
	assetFinding.Revision = to.Ptr(1)

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

	// Check the existing DB entries to ensure that the finding id and asset id fields are unique
	existingAssetFinding, err := t.checkUniqueness(assetFinding)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return existingAssetFinding, err
		}
		return types.AssetFinding{}, fmt.Errorf("failed to check existing asset finding: %w", err)
	}

	marshaled, err := json.Marshal(assetFinding)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newAssetFinding := AssetFinding{}
	newAssetFinding.Data = marshaled

	if err = t.DB.Create(&newAssetFinding).Error; err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to create assetFinding in db: %w", err)
	}

	var sc types.AssetFinding
	err = json.Unmarshal(newAssetFinding.Data, &sc)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return sc, nil
}

// nolint:cyclop
func (t *AssetFindingsTableHandler) SaveAssetFinding(assetFinding types.AssetFinding, params types.PutAssetFindingsAssetFindingIDParams) (types.AssetFinding, error) {
	if assetFinding.Id == nil || *assetFinding.Id == "" {
		return types.AssetFinding{}, &common.BadRequestError{
			Reason: "id is required to save assetFinding",
		}
	}

	var dbObj AssetFinding
	if err := getExistingObjByID(t.DB, assetFindingSchemaName, *assetFinding.Id, &dbObj); err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to get assetFinding from db: %w", err)
	}

	var dbAssetFinding types.AssetFinding
	err := json.Unmarshal(dbObj.Data, &dbAssetFinding)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbAssetFinding.Revision); err != nil {
		return types.AssetFinding{}, err
	}

	assetFinding.Revision = bumpRevision(dbAssetFinding.Revision)

	// Check the existing DB entries to ensure that the finding id and asset id fields are unique
	existingAssetFinding, err := t.checkUniqueness(assetFinding)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return existingAssetFinding, err
		}
		return types.AssetFinding{}, fmt.Errorf("failed to check existing asset finding: %w", err)
	}

	marshaled, err := json.Marshal(assetFinding)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err = t.DB.Save(&dbObj).Error; err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to save assetFinding in db: %w", err)
	}

	var sc types.AssetFinding
	err = json.Unmarshal(dbObj.Data, &sc)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return sc, nil
}

func (t *AssetFindingsTableHandler) UpdateAssetFinding(assetFinding types.AssetFinding, params types.PatchAssetFindingsAssetFindingIDParams) (types.AssetFinding, error) {
	if assetFinding.Id == nil || *assetFinding.Id == "" {
		return types.AssetFinding{}, errors.New("ID is required to update assetFinding in DB")
	}

	var dbObj AssetFinding
	if err := getExistingObjByID(t.DB, assetFindingSchemaName, *assetFinding.Id, &dbObj); err != nil {
		return types.AssetFinding{}, err
	}

	var err error
	var dbAssetFinding types.AssetFinding
	err = json.Unmarshal(dbObj.Data, &dbAssetFinding)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbAssetFinding.Revision); err != nil {
		return types.AssetFinding{}, err
	}

	assetFinding.Revision = bumpRevision(dbAssetFinding.Revision)

	dbObj.Data, err = patchObject(dbObj.Data, assetFinding)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret types.AssetFinding
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	// Check the existing DB entries to ensure that the finding id and asset id fields are unique
	existingAssetFinding, err := t.checkUniqueness(ret)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return existingAssetFinding, err
		}
		return types.AssetFinding{}, fmt.Errorf("failed to check existing asset finding: %w", err)
	}

	if err := t.DB.Save(&dbObj).Error; err != nil {
		return types.AssetFinding{}, fmt.Errorf("failed to save assetFinding in db: %w", err)
	}

	return ret, nil
}

func (t *AssetFindingsTableHandler) DeleteAssetFinding(assetFindingID types.AssetFindingID) error {
	if err := deleteObjByID(t.DB, assetFindingID, &AssetFinding{}); err != nil {
		return fmt.Errorf("failed to delete assetFinding: %w", err)
	}

	return nil
}

func (t *AssetFindingsTableHandler) checkUniqueness(assetFinding types.AssetFinding) (types.AssetFinding, error) {
	// Only check unique if asset and finding are set.
	if assetFinding.Asset == nil || assetFinding.Asset.Id == "" || assetFinding.Finding == nil || assetFinding.Finding.Id == "" {
		return types.AssetFinding{}, &common.BadRequestError{
			Reason: "asset and finding are required fields for creating a new AssetFinding",
		}
	}

	// Check if there is another asset finding with the same asset id and finding id.
	var assetFindings []AssetFinding
	filter := fmt.Sprintf("id ne '%s' and asset/id eq '%s' and finding/id eq '%s'", *assetFinding.Id, assetFinding.Asset.Id, assetFinding.Finding.Id)
	err := ODataQuery(t.DB, assetFindingSchemaName, &filter, nil, nil, nil, nil, nil, true, &assetFindings)
	if err != nil {
		return types.AssetFinding{}, err
	}

	if len(assetFindings) > 0 {
		var as types.AssetFinding
		if err = json.Unmarshal(assetFindings[0].Data, &as); err != nil {
			return types.AssetFinding{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return as, &common.ConflictError{
			Reason: fmt.Sprintf("AssetFinding exists with same asset id=%s and finding id=%s)", assetFinding.Asset.Id, assetFinding.Finding.Id),
		}
	}
	return types.AssetFinding{}, nil
}
