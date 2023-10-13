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
	"github.com/openclarity/vmclarity/pkg/apiserver/common"
	"github.com/openclarity/vmclarity/pkg/apiserver/database/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	assetSchemaName = "Asset"
)

type Asset struct {
	ODataObject
}

type AssetsTableHandler struct {
	DB *gorm.DB
}

func (db *Handler) AssetsTable() types.AssetsTable {
	return &AssetsTableHandler{
		DB: db.DB,
	}
}

func (t *AssetsTableHandler) GetAssets(params models.GetAssetsParams) (models.Assets, error) {
	var assets []Asset
	err := ODataQuery(t.DB, assetSchemaName, params.Filter, params.Select, params.Expand, params.OrderBy, params.Top, params.Skip, true, &assets)
	if err != nil {
		return models.Assets{}, err
	}

	items := make([]models.Asset, len(assets))
	for i, tr := range assets {
		var asset models.Asset
		err = json.Unmarshal(tr.Data, &asset)
		if err != nil {
			return models.Assets{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		items[i] = asset
	}

	output := models.Assets{Items: &items}

	if params.Count != nil && *params.Count {
		count, err := ODataCount(t.DB, assetSchemaName, params.Filter)
		if err != nil {
			return models.Assets{}, fmt.Errorf("failed to count records: %w", err)
		}
		output.Count = &count
	}

	return output, nil
}

func (t *AssetsTableHandler) GetAsset(assetID models.AssetID, params models.GetAssetsAssetIDParams) (models.Asset, error) {
	var dbAsset Asset
	filter := fmt.Sprintf("id eq '%s'", assetID)
	err := ODataQuery(t.DB, assetSchemaName, &filter, params.Select, params.Expand, nil, nil, nil, false, &dbAsset)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Asset{}, types.ErrNotFound
		}
		return models.Asset{}, err
	}

	var apiAsset models.Asset
	err = json.Unmarshal(dbAsset.Data, &apiAsset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiAsset, nil
}

func (t *AssetsTableHandler) CreateAsset(asset models.Asset) (models.Asset, error) {
	// Check the user didn't provide an ID
	if asset.Id != nil {
		return models.Asset{}, &common.BadRequestError{
			Reason: "can not specify id field when creating a new Asset",
		}
	}

	// Check that assetInfo was provided by the user, it's a required field for an asset.
	if asset.AssetInfo == nil {
		return models.Asset{}, &common.BadRequestError{
			Reason: "assetInfo is a required field",
		}
	}

	// Generate a new UUID
	asset.Id = utils.PointerTo(uuid.New().String())

	// Initialise revision
	asset.Revision = utils.PointerTo(1)

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

	existingAsset, err := t.checkUniqueness(asset)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingAsset, err
		}
		return models.Asset{}, fmt.Errorf("failed to check existing asset: %w", err)
	}

	marshaled, err := json.Marshal(asset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	newAsset := Asset{}
	newAsset.Data = marshaled

	if err = t.DB.Create(&newAsset).Error; err != nil {
		return models.Asset{}, fmt.Errorf("failed to create asset in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// asset pre-marshal above.
	var apiAsset models.Asset
	err = json.Unmarshal(newAsset.Data, &apiAsset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiAsset, nil
}

// nolint:cyclop
func (t *AssetsTableHandler) SaveAsset(asset models.Asset, params models.PutAssetsAssetIDParams) (models.Asset, error) {
	if asset.Id == nil || *asset.Id == "" {
		return models.Asset{}, &common.BadRequestError{
			Reason: "id is required to save asset",
		}
	}

	// Check that assetInfo was provided by the user, it's a required field for an asset.
	if asset.AssetInfo == nil {
		return models.Asset{}, &common.BadRequestError{
			Reason: "assetInfo is a required field",
		}
	}

	var dbObj Asset
	if err := getExistingObjByID(t.DB, assetSchemaName, *asset.Id, &dbObj); err != nil {
		return models.Asset{}, fmt.Errorf("failed to get asset from db: %w", err)
	}

	var dbAsset models.Asset
	err := json.Unmarshal(dbObj.Data, &dbAsset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbAsset.Revision); err != nil {
		return models.Asset{}, err
	}

	asset.Revision = bumpRevision(dbAsset.Revision)

	existingAsset, err := t.checkUniqueness(asset)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingAsset, err
		}
		return models.Asset{}, fmt.Errorf("failed to check existing asset: %w", err)
	}

	marshaled, err := json.Marshal(asset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert API model to DB model: %w", err)
	}

	dbObj.Data = marshaled

	if err = t.DB.Save(&dbObj).Error; err != nil {
		return models.Asset{}, fmt.Errorf("failed to save asset in db: %w", err)
	}

	// TODO(sambetts) Maybe this isn't required now because the DB isn't
	// creating any of the data (like the ID) so we can just return the
	// asset pre-marshal above.
	var apiAsset models.Asset
	if err = json.Unmarshal(dbObj.Data, &apiAsset); err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	return apiAsset, nil
}

// nolint:cyclop
func (t *AssetsTableHandler) UpdateAsset(asset models.Asset, params models.PatchAssetsAssetIDParams) (models.Asset, error) {
	if asset.Id == nil || *asset.Id == "" {
		return models.Asset{}, fmt.Errorf("ID is required to update asset in DB")
	}

	var dbObj Asset
	if err := getExistingObjByID(t.DB, assetSchemaName, *asset.Id, &dbObj); err != nil {
		return models.Asset{}, err
	}

	var err error
	var dbAsset models.Asset
	err = json.Unmarshal(dbObj.Data, &dbAsset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	if err := checkRevisionEtag(params.IfMatch, dbAsset.Revision); err != nil {
		return models.Asset{}, err
	}

	asset.Revision = bumpRevision(dbAsset.Revision)

	dbObj.Data, err = patchObject(dbObj.Data, asset)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to apply patch: %w", err)
	}

	var ret models.Asset
	err = json.Unmarshal(dbObj.Data, &ret)
	if err != nil {
		return models.Asset{}, fmt.Errorf("failed to convert DB model to API model: %w", err)
	}

	existingAsset, err := t.checkUniqueness(ret)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			return *existingAsset, err
		}
		return models.Asset{}, fmt.Errorf("failed to check existing asset: %w", err)
	}

	if err := t.DB.Save(&dbObj).Error; err != nil {
		return models.Asset{}, fmt.Errorf("failed to save asset in db: %w", err)
	}

	return ret, nil
}

func (t *AssetsTableHandler) DeleteAsset(assetID models.AssetID) error {
	if err := deleteObjByID(t.DB, assetID, &Asset{}); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}

// nolint: cyclop
func (t *AssetsTableHandler) checkUniqueness(asset models.Asset) (*models.Asset, error) {
	discriminator, err := asset.AssetInfo.ValueByDiscriminator()
	if err != nil {
		return nil, fmt.Errorf("failed to get value by discriminator: %w", err)
	}

	var filter string
	switch info := discriminator.(type) {
	case models.VMInfo:
		filter = fmt.Sprintf(
			"id ne '%s' and assetInfo/instanceID eq '%s' and assetInfo/location eq '%s'",
			*asset.Id, info.InstanceID, info.Location,
		)
	case models.DirInfo:
		filter = fmt.Sprintf(
			"id ne '%s' and assetInfo/dirName eq '%s' and assetInfo/location eq '%s'",
			*asset.Id, *info.DirName, *info.Location,
		)
	case models.ContainerInfo:
		filter = fmt.Sprintf("id ne '%s' and assetInfo/containerID eq '%s'", *asset.Id, info.ContainerID)

	case models.ContainerImageInfo:
		filter = fmt.Sprintf("id ne '%s' and assetInfo/imageID eq '%s'", *asset.Id, info.ImageID)

	default:
		return nil, fmt.Errorf("asset type is not supported (%T): %w", discriminator, err)
	}

	// In the case of creating or updating an asset, needs to be checked whether other asset exists with same properties.
	var assets []Asset
	err = ODataQuery(t.DB, assetSchemaName, &filter, nil, nil, nil, nil, nil, true, &assets)
	if err != nil {
		return nil, err
	}
	if len(assets) > 0 {
		var apiAsset models.Asset
		if err := json.Unmarshal(assets[0].Data, &apiAsset); err != nil {
			return nil, fmt.Errorf("failed to convert DB model to API model: %w", err)
		}
		return &apiAsset, &common.ConflictError{
			Reason: fmt.Sprintf("Asset exists with same properties ($filter=%s)", filter),
		}
	}
	return nil, nil // nolint:nilnil
}
