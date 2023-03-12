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
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/openclarity/vmclarity/api/models"
)

func (s *ServerImpl) GetDiscoveryScopes(ctx echo.Context, params models.GetDiscoveryScopesParams) error {
	dbScopes, err := s.dbHandler.ScopesTable().GetScopes(params)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Sprintf("failed to get scopes from db: %v", err))
	}

	return sendResponse(ctx, http.StatusOK, dbScopes)
}

func (s *ServerImpl) PutDiscoveryScopes(ctx echo.Context) error {
	var scopes models.Scopes
	err := ctx.Bind(&scopes)
	if err != nil {
		return sendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to bind request: %v", err).Error())
	}

	updatedScopes, err := s.dbHandler.ScopesTable().SetScopes(scopes)
	if err != nil {
		return sendError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to set scopes in db: %v", err).Error())
	}

	return sendResponse(ctx, http.StatusOK, &updatedScopes)
}
