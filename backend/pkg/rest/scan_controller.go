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
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	"github.com/openclarity/vmclarity/backend/pkg/rest/convert/dbtorest"
	"github.com/openclarity/vmclarity/backend/pkg/rest/convert/resttodb"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func (s *ServerImpl) GetScans(ctx echo.Context, params models.GetScansParams) error {
	dbScans, total, err := s.dbHandler.ScansTable().GetScansAndTotal(resttodb.ConvertGetScansParams(params))
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scans from db: %v", err))
	}

	converted, err := dbtorest.ConvertScans(dbScans, total)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scans: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}

func (s *ServerImpl) PostScans(ctx echo.Context) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	convertedDB, err := resttodb.ConvertScan(&scan, "")
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}
	createdScan, err := s.dbHandler.ScansTable().CreateScan(convertedDB)
	if err != nil {
		if errors.Is(err, common.ErrConflict) {
			convertedExist, err := dbtorest.ConvertScan(createdScan)
			if err != nil {
				return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert existing scan: %v", err))
			}
			return sendResponse(ctx, http.StatusConflict, convertedExist)
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create scan in db: %v", err))
	}

	converted, err := dbtorest.ConvertScan(createdScan)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}
	return sendResponse(ctx, http.StatusCreated, converted)
}

func (s *ServerImpl) DeleteScansScanID(ctx echo.Context, scanID models.ScanID) error {
	success := models.Success{
		Message: utils.StringPtr(fmt.Sprintf("scan %v deleted", scanID)),
	}

	scanUUID, err := uuid.FromString(scanID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scanID %v to uuid: %v", scanID, err))
	}

	if err := s.dbHandler.ScansTable().DeleteScan(scanUUID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	return sendResponse(ctx, http.StatusNoContent, &success)
}

func (s *ServerImpl) GetScansScanID(ctx echo.Context, scanID models.ScanID) error {
	scanUUID, err := uuid.FromString(scanID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scanID %v to uuid: %v", scanID, err))
	}

	scan, err := s.dbHandler.ScansTable().GetScan(scanUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. id=%v: %v", scanID, err))
	}

	converted, err := dbtorest.ConvertScan(scan)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}

func (s *ServerImpl) PatchScansScanID(ctx echo.Context, scanID models.ScanID) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	convertedDB, err := resttodb.ConvertScan(&scan, scanID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}

	// check that a scan with that id exists.
	_, err = s.dbHandler.ScansTable().GetScan(convertedDB.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan was not found in db. scanID=%v: %v", scanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. scanID=%v: %v", scanID, err))
	}

	updatedScan, err := s.dbHandler.ScansTable().UpdateScan(convertedDB)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan in db. scanID=%v: %v", scanID, err))
	}

	converted, err := dbtorest.ConvertScan(updatedScan)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}

func (s *ServerImpl) PutScansScanID(ctx echo.Context, scanID models.ScanID) error {
	var scan models.Scan
	err := ctx.Bind(&scan)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	convertedDB, err := resttodb.ConvertScan(&scan, scanID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}

	// check that a scan with that id exists.
	_, err = s.dbHandler.ScansTable().GetScan(convertedDB.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan was not found in db. scanID=%v: %v", scanID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan from db. scanID=%v: %v", scanID, err))
	}

	updatedScan, err := s.dbHandler.ScansTable().SaveScan(convertedDB)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan in db. scanID=%v: %v", scanID, err))
	}

	converted, err := dbtorest.ConvertScan(updatedScan)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}
