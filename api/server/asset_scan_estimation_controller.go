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

package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/server/common"
	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func (s *ServerImpl) GetAssetScanEstimations(ctx echo.Context, params types.GetAssetScanEstimationsParams) error {
	dbAssetScanEstimations, err := s.dbHandler.AssetScanEstimationsTable().GetAssetScanEstimations(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to get asset scan estimations results from db: %w", err).Error())
	}

	return sendResponse(ctx, http.StatusOK, dbAssetScanEstimations)
}

func (s *ServerImpl) PostAssetScanEstimations(ctx echo.Context) error {
	var assetScanEstimation types.AssetScanEstimation
	err := ctx.Bind(&assetScanEstimation)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	status, ok := assetScanEstimation.GetStatus()
	switch {
	case !ok:
		return sendError(ctx, http.StatusBadRequest, "invalid request: status is missing")
	case status.State != types.AssetScanEstimationStatusStatePending:
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid request: initial state for asset scan estimation is invalid: %s", status.State))
	default:
	}

	createdAssetScanEstimation, err := s.dbHandler.AssetScanEstimationsTable().CreateAssetScanEstimation(assetScanEstimation)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			existResponse := &types.AssetScanEstimationExists{
				Message:             to.Ptr(conflictErr.Reason),
				AssetScanEstimation: &createdAssetScanEstimation,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create asset scan estimation in db: %v", err))
	}

	return sendResponse(ctx, http.StatusCreated, createdAssetScanEstimation)
}

func (s *ServerImpl) GetAssetScanEstimationsAssetScanEstimationID(ctx echo.Context, assetScanEstimationID types.AssetScanEstimationID, params types.GetAssetScanEstimationsAssetScanEstimationIDParams) error {
	dbAssetScanEstimation, err := s.dbHandler.AssetScanEstimationsTable().GetAssetScanEstimation(assetScanEstimationID, params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset scan estimation from db. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
	}

	return sendResponse(ctx, http.StatusOK, dbAssetScanEstimation)
}

// nolint:cyclop
func (s *ServerImpl) PatchAssetScanEstimationsAssetScanEstimationID(ctx echo.Context, assetScanEstimationID types.AssetScanEstimationID, params types.PatchAssetScanEstimationsAssetScanEstimationIDParams) error {
	var assetScanEstimation types.AssetScanEstimation
	err := ctx.Bind(&assetScanEstimation)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if assetScanEstimation.Id != nil && *assetScanEstimation.Id != assetScanEstimationID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *assetScanEstimation.Id, assetScanEstimationID))
	}
	assetScanEstimation.Id = &assetScanEstimationID

	// check that an asset scan estimation with that id exists.
	existingAssetScanEstimation, err := s.dbHandler.AssetScanEstimationsTable().GetAssetScanEstimation(assetScanEstimationID, types.GetAssetScanEstimationsAssetScanEstimationIDParams{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("asset scan estimation was not found. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset scan estimation. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
	}

	// check for valid state transition if the status was provided
	if status, ok := assetScanEstimation.GetStatus(); ok {
		existingStatus, ok := existingAssetScanEstimation.GetStatus()
		if !ok {
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve Status for existing scan: scanID=%v", existingAssetScanEstimation.Id))
		}
		err = existingStatus.IsValidTransition(status)
		if err != nil {
			return sendError(ctx, http.StatusBadRequest, err.Error())
		}
	}

	updatedAssetScanEstimation, err := s.dbHandler.AssetScanEstimationsTable().UpdateAssetScanEstimation(assetScanEstimation, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.AssetScanEstimationExists{
				Message:             to.Ptr(conflictErr.Reason),
				AssetScanEstimation: &updatedAssetScanEstimation,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update asset scan estimation in db. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAssetScanEstimation)
}

// nolint:cyclop
func (s *ServerImpl) PutAssetScanEstimationsAssetScanEstimationID(ctx echo.Context, assetScanEstimationID types.AssetScanEstimationID, params types.PutAssetScanEstimationsAssetScanEstimationIDParams) error {
	var assetScanEstimation types.AssetScanEstimation
	err := ctx.Bind(&assetScanEstimation)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if assetScanEstimation.Id != nil && *assetScanEstimation.Id != assetScanEstimationID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *assetScanEstimation.Id, assetScanEstimationID))
	}
	assetScanEstimation.Id = &assetScanEstimationID

	// check that an asset scan estimation with that id exists.
	existingAssetScanEstimation, err := s.dbHandler.AssetScanEstimationsTable().GetAssetScanEstimation(assetScanEstimationID, types.GetAssetScanEstimationsAssetScanEstimationIDParams{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("asset scan estimation was not found. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset scan estimation. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
	}

	// check for valid state transition
	status, ok := assetScanEstimation.GetStatus()
	if !ok {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	existingStatus, ok := existingAssetScanEstimation.GetStatus()
	if !ok {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve Status for existing asset scan estimation: assetScanEstimationID=%v", existingAssetScanEstimation.Id))
	}
	err = existingStatus.IsValidTransition(status)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	updatedAssetScanEstimation, err := s.dbHandler.AssetScanEstimationsTable().SaveAssetScanEstimation(assetScanEstimation, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.AssetScanEstimationExists{
				Message:             to.Ptr(conflictErr.Reason),
				AssetScanEstimation: &updatedAssetScanEstimation,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update asset scan estimation in db. assetScanEstimationID=%v: %v", assetScanEstimationID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAssetScanEstimation)
}

func (s *ServerImpl) DeleteAssetScanEstimationsAssetScanEstimationID(ctx echo.Context, assetScanEstimationID types.AssetScanEstimationID) error {
	success := types.Success{
		Message: to.Ptr(fmt.Sprintf("asset scan estimation %v deleted", assetScanEstimationID)),
	}

	if err := s.dbHandler.AssetScanEstimationsTable().DeleteAssetScanEstimation(assetScanEstimationID); err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("AssetScanEstimation with ID %v not found", assetScanEstimationID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to delete asset scan estimation from db. scanEstimationID=%v: %v", assetScanEstimationID, err))
	}

	return sendResponse(ctx, http.StatusOK, &success)
}
