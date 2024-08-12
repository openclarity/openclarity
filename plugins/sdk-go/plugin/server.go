// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package plugin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/openclarity/vmclarity/plugins/sdk-go/internal/plugin"

	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"

	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

// APIVersion defines the current version of the Scanner Plugin API.
const APIVersion = "1.0.0"

// server implements Scanner Plugin Server from OpenAPI specs and safely executes
// and handles the operations on given Scanner.
type server struct {
	echo    *echo.Echo
	scanner types.Scanner
}

func newServer(scanner types.Scanner) (*server, error) {
	_, err := plugin.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec: %w", err)
	}

	server := &server{
		echo:    echo.New(),
		scanner: scanner,
	}

	server.echo.Use(echomiddleware.Logger())
	server.echo.Use(echomiddleware.Recover())

	plugin.RegisterHandlers(server.echo, server)

	return server, nil
}

func (s *server) Start(address string) error {
	err := s.echo.Start(address)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd
	defer cancel()

	err := s.echo.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	return nil
}

// Echo returns the underlying Echo server that can be used to e.g. add
// middlewares or register new routes. This should happen before Start.
func (s *server) Echo() *echo.Echo {
	return s.echo
}

//nolint:wrapcheck
func (s *server) GetHealthz(ctx echo.Context) error {
	ready := false
	if status := s.scanner.GetStatus(); status != nil {
		ready = status.State != types.StateNotReady
	}

	if ready {
		return ctx.JSON(http.StatusOK, nil)
	}

	return ctx.JSON(http.StatusServiceUnavailable, nil)
}

//nolint:wrapcheck
func (s *server) GetMetadata(ctx echo.Context) error {
	metadata := s.scanner.Metadata()
	if metadata == nil {
		metadata = &types.Metadata{}
	}

	// Override API version so that we know on host which the actual API server being
	// used for compatibility purposes.
	metadata.ApiVersion = types.Ptr(APIVersion)

	return ctx.JSON(http.StatusOK, metadata)
}

//nolint:wrapcheck
func (s *server) PostConfig(ctx echo.Context) error {
	var config types.Config
	if err := ctx.Bind(&config); err != nil {
		return ctx.JSON(http.StatusBadRequest, &types.ErrorResponse{
			Message: types.Ptr("failed to bind request"),
		})
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return ctx.JSON(http.StatusBadRequest, &types.ErrorResponse{
			Message: types.Ptr("failed to validate request"),
		})
	}

	if s.scanner.GetStatus().State != types.StateReady {
		return ctx.JSON(http.StatusConflict, &types.ErrorResponse{
			Message: types.Ptr("scanner is not in Ready state"),
		})
	}

	s.scanner.Start(config)

	return ctx.JSON(http.StatusCreated, nil)
}

//nolint:wrapcheck
func (s *server) GetStatus(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, s.scanner.GetStatus())
}

//nolint:wrapcheck
func (s *server) PostStop(ctx echo.Context) error {
	var requestBody types.Stop
	if err := ctx.Bind(&requestBody); err != nil {
		return ctx.JSON(http.StatusBadRequest, &types.ErrorResponse{
			Message: types.Ptr("failed to bind request"),
		})
	}

	validate := validator.New()
	if err := validate.Struct(requestBody); err != nil {
		return ctx.JSON(http.StatusBadRequest, &types.ErrorResponse{
			Message: types.Ptr("failed to validate request"),
		})
	}

	s.scanner.Stop(requestBody)

	return ctx.JSON(http.StatusCreated, nil)
}
