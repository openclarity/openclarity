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

	"gorm.io/gorm"

	"github.com/labstack/echo/v4"

	"github.com/openclarity/vmclarity/api/server/common"
	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func (s *ServerImpl) GetScans(ctx echo.Context, params types.GetScansParams) error {
	scans, err := s.dbHandler.ScansTable().GetScans(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scans from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, scans)
}

func (s *ServerImpl) PostScans(ctx echo.Context) error {
	var scan types.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	status, ok := scan.GetStatus()
	switch {
	case !ok:
		return sendError(ctx, http.StatusBadRequest, "invalid request: status is missing")
	case status.State != types.ScanStatusStatePending:
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid request: initial state for scan is invalid: %s", status.State))
	default:
	}

	createdScan, err := s.dbHandler.ScansTable().CreateScan(scan)
	if err != nil {
		var conflictErr *common.ConflictError
		var validationErr *common.BadRequestError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.ScanExists{
				Message: to.Ptr(conflictErr.Reason),
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

func (s *ServerImpl) DeleteScansScanID(ctx echo.Context, scanID types.ScanID) error {
	success := types.Success{
		Message: to.Ptr(fmt.Sprintf("scan %v deleted", scanID)),
	}

	if err := s.dbHandler.ScansTable().DeleteScan(scanID); err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to delete scan from db. scanID=%v: %v", scanID, err))
	}

	return sendResponse(ctx, http.StatusOK, &success)
}

func (s *ServerImpl) GetScansScanID(ctx echo.Context, scanID types.ScanID, params types.GetScansScanIDParams) error {
	scan, err := s.dbHandler.ScansTable().GetScan(scanID, params)
	if err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. id=%v: %v", scanID, err))
	}
	return sendResponse(ctx, http.StatusOK, scan)
}

func (s *ServerImpl) PatchScansScanID(ctx echo.Context, scanID types.ScanID, params types.PatchScansScanIDParams) error {
	var scan types.Scan
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

	// check if a scan with id already exists
	existingScan, err := s.dbHandler.ScansTable().GetScan(scanID, types.GetScansScanIDParams{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan was not found: scanID=%v: %v", scanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan: scanID=%v: %v", scanID, err))
	}

	// check for valid state transition if the status was provided
	if status, ok := scan.GetStatus(); ok {
		existingStatus, ok := existingScan.GetStatus()
		if !ok {
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve Status for existing scan: scanID=%v", existingScan.Id))
		}
		err = existingStatus.IsValidTransition(status)
		if err != nil {
			return sendError(ctx, http.StatusBadRequest, err.Error())
		}
	}

	updatedScan, err := s.dbHandler.ScansTable().UpdateScan(scan, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.BadRequestError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		case errors.As(err, &conflictErr):
			existResponse := &types.ScanExists{
				Message: to.Ptr(conflictErr.Reason),
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

// nolint:cyclop
func (s *ServerImpl) PutScansScanID(ctx echo.Context, scanID types.ScanID, params types.PutScansScanIDParams) error {
	var scan types.Scan
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

	// check if a scan with id already exists
	existingScan, err := s.dbHandler.ScansTable().GetScan(scanID, types.GetScansScanIDParams{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan was not found: scanID=%v: %v", scanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan: scanID=%v: %v", scanID, err))
	}

	// check for valid state transition
	status, ok := scan.GetStatus()
	if !ok {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}
	existingStatus, ok := existingScan.GetStatus()
	if !ok {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to retrieve Status for existing scan: scanID=%v", existingScan.Id))
	}
	err = existingStatus.IsValidTransition(status)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, err.Error())
	}

	updatedScan, err := s.dbHandler.ScansTable().SaveScan(scan, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Scan with ID %v not found", scanID))
		case errors.As(err, &conflictErr):
			existResponse := &types.ScanExists{
				Message: to.Ptr(conflictErr.Reason),
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
