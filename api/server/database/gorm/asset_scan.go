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
	assetScansSchemaName = "AssetScan"
)

type AssetScan struct {
	ODataObject
}

type AssetScansTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) AssetScansTable() dbtypes.AssetScansTable {
	return &AssetScansTableHandler{
		DB: db.DB,
	}
}

func (s *AssetScansTableHandler) GetAssetScans(params types.GetAssetScansParams) (types.AssetScans, error) {
	var assetScans []AssetScan
	err := ODataQuery(s.DB, assetScansSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &assetScans)
	if err != nil {
		return types.AssetScans{}, err
	}

	items := make([]types.AssetScan, len(assetScans))
	for i, assetScan := range assetScans {
		var as types.AssetScan
		if err = json.Unmarshal(assetScan.Data, &as); err != nil {
			return types.AssetScans{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = as
	}

	output := types.AssetScans{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(s.DB, assetScansSchemaName, params.Filter)
		if err != nil {
			return types.AssetScans{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (s *AssetScansTableHandler) GetAssetScan(assetScanID types.AssetScanID, params types.GetAssetScansAssetScanIDParams) (types.AssetScan, error) {
	var dbAssetScan AssetScan
	filter := fmt.Sprintf("id eq '%s'", assetScanID)
	err := ODataQuery(s.DB, assetScansSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbAssetScan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.AssetScan{}, dbtypes.ErrNotFound
		}
		return types.AssetScan{}, err
	}

	var as types.AssetScan
	err = json.Unmarshal(dbAssetScan.Data, &as)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return as, nil
}

// nolint:cyclop
func (s *AssetScansTableHandler) CreateAssetScan(assetScan types.AssetScan) (types.AssetScan, error) {
	// Check the user provided asset id field
	if assetScan.Asset != nil && assetScan.Asset.Id == "" {
		return types.AssetScan{}, &common.BadRequestError{
			Reason: "asset.id is a required field",
		}
	}

	// Check the user didn't provide an ID
	if assetScan.Id != nil {
		return types.AssetScan{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new AssetScan",
		}
	}

	// Generate a new UUID
	assetScan.Id = to.Ptr(uuid.New().String())

	// Initialise revision
	assetScan.Revision = to.Ptr(1)

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

	// Check the existing DB entries to ensure that the scan id and asset id fields are unique
	existingAssetScan, err := s.checkUniqueness(assetScan)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return existingAssetScan, err
		}
		return types.AssetScan{}, fmt.Errorf("failed to check existing scan: %w", err)
	}

	marshaled, err := json.Marshal(assetScan)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newAssetScan := AssetScan{}
	newAssetScan.Data = marshaled

	if err := s.DB.Create(&newAssetScan).Error; err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to create asset scan in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// assetScan pre-marshal above.
	var as types.AssetScan
	err = json.Unmarshal(newAssetScan.Data, &as)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return as, nil
}

// nolint:cyclop,gocognit
func (s *AssetScansTableHandler) SaveAssetScan(assetScan types.AssetScan, params types.PutAssetScansAssetScanIDParams) (types.AssetScan, error) {
	if assetScan.Id == nil || *assetScan.Id == "" {
		return types.AssetScan{}, &common.BadRequestError{
			Reason: "id is required to save asset scan",
		}
	}

	// Check the user provided asset id field
	if assetScan.Asset != nil && assetScan.Asset.Id == "" {
		return types.AssetScan{}, &common.BadRequestError{
			Reason: "asset.id is a required field",
		}
	}

	// Check the existing DB entries to ensure that the scan id and asset id fields are unique
	existingAssetScan, err := s.checkUniqueness(assetScan)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return existingAssetScan, err
		}
		return types.AssetScan{}, fmt.Errorf("failed to check existing scan: %w", err)
	}

	var dbObj AssetScan
	if err := getExistingObjByID(s.DB, assetScansSchemaName, *assetScan.Id, &dbObj); err != nil {
		return types.AssetScan{}, err
	}

	var dbAssetScan types.AssetScan
	err = json.Unmarshal(dbObj.Data, &dbAssetScan)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbAssetScan.Revision); err != nil {
		return types.AssetScan{}, err
	}

	assetScan.Revision = bumpRevision(dbAssetScan.Revision)

	marshaled, err := json.Marshal(assetScan)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err := s.DB.Save(&dbObj).Error; err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to save asset scan in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// assetScan pre-marshal above.
	var as types.AssetScan
	err = json.Unmarshal(dbObj.Data, &as)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return as, nil
}

// nolint:cyclop
func (s *AssetScansTableHandler) UpdateAssetScan(assetScan types.AssetScan, params types.PatchAssetScansAssetScanIDParams) (types.AssetScan, error) {
	if assetScan.Id == nil || *assetScan.Id == "" {
		return types.AssetScan{}, &common.BadRequestError{
			Reason: "id is required to update asset scan",
		}
	}

	var dbObj AssetScan
	if err := getExistingObjByID(s.DB, assetScansSchemaName, *assetScan.Id, &dbObj); err != nil {
		return types.AssetScan{}, err
	}

	var err error
	var dbAssetScan types.AssetScan
	err = json.Unmarshal(dbObj.Data, &dbAssetScan)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbAssetScan.Revision); err != nil {
		return types.AssetScan{}, err
	}

	assetScan.Revision = bumpRevision(dbAssetScan.Revision)

	dbObj.Data, err = patchObject(dbObj.Data, assetScan)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var as types.AssetScan
	err = json.Unmarshal(dbObj.Data, &as)
	if err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	// Check the existing DB entries to ensure that the scan id and asset id fields are unique
	existingAssetScan, err := s.checkUniqueness(as)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return existingAssetScan, err
		}
		return types.AssetScan{}, fmt.Errorf("failed to check existing scan: %w", err)
	}

	if err := s.DB.Save(&dbObj).Error; err != nil {
		return types.AssetScan{}, fmt.Errorf("failed to save asset scan in db: %w", err)
	}

	return as, nil
}

func (s *AssetScansTableHandler) checkUniqueness(assetScan types.AssetScan) (types.AssetScan, error) {
	// We only check unique if scan is set, so return early if it's not set.
	if assetScan.Scan == nil || assetScan.Scan.Id == "" {
		return types.AssetScan{}, nil
	}

	// If Scan is set we need to check if there is another asset scan with
	// the same scan id and asset id.
	var assetScans []AssetScan
	filter := fmt.Sprintf("id ne '%s' and asset/id eq '%s' and scan/id eq '%s'", *assetScan.Id, assetScan.Asset.Id, assetScan.Scan.Id)
	err := ODataQuery(s.DB, assetScansSchemaName, &filter, nil, nil, nil, nil, nil, true, &assetScans)
	if err != nil {
		return types.AssetScan{}, err
	}

	if len(assetScans) > 0 {
		var as types.AssetScan
		if err = json.Unmarshal(assetScans[0].Data, &as); err != nil {
			return types.AssetScan{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return as, &common.ConflictError{
			Reason: fmt.Sprintf("AssetScan exists with same asset id=%s and scan id=%s)", assetScan.Asset.Id, assetScan.Scan.Id),
		}
	}
	return types.AssetScan{}, nil
}
