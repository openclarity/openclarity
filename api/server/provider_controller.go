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

func (s *ServerImpl) GetProviders(ctx echo.Context, params types.GetProvidersParams) error {
	dbProviders, err := s.dbHandler.ProvidersTable().GetProviders(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get providers from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, dbProviders)
}

// nolint:cyclop
func (s *ServerImpl) PostProviders(ctx echo.Context) error {
	var provider types.Provider
	err := ctx.Bind(&provider)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	createdProvider, err := s.dbHandler.ProvidersTable().CreateProvider(provider)
	if err != nil {
		var conflictErr *common.ConflictError
		var validationErr *common.BadRequestError
		switch true {
		case errors.As(err, &conflictErr):
			existResponse := &types.ProviderExists{
				Message:  to.Ptr(conflictErr.Reason),
				Provider: &createdProvider,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to create provider in db: %v", err))
		}
	}

	return sendResponse(ctx, http.StatusCreated, createdProvider)
}

func (s *ServerImpl) GetProvidersProviderID(ctx echo.Context, providerID types.ProviderID, params types.GetProvidersProviderIDParams) error {
	provider, err := s.dbHandler.ProvidersTable().GetProvider(providerID, params)
	if err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Provider with ID %v not found", providerID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get provider from db. providerID=%v: %v", providerID, err))
	}

	return sendResponse(ctx, http.StatusOK, provider)
}

func (s *ServerImpl) PutProvidersProviderID(ctx echo.Context, providerID types.ProviderID, params types.PutProvidersProviderIDParams) error {
	var provider types.Provider
	err := ctx.Bind(&provider)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PUT request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if provider.Id != nil && *provider.Id != providerID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *provider.Id, providerID))
	}
	provider.Id = &providerID

	updatedProvider, err := s.dbHandler.ProvidersTable().SaveProvider(provider, params)
	if err != nil {
		var validationErr *common.BadRequestError
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Provider with ID %v not found", providerID))
		case errors.As(err, &conflictErr):
			existResponse := &types.ProviderExists{
				Message:  to.Ptr(conflictErr.Reason),
				Provider: &updatedProvider,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &validationErr):
			return sendError(ctx, http.StatusBadRequest, err.Error())
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get provider from db. providerID=%v: %v", providerID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedProvider)
}

func (s *ServerImpl) PatchProvidersProviderID(ctx echo.Context, providerID types.ProviderID, params types.PatchProvidersProviderIDParams) error {
	var provider types.Provider
	err := ctx.Bind(&provider)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("failed to bind request: %v", err))
	}

	// PATCH request might not contain the ID in the body, so set it from
	// the URL field so that the DB layer knows which object is being updated.
	if provider.Id != nil && *provider.Id != providerID {
		return sendError(ctx, http.StatusBadRequest, fmt.Sprintf("id in body %s does not match object %s to be updated", *provider.Id, providerID))
	}
	provider.Id = &providerID

	updatedProvider, err := s.dbHandler.ProvidersTable().UpdateProvider(provider, params)
	if err != nil {
		var conflictErr *common.ConflictError
		var preconditionFailedErr *dbtypes.PreconditionFailedError
		switch true {
		case errors.Is(err, dbtypes.ErrNotFound):
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Provider with ID %v not found", providerID))
		case errors.As(err, &conflictErr):
			existResponse := &types.ProviderExists{
				Message:  to.Ptr(conflictErr.Reason),
				Provider: &updatedProvider,
			}
			return sendResponse(ctx, http.StatusConflict, existResponse)
		case errors.As(err, &preconditionFailedErr):
			return sendError(ctx, http.StatusPreconditionFailed, err.Error())
		default:
			return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get provider from db. providerID=%v: %v", providerID, err))
		}
	}

	return sendResponse(ctx, http.StatusOK, updatedProvider)
}

func (s *ServerImpl) DeleteProvidersProviderID(ctx echo.Context, providerID types.ProviderID) error {
	success := types.Success{
		Message: to.Ptr(fmt.Sprintf("provider %v deleted", providerID)),
	}

	if err := s.dbHandler.ProvidersTable().DeleteProvider(providerID); err != nil {
		if errors.Is(err, dbtypes.ErrNotFound) {
			return sendError(ctx, http.StatusNotFound, fmt.Sprintf("Provider with ID %v not found", providerID))
		}
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to delete provider from db. providerID=%v: %v", providerID, err))
	}

	return sendResponse(ctx, http.StatusOK, &success)
}
