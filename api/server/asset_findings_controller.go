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

	"github.com/openclarity/vmclarity/api/server/common"
	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/to"
)

func (s *ServerImpl) GetAssetFindings(ctx echo.Context, params types.GetAssetFindingsParams) error {
	assetFindings, err := s.dbHandler.AssetFindingsTable().GetAssetFindings(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get assetFindings from db: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, assetFindings)
}

func (s *ServerImpl) GetAssetFindingsAssetFindingID(ctx echo.Context, assetFindingID types.AssetFindingID, params types.GetAssetFindingsAssetFindingIDParams) error {
	sc, err := s.dbHandler.AssetFindingsTable().GetAssetFinding(assetFindingID, params)
	if err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("AssetFinding with ID %v not found", assetFindingID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get assetFinding from db. assetFindingID=%v: %v", assetFindingID, err))
	}
	return sendResponse(ctx, http.StatusOK, sc)
}

func (s *ServerImpl) PostAssetFindings(ctx echo.Context) error {
	var assetFinding types.AssetFinding
	err := ctx.Bind(&assetFinding)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdAssetFinding, err := s.dbHandler.AssetFindingsTable().CreateAssetFinding(assetFinding)
	if err != nil {
		var conflictErr *common.ConflictError
		var validationErr *common.BadRequestError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.AssetFindingExists{
				Message: to.Ptr(conflictErr.Reason),
				Finding: &createdAssetFinding,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create assetFinding in db: %v", err))
		}
	}

	return sendResponse(ctx, http.StatusCreated, createdAssetFinding)
}

func (s *ServerImpl) DeleteAssetFindingsAssetFindingID(ctx echo.Context, assetFindingID types.AssetFindingID) error {
	success := types.Success{
		Message: to.Ptr(fmt.Sprintf("assetFinding %v deleted", assetFindingID)),
	}

	if err := s.dbHandler.AssetFindingsTable().DeleteAssetFinding(assetFindingID); err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("AssetFinding with ID %v not found", assetFindingID))
		}
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	return sendResponse(ctx, http.StatusOK, &success)
}

func (s *ServerImpl) PatchAssetFindingsAssetFindingID(ctx echo.Context, assetFindingID types.AssetFindingID, params types.PatchAssetFindingsAssetFindingIDParams) error {
	var assetFinding types.AssetFinding
	err := ctx.Bind(&assetFinding)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if assetFinding.Id != nil && *assetFinding.Id != assetFindingID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *assetFinding.Id, assetFindingID))
	}
	assetFinding.Id = &assetFindingID
	updatedAssetFinding, err := s.dbHandler.AssetFindingsTable().UpdateAssetFinding(assetFinding, params)
	if err != nil {
		var validationErr *common.BadRequestError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("AssetFinding with ID %v not found", assetFindingID))
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update assetFinding in db. assetFindingID=%v: %v", assetFindingID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAssetFinding)
}

func (s *ServerImpl) PutAssetFindingsAssetFindingID(ctx echo.Context, assetFindingID types.AssetFindingID, params types.PutAssetFindingsAssetFindingIDParams) error {
	var assetFinding types.AssetFinding
	err := ctx.Bind(&assetFinding)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from the
	// URL field so that the DB layer knows which object is being updated.
	if assetFinding.Id != nil && *assetFinding.Id != assetFindingID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *assetFinding.Id, assetFindingID))
	}
	assetFinding.Id = &assetFindingID

	updatedAssetFinding, err := s.dbHandler.AssetFindingsTable().SaveAssetFinding(assetFinding, params)
	if err != nil {
		var validationErr *common.BadRequestError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("AssetFinding with ID %v not found", assetFindingID))
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update assetFinding in db. assetFindingID=%v: %v", assetFindingID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedAssetFinding)
}
