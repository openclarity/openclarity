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
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/pkg/uibackend/api/models"
)

// nolint:wrapcheck,unparam
func sendError(ctx echo.Context, code int, message string) error {
	log.Error(message)
	response := &models.ApiResponse{Message: &message}
	return ctx.JSON(code, response)
}

// nolint:wrapcheck,unparam
func sendResponse(ctx echo.Context, code int, object interface{}) error {
	return ctx.JSON(code, object)
}
