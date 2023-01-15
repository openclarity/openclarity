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

func (s *ServerImpl) GetScanConfigs(ctx echo.Context, params models.GetScanConfigsParams) error {
	dbScanConfigs, total, err := s.dbHandler.ScanConfigsTable().GetScanConfigsAndTotal(resttodb.ConvertGetScanConfigsParams(params))
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan configs from db: %v", err))
	}

	converted, err := dbtorest.ConvertScanConfigs(dbScanConfigs, total)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan configs: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}

func (s *ServerImpl) PostScanConfigs(ctx echo.Context) error {
	var scanConfig models.ScanConfig
	err := ctx.Bind(&scanConfig)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	convertedDB, err := resttodb.ConvertScanConfig(&scanConfig, "")
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config: %v", err))
	}
	createdScanConfig, err := s.dbHandler.ScanConfigsTable().CreateScanConfig(convertedDB)
	if err != nil {
		if errors.Is(err, common.ErrConflict) {
			convertedExist, err := dbtorest.ConvertScanConfig(createdScanConfig)
			if err != nil {
				return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert existing scan config: %v", err))
			}
			return sendResponse(ctx, http.StatusConflict, convertedExist)
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create scan config in db: %v", err))
	}

	converted, err := dbtorest.ConvertScanConfig(createdScanConfig)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config: %v", err))
	}
	return sendResponse(ctx, http.StatusCreated, converted)
}

func (s *ServerImpl) DeleteScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	success := models.Success{
		Message: utils.StringPtr(fmt.Sprintf("scan config %v deleted", scanConfigID)),
	}

	scanConfigUUID, err := uuid.FromString(scanConfigID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scanConfigID %v to uuid: %v", scanConfigID, err))
	}

	if err := s.dbHandler.ScanConfigsTable().DeleteScanConfig(scanConfigUUID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	return sendResponse(ctx, http.StatusNoContent, &success)
}

func (s *ServerImpl) GetScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	scanConfigUUID, err := uuid.FromString(scanConfigID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scanConfigID %v to uuid: %v", scanConfigID, err))
	}

	sc, err := s.dbHandler.ScanConfigsTable().GetScanConfig(scanConfigUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, err.Error())
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan config from db. scanConfigID=%v: %v", scanConfigID, err))
	}

	converted, err := dbtorest.ConvertScanConfig(sc)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config. scanConfigID=%v: %v", scanConfigID, err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}

func (s *ServerImpl) PatchScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	var scanConfig models.ScanConfig
	err := ctx.Bind(&scanConfig)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	convertedDB, err := resttodb.ConvertScanConfig(&scanConfig, scanConfigID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config: %v", err))
	}

	// check that a scan config with that id exists.
	_, err = s.dbHandler.ScanConfigsTable().GetScanConfig(convertedDB.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan config was not found. scanConfigID=%v: %v", scanConfigID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan config from db. scanConfigID=%v: %v", scanConfigID, err))
	}

	updatedScanConfig, err := s.dbHandler.ScanConfigsTable().UpdateScanConfig(convertedDB)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan config in db. scanConfigID=%v: %v", scanConfigID, err))
	}

	converted, err := dbtorest.ConvertScanConfig(updatedScanConfig)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config. scanConfigID=%v: %v", scanConfigID, err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}

func (s *ServerImpl) PutScanConfigsScanConfigID(ctx echo.Context, scanConfigID models.ScanConfigID) error {
	var scanConfig models.ScanConfig
	err := ctx.Bind(&scanConfig)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	convertedDB, err := resttodb.ConvertScanConfig(&scanConfig, scanConfigID)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config: %v", err))
	}

	// check that a scan config with that id exists.
	_, err = s.dbHandler.ScanConfigsTable().GetScanConfig(convertedDB.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("scan config was not found. scanConfigID=%v: %v", scanConfigID, err))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scan config from db. scanConfigID=%v: %v", scanConfigID, err))
	}

	updatedScanConfig, err := s.dbHandler.ScanConfigsTable().SaveScanConfig(convertedDB)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update scan config in db. scanConfigID=%v: %v", scanConfigID, err))
	}

	converted, err := dbtorest.ConvertScanConfig(updatedScanConfig)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to convert scan config. scanConfigID=%v: %v", scanConfigID, err))
	}
	return sendResponse(ctx, http.StatusOK, converted)
}
