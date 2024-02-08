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

func (s *ServerImpl) GetFindings(ctx echo.Context, params types.GetFindingsParams) error {
	findings, err := s.dbHandler.FindingsTable().GetFindings(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get findings from db: %v", err))
	}
	return sendResponse(ctx, http.StatusOK, findings)
}

func (s *ServerImpl) GetFindingsFindingID(ctx echo.Context, findingID types.FindingID, params types.GetFindingsFindingIDParams) error {
	sc, err := s.dbHandler.FindingsTable().GetFinding(findingID, params)
	if err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Finding with ID %v not found", findingID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get finding from db. findingID=%v: %v", findingID, err))
	}
	return sendResponse(ctx, http.StatusOK, sc)
}

func (s *ServerImpl) PostFindings(ctx echo.Context) error {
	var finding types.Finding
	err := ctx.Bind(&finding)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdFinding, err := s.dbHandler.FindingsTable().CreateFinding(finding)
	if err != nil {
		var conflictErr *common.ConflictError
		var validationErr *common.BadRequestError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.FindingExists{
				Message: to.Ptr(conflictErr.Reason),
				Finding: &createdFinding,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create finding in db: %v", err))
		}
	}

	return sendResponse(ctx, http.StatusCreated, createdFinding)
}

func (s *ServerImpl) DeleteFindingsFindingID(ctx echo.Context, findingID types.FindingID) error {
	success := types.Success{
		Message: to.Ptr(fmt.Sprintf("finding %v deleted", findingID)),
	}

	if err := s.dbHandler.FindingsTable().DeleteFinding(findingID); err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Finding with ID %v not found", findingID))
		}
		return sendError(ctx, http.StatusInternalServerError, err.Error())
	}

	return sendResponse(ctx, http.StatusOK, &success)
}

func (s *ServerImpl) PatchFindingsFindingID(ctx echo.Context, findingID types.FindingID) error {
	var finding types.Finding
	err := ctx.Bind(&finding)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if finding.Id != nil && *finding.Id != findingID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *finding.Id, findingID))
	}
	finding.Id = &findingID
	updatedFinding, err := s.dbHandler.FindingsTable().UpdateFinding(finding)
	if err != nil {
		var validationErr *common.BadRequestError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Finding with ID %v not found", findingID))
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update finding in db. findingID=%v: %v", findingID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedFinding)
}

func (s *ServerImpl) PutFindingsFindingID(ctx echo.Context, findingID types.FindingID) error {
	var finding types.Finding
	err := ctx.Bind(&finding)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from the
	// URL field so that the DB layer knows which object is being updated.
	if finding.Id != nil && *finding.Id != findingID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *finding.Id, findingID))
	}
	finding.Id = &findingID

	updatedFinding, err := s.dbHandler.FindingsTable().SaveFinding(finding)
	if err != nil {
		var validationErr *common.BadRequestError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Finding with ID %v not found", findingID))
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to update finding in db. findingID=%v: %v", findingID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedFinding)
}
