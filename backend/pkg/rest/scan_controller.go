// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
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
		if errors.As(err, &conflictErr) {
			existResponse := &models.ScanExists{
				Message: utils.StringPtr(conflictErr.Reason),
				Scan:    &createdScan,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create scan in db: %v", err))
	}

	return sendResponse(ctx, http.StatusCreated, createdScan)
}

func (s *ServerImpl) DeleteScansScanID(ctx echo.Context, scanID models.ScanID) error {
	success := models.Success{
		Message: utils.StringPtr(fmt.Sprintf("scan %v deleted", scanID)),
	}

	if err := s.dbHandler.ScansTable().DeleteScan(scanID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	return sendResponse(ctx, http.StatusNoContent, &success)
}

func (s *ServerImpl) GetScansScanID(ctx echo.Context, scanID models.ScanID) error {
	scan, err := s.dbHandler.ScansTable().GetScan(scanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. id=%v: %v", scanID, err))
	}
	return sendResponse(ctx, http.StatusOK, scan)
}

func (s *ServerImpl) PatchScansScanID(ctx echo.Context, scanID models.ScanID) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// check that a scan with that id exists.
	_, err = s.dbHandler.ScansTable().GetScan(scanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan was not found in db. scanID=%v: %v", scanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. scanID=%v: %v", scanID, err))
	}

	scan.Id = &scanID
	updatedScan, err := s.dbHandler.ScansTable().UpdateScan(scan)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan in db. scanID=%v: %v", scanID, err))
	}

	return sendResponse(ctx, http.StatusOK, updatedScan)
}

func (s *ServerImpl) PutScansScanID(ctx echo.Context, scanID models.ScanID) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// check that a scan with that id exists.
	_, err = s.dbHandler.ScansTable().GetScan(scanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan was not found in db. scanID=%v: %v", scanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. scanID=%v: %v", scanID, err))
	}

	scan.Id = &scanID
	updatedScan, err := s.dbHandler.ScansTable().SaveScan(scan)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan in db. scanID=%v: %v", scanID, err))
	}

	return sendResponse(ctx, http.StatusOK, updatedScan)
}
