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

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/backend/pkg/common"
	databaseTypes "github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func (s *ServerImpl) GetScanConfigs(ctx echo.Context, params models.GetScanConfigsParams) error {
	scanConfigs, err := s.dbHandler.ScanConfigsTable().GetScanConfigs(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan configs from db: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, scanConfigs)
}

func (s *ServerImpl) GetScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID, params models.GetScanConfigsScanConfigIDParams) error {
	sc, err := s.dbHandler.ScanConfigsTable().GetScanConfig(scanConfigID, params)
	if err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanConfig with ID %v not found", scanConfigID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan config from db. scanConfigID=%v: %v", scanConfigID, err))
	}
	return sendResponse(ctx, http.StatusOK, sc)
}

func (s *ServerImpl) PostScanConfigs(ctx echo.Context) error {
	var scanConfig models.ScanConfig
	err := ctx.Bind(&scanConfig)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdScanConfig, err := s.dbHandler.ScanConfigsTable().CreateScanConfig(scanConfig)
	if err != nil {
		var conflictErr *common.ConflictError
		if errors.As(err, &conflictErr) {
			existResponse := &models.ScanConfigExists{
				Message:    utils.StringPtr(conflictErr.Reason),
				ScanConfig: &createdScanConfig,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create scan config in db: %v", err))
	}

	return sendResponse(ctx, http.StatusCreated, createdScanConfig)
}

func (s *ServerImpl) DeleteScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	success := models.Success{
		Message: utils.StringPtr(fmt.Sprintf("scan config %v deleted", scanConfigID)),
	}

	if err := s.dbHandler.ScanConfigsTable().DeleteScanConfig(scanConfigID); err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanConfig with ID %v not found", scanConfigID))
		}
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	return sendResponse(ctx, http.StatusNoContent, &success)
}

func (s *ServerImpl) PatchScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	var scanConfig models.ScanConfig
	err := ctx.Bind(&scanConfig)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	scanConfig.Id = &scanConfigID
	updatedScanConfig, err := s.dbHandler.ScanConfigsTable().UpdateScanConfig(scanConfig)
	if err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanConfig with ID %v not found", scanConfigID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan config in db. scanConfigID=%v: %v", scanConfigID, err))
	}

	return sendResponse(ctx, http.StatusOK, updatedScanConfig)
}

func (s *ServerImpl) PutScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	var scanConfig models.ScanConfig
	err := ctx.Bind(&scanConfig)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from the
	// URL field so that the DB layer knows which object is being updated.
	scanConfig.Id = &scanConfigID
	updatedScanConfig, err := s.dbHandler.ScanConfigsTable().SaveScanConfig(scanConfig)
	if err != nil {
		if errors.Is(err, databaseTypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("ScanConfig with ID %v not found", scanConfigID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan config in db. scanConfigID=%v: %v", scanConfigID, err))
	}

	return sendResponse(ctx, http.StatusOK, updatedScanConfig)
}
