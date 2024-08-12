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

package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/apiserver/common"
	databaseTypes "github.com/openclarity/vmclarity/pkg/apiserver/database/types"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func (s *ServerImpl) GetAssets(ctx echo.Context, params models.GetAssetsParams) error {
	dbAssets, err := s.dbHandler.AssetsTable().GetAssets(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get assets from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, dbAssets)
}

// nolint:cyclop
func (s *ServerImpl) PostAssets(ctx echo.Context) error {
	var asset models.Asset
	err := ctx.Bind(&asset)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdAsset, err := s.dbHandler.AssetsTable().CreateAsset(asset)
	if err != nil {
		var conflictErr *common.ConflictError
		var validationErr *common.BadRequestError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &models.AssetExists{
				Message: utils.PointerTo(conflictErr.Reason),
				Asset:   &createdAsset,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create asset in db: %v", err))
		}
	}

	return sendResponse(ctx, http.StatusCreated, createdAsset)
}

func (s *ServerImpl) GetAssetsAssetID(ctx echo.Context, assetID models.AssetID, params models.GetAssetsAssetIDParams) error {
	asset, err := s.dbHandler.AssetsTable().GetAsset(assetID, params)
	if err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Asset with ID %v not found", assetID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset from db. assetID=%v: %v", assetID, err))
	}

	return sendResponse(ctx, http.StatusOK, asset)
}

func (s *ServerImpl) PutAssetsAssetID(ctx echo.Context, assetID models.AssetID, params models.PutAssetsAssetIDParams) error {
	var asset models.Asset
	err := ctx.Bind(&asset)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if asset.Id != nil && *asset.Id != assetID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *asset.Id, assetID))
	}
	asset.Id = &assetID

	updatedAsset, err := s.dbHandler.AssetsTable().SaveAsset(asset, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *databaseTypes.PreconditionFailedError
		switch true {
		case errors.Is(err, databaseTypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Asset with ID %v not found", assetID))
		case errors.As(err, &conflictErr):
			existResponse := &models.AssetExists{
				Message: utils.PointerTo(conflictErr.Reason),
				Asset:   &updatedAsset,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset from db. assetID=%v: %v", assetID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAsset)
}

func (s *ServerImpl) PatchAssetsAssetID(ctx echo.Context, assetID models.AssetID, params models.PatchAssetsAssetIDParams) error {
	var asset models.Asset
	err := ctx.Bind(&asset)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if asset.Id != nil && *asset.Id != assetID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *asset.Id, assetID))
	}
	asset.Id = &assetID

	updatedAsset, err := s.dbHandler.AssetsTable().UpdateAsset(asset, params)
	if err != nil {
		var conflictErr *common.ConflictError
		var preconditionFailedErr *databaseTypes.PreconditionFailedError
		switch true {
		case errors.Is(err, databaseTypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Asset with ID %v not found", assetID))
		case errors.As(err, &conflictErr):
			existResponse := &models.AssetExists{
				Message: utils.PointerTo(conflictErr.Reason),
				Asset:   &updatedAsset,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset from db. assetID=%v: %v", assetID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAsset)
}

func (s *ServerImpl) DeleteAssetsAssetID(ctx echo.Context, assetID models.AssetID) error {
	success := models.Success{
		Message: utils.PointerTo(fmt.Sprintf("asset %v deleted", assetID)),
	}

	if err := s.dbHandler.AssetsTable().DeleteAsset(assetID); err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Asset with ID %v not found", assetID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to delete asset from db. assetID=%v: %v", assetID, err))
	}

	return sendResponse(ctx, http.StatusOK, &success)
}
