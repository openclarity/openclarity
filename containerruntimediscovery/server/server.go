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
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/openclarity/vmclarity/containerruntimediscovery/types"
	"github.com/openclarity/vmclarity/core/log"
)

type ContainerRuntimeDiscoveryServer struct {
	server *echo.Echo

	discoverer types.Discoverer
}

func NewContainerRuntimeDiscoveryServer(discoverer types.Discoverer) *ContainerRuntimeDiscoveryServer {
	crds := &ContainerRuntimeDiscoveryServer{
		discoverer: discoverer,
	}

	e := echo.New()
	e.Use(middleware.Recover())

	e.GET("/images", crds.ListImages)
	e.GET("/images/:id", crds.GetImage)
	e.GET("/exportimage/:id", crds.ExportImage)
	e.GET("/containers", crds.ListContainers)
	e.GET("/containers/:id", crds.GetContainer)
	e.GET("/exportcontainer/:id", crds.ExportContainer)

	crds.server = e

	return crds
}

func (crds *ContainerRuntimeDiscoveryServer) Serve(ctx context.Context, listenAddr string) {
	logger := log.GetLoggerFromContextOrDefault(ctx)
	go func() {
		if err := crds.server.Start(listenAddr); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("image resolver server error: %v", err)
		}
	}()
}

func (crds *ContainerRuntimeDiscoveryServer) Shutdown(ctx context.Context) error {
	err := crds.server.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	return nil
}

func (crds *ContainerRuntimeDiscoveryServer) ListImages(c echo.Context) error {
	ctx := c.Request().Context()

	images, err := crds.discoverer.Images(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover images: %w", err)
	}

	response := types.ListImagesResponse{
		Images: images,
	}
	// nolint: wrapcheck
	return c.JSON(http.StatusOK, response)
}

func (crds *ContainerRuntimeDiscoveryServer) GetImage(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	image, err := crds.discoverer.Image(ctx, id)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Image not found with ID %v", id))
		}
		return fmt.Errorf("failed to discover image %s: %w", id, err)
	}

	// nolint: wrapcheck
	return c.JSON(http.StatusOK, image)
}

func (crds *ContainerRuntimeDiscoveryServer) ExportImage(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	reader, err := crds.discoverer.ExportImage(ctx, id)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Image not found with ID %v", id))
		}
		return fmt.Errorf("failed to export image %s: %w", id, err)
	}
	defer reader.Close()

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response(), reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (crds *ContainerRuntimeDiscoveryServer) ListContainers(c echo.Context) error {
	containers, err := crds.discoverer.Containers(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to discover containers: %w", err)
	}

	response := types.ListContainersResponse{
		Containers: containers,
	}

	// nolint: wrapcheck
	return c.JSON(http.StatusOK, response)
}

func (crds *ContainerRuntimeDiscoveryServer) GetContainer(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	container, err := crds.discoverer.Container(ctx, id)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Container not found with ID %v", id))
		}
		return fmt.Errorf("failed to discover container %s: %w", id, err)
	}

	// nolint: wrapcheck
	return c.JSON(http.StatusOK, container)
}

func (crds *ContainerRuntimeDiscoveryServer) ExportContainer(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	reader, cleanup, err := crds.discoverer.ExportContainer(ctx, id)
	if err != nil {
		if errors.Is(err, types.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Container not found with ID %v", id))
		}
		return fmt.Errorf("failed to export container %s: %w", id, err)
	}
	defer cleanup()
	defer reader.Close()

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().WriteHeader(http.StatusOK)

	_, err = io.Copy(c.Response(), reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
