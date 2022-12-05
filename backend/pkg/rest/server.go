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
	"context"
	"fmt"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/server"
	"github.com/openclarity/vmclarity/backend/pkg/database"

	"github.com/openclarity/vmclarity/backend/pkg/common"
)

const (
	shutdownTimeoutSec = 10
	baseURL            = "/api"
)

var oops = "oops"

type ServerImpl struct {
	dbHandler database.Database
}

type Server struct {
	port       int
	echoServer *echo.Echo
}

func CreateRESTServer(port int) (*Server, error) {
	e, err := createEchoServer()
	if err != nil {
		return nil, fmt.Errorf("failed to create rest server: %v", err)
	}
	return &Server{
		port:       port,
		echoServer: e,
	}, nil
}

func createEchoServer() (*echo.Echo, error) {
	swagger, err := server.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec: %v", err)
	}
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match.
	swagger.Servers = nil

	e := echo.New()
	// Log all requests
	e.Use(echomiddleware.Logger())
	// Create a router group for baseURL
	g := e.Group(baseURL)
	// Use oapi-codegen validation middleware to validate
	// the base URL router group against the OpenAPI schema.
	g.Use(middleware.OapiRequestValidator(swagger))

	return e, nil
}

func (s *Server) RegisterHandlers(dbHandler database.Database) {
	serverImpl := &ServerImpl{
		dbHandler: dbHandler,
	}

	// Register server above as the handler for the interface
	server.RegisterHandlersWithBaseURL(s.echoServer, serverImpl, baseURL)
}

func (s *Server) Start(errChan chan struct{}) {
	log.Infof("Starting REST server")
	go func() {
		if err := s.echoServer.Start(fmt.Sprintf("0.0.0.0:%d", s.port)); err != nil {
			log.Errorf("Failed to start REST server: %v", err)
			errChan <- common.Empty
		}
	}()
	log.Infof("REST server is running")
}

func (s *Server) Stop() {
	log.Infof("Stopping REST server")
	if s.echoServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeoutSec*time.Second)
		defer cancel()
		if err := s.echoServer.Shutdown(ctx); err != nil {
			log.Errorf("Failed to shutdown REST server: %v", err)
		}
	}
}
