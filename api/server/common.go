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
	"fmt"
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	apiserver "github.com/openclarity/vmclarity/api/server/internal/server"
	"github.com/openclarity/vmclarity/api/types"
)

// nolint:wrapcheck
func sendError(ctx echo.Context, code int, message string) error {
	log.Error(message)
	response := &types.ApiResponse{Message: &message}
	return ctx.JSON(code, response)
}

// nolint:wrapcheck,unparam
func sendResponse(ctx echo.Context, code int, object interface{}) error {
	return ctx.JSON(code, object)
}

func (s *ServerImpl) GetOpenAPISpec(ctx echo.Context) error {
	swagger, err := apiserver.GetSwagger()
	if err != nil {
		return fmt.Errorf("failed to load swagger spec: %w", err)
	}

	// Use the X-Forwarded-* headers to populate the OpenAPI spec with the
	// location where the API is being served. Using this trick the
	// swagger-ui service dynamically loads the OpenAPI spec from the
	// APIServer and knows where to send the TryItNow requests.
	// Wherever the API server is accessed from through a proxy or
	// subdomain this will correct the servers entry to match that clients
	// access path.
	headers := ctx.Request().Header

	log.Debugf("Got headers %#v", headers)

	serverurl := &url.URL{}
	if forwardedHost, ok := headers["X-Forwarded-Host"]; ok {
		proto := "http"
		if forwardedProto, ok := headers["X-Forwarded-Proto"]; ok {
			proto = forwardedProto[0]
		}
		serverurl.Scheme = proto
		serverurl.Host = forwardedHost[0]
	}

	if forwardedPrefix, ok := headers["X-Forwarded-Prefix"]; ok {
		serverurl = serverurl.JoinPath(forwardedPrefix[0])
	}

	if *serverurl != (url.URL{}) {
		swagger.AddServer(&openapi3.Server{
			URL: serverurl.String(),
		})
	}

	return sendResponse(ctx, http.StatusOK, swagger)
}
