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

func (s *ServerImpl) GetScans(ctx echo.Context, params models.GetScansParams) error {
	scans, err := s.dbHandler.ScansTable().GetScans(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scans from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, scans)
}

func (s *ServerImpl) PostScans(ctx echo.Context) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdScan, err := s.dbHandler.ScansTable().CreateScan(scan)
	if err != nil {
		var conflictErr *common.ConflictError
		var validationErr *common.BadRequestError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &models.ScanExists{
				Message: utils.PointerTo(conflictErr.Reason),
				Scan:    &createdScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create scan in db: %v", err))
		}
	}

	return sendResponse(ctx, http.StatusCreated, createdScan)
}

func (s *ServerImpl) DeleteScansScanID(ctx echo.Context, scanID models.ScanID) error {
	success := models.Success{
		Message: utils.PointerTo(fmt.Sprintf("scan %v deleted", scanID)),
	}

	if err := s.dbHandler.ScansTable().DeleteScan(scanID); err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to delete scan from db. scanID=%v: %v", scanID, err))
	}

	return sendResponse(ctx, http.StatusOK, &success)
}

func (s *ServerImpl) GetScansScanID(ctx echo.Context, scanID models.ScanID, params models.GetScansScanIDParams) error {
	scan, err := s.dbHandler.ScansTable().GetScan(scanID, params)
	if err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. id=%v: %v", scanID, err))
	}
	return sendResponse(ctx, http.StatusOK, scan)
}

func (s *ServerImpl) PatchScansScanID(ctx echo.Context, scanID models.ScanID, params models.PatchScansScanIDParams) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if scan.Id != nil && *scan.Id != scanID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *scan.Id, scanID))
	}
	scan.Id = &scanID

	updatedScan, err := s.dbHandler.ScansTable().UpdateScan(scan, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.BadRequestError
		var preconditionFailedErr *databaseTypes.PreconditionFailedError
		switch true {
		case errors.Is(err, databaseTypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		case errors.As(err, &conflictErr):
			existResponse := &models.ScanExists{
				Message: utils.PointerTo(conflictErr.Reason),
				Scan:    &updatedScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan in db. scanID=%v: %v", scanID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedScan)
}

func (s *ServerImpl) PutScansScanID(ctx echo.Context, scanID models.ScanID, params models.PutScansScanIDParams) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from the
	// URL field so that the DB layer knows which object is being updated.
	if scan.Id != nil && *scan.Id != scanID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *scan.Id, scanID))
	}
	scan.Id = &scanID

	updatedScan, err := s.dbHandler.ScansTable().SaveScan(scan, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *databaseTypes.PreconditionFailedError
		switch true {
		case errors.Is(err, databaseTypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		case errors.As(err, &conflictErr):
			existResponse := &models.ScanExists{
				Message: utils.PointerTo(conflictErr.Reason),
				Scan:    &updatedScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to save scan in db. scanID=%v: %v", scanID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedScan)
}
