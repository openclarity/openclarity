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

func (s *ServerImpl) GetAssetScans(ctx echo.Context, params types.GetAssetScansParams) error {
	dbAssetScans, err := s.dbHandler.AssetScansTable().GetAssetScans(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scans results from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, dbAssetScans)
}

func (s *ServerImpl) PostAssetScans(ctx echo.Context) error {
	var assetScan types.AssetScan
	err := ctx.Bind(&assetScan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	status, ok := assetScan.GetStatus()
	switch {
	case !ok:
		return sendError(ctx, http.StatusBadRequest, "invalid request: status is missing")
	case status.State != types.AssetScanStatusStatePending:
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid request: initial state for asset scan is invalid: %s", status.State))
	default:
	}

	cleanupStatus, ok := assetScan.GetResourceCleanupStatus()
	switch {
	case !ok:
		return sendError(ctx, http.StatusBadRequest, "invalid request: resource cleanup status is missing")
	case cleanupStatus.State == types.ResourceCleanupStatusStatePending:
		fallthrough
	case cleanupStatus.State == types.ResourceCleanupStatusStateSkipped:
		break
	default:
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid request: initial state for resource cleanup status is invalid: %s", cleanupStatus.State))
	}

	createdAssetScan, err := s.dbHandler.AssetScansTable().CreateAssetScan(assetScan)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			existResponse := &types.AssetScanExists{
				Message:   to.Ptr(conflictErr.Reason),
				AssetScan: &createdAssetScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create asset scan in db: %v", err))
	}

	return sendResponse(ctx, http.StatusCreated, createdAssetScan)
}

func (s *ServerImpl) GetAssetScansAssetScanID(ctx echo.Context, assetScanID types.AssetScanID, params types.GetAssetScansAssetScanIDParams) error {
	dbAssetScan, err := s.dbHandler.AssetScansTable().GetAssetScan(assetScanID, params)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset scan from db. assetScanID=%v: %v", assetScanID, err))
	}

	return sendResponse(ctx, http.StatusOK, dbAssetScan)
}

// nolint:cyclop
func (s *ServerImpl) PatchAssetScansAssetScanID(ctx echo.Context, assetScanID types.AssetScanID, params types.PatchAssetScansAssetScanIDParams) error {
	// TODO: check that the provided scan and asset IDs are valid
	var assetScan types.AssetScan
	err := ctx.Bind(&assetScan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if assetScan.Id != nil && *assetScan.Id != assetScanID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *assetScan.Id, assetScanID))
	}
	assetScan.Id = &assetScanID

	// check that an asset scan with that id exists.
	existingAssetScan, err := s.dbHandler.AssetScansTable().GetAssetScan(assetScanID, types.GetAssetScansAssetScanIDParams{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("asset scan was not found. assetScanID=%v: %v", assetScanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset scan. assetScanID=%v: %v", assetScanID, err))
	}

	// check for valid state transition
	if status, ok := assetScan.GetStatus(); ok {
		existingStatus, ok := existingAssetScan.GetStatus()
		if !ok {
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve Status for existing asset scan: assetScanID=%v", existingAssetScan.Id))
		}
		err = existingStatus.IsValidTransition(status)
		if err != nil {
			return sendError(ctx, http.StatusBadRequest, err.Error())
		}
	}

	// check for valid resource cleanup state transition
	if resourceCleanupStatus, ok := assetScan.GetResourceCleanupStatus(); ok {
		existingResourceCleanupStatus, ok := existingAssetScan.GetResourceCleanupStatus()
		if !ok {
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve ResourceCleanupStatus for existing asset scan. assetScanID=%v", existingAssetScan.Id))
		}
		err = existingResourceCleanupStatus.IsValidTransition(resourceCleanupStatus)
		if err != nil {
			return sendError(ctx, http.StatusBadRequest, err.Error())
		}
	}

	updatedAssetScan, err := s.dbHandler.AssetScansTable().UpdateAssetScan(assetScan, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.AssetScanExists{
				Message:   to.Ptr(conflictErr.Reason),
				AssetScan: &updatedAssetScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update asset scan in db. assetScanID=%v: %v", assetScanID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAssetScan)
}

// nolint:cyclop
func (s *ServerImpl) PutAssetScansAssetScanID(ctx echo.Context, assetScanID types.AssetScanID, params types.PutAssetScansAssetScanIDParams) error {
	// TODO: check that the provided scan and asset IDs are valid
	var assetScan types.AssetScan
	err := ctx.Bind(&assetScan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if assetScan.Id != nil && *assetScan.Id != assetScanID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *assetScan.Id, assetScanID))
	}
	assetScan.Id = &assetScanID

	// check that an asset scan with that id exists.
	existingAssetScan, err := s.dbHandler.AssetScansTable().GetAssetScan(assetScanID, types.GetAssetScansAssetScanIDParams{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("asset scan was not found. assetScanID=%v: %v", assetScanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get asset scan. assetScanID=%v: %v", assetScanID, err))
	}

	// check for valid state transition if the status was provided
	status, ok := assetScan.GetStatus()
	if !ok {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	existingStatus, ok := existingAssetScan.GetStatus()
	if !ok {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve Status for existing asset scan: assetScanID=%v", existingAssetScan.Id))
	}
	err = existingStatus.IsValidTransition(status)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	// check for valid resource cleanup state transition
	resourceCleanupStatus, ok := assetScan.GetResourceCleanupStatus()
	if !ok {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	existingResourceCleanupStatus, ok := existingAssetScan.GetResourceCleanupStatus()
	if !ok {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve ResourceCleanupStatus for existing asset scan. assetScanID=%v", existingAssetScan.Id))
	}
	err = existingResourceCleanupStatus.IsValidTransition(resourceCleanupStatus)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	updatedAssetScan, err := s.dbHandler.AssetScansTable().SaveAssetScan(assetScan, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.AssetScanExists{
				Message:   to.Ptr(conflictErr.Reason),
				AssetScan: &updatedAssetScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update asset scan in db. assetScanID=%v: %v", assetScanID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAssetScan)
}
