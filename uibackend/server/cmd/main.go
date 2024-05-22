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

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Portshift/go-utils/healthz"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	middleware "github.com/oapi-codegen/echo-middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/version"
	server "github.com/openclarity/vmclarity/uibackend/server"
	apiserver "github.com/openclarity/vmclarity/uibackend/server/internal/server"
)

const (
	LogLevelFlag         = "log-level"
	LogLevelDefaultValue = "warning"
	ExecutableName       = "vmclarity-ui-backend"
)

var (
	logLevel = LogLevelDefaultValue
	rootCmd  = &cobra.Command{
		Use:   ExecutableName,
		Short: "VMClarity UI Backend",
		Long:  "VMClarity UI Backend",
		Version: fmt.Sprintf("Version: %s \nCommit: %s\nBuild Time: %s",
			version.Version, version.CommitHash, version.BuildTimestamp),
		SilenceUsage: true,
	}
)

func init() {
	cmdRun := cobra.Command{
		Use:     "run",
		Run:     runCommand,
		Short:   "Starts the UI Backend",
		Long:    "Starts the VMClarity UI Backend",
		Example: ExecutableName + " run",
	}

	cmdRun.PersistentFlags().StringVar(&logLevel,
		LogLevelFlag,
		LogLevelDefaultValue,
		"Set log level [panic fatal error warning info debug trace]")

	cmdVersion := cobra.Command{
		Use:     "version",
		Run:     versionCommand,
		Short:   "Displays the version",
		Long:    "Displays the version of the VMClarity UI Backend",
		Example: ExecutableName + " version",
	}

	rootCmd.AddCommand(&cmdRun)
	rootCmd.AddCommand(&cmdVersion)
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}

// Main entry point for the orchestrator, triggered by the `run` command.
func runCommand(cmd *cobra.Command, _ []string) {
	log.InitLogger(logLevel, os.Stderr)
	logger := logrus.WithContext(cmd.Context())
	ctx := log.SetLoggerForContext(cmd.Context(), logger)

	ctx, cancel := context.WithCancel(ctx)

	config, err := server.NewConfig()
	if err != nil {
		logger.Fatalf("unable to load configuration")
	}

	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start()
	healthServer.SetIsReady(false)

	client, err := client.New(config.APIServerAddress)
	if err != nil {
		logger.Fatalf("Failed to create an API client: %v", err)
	}
	handler := server.CreateServer(client)
	handler.StartBackgroundProcessing(ctx)

	swagger, err := apiserver.GetSwagger()
	if err != nil {
		logger.Fatalf("failed to load UI swagger spec: %v", err)
	}

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(middleware.OapiRequestValidator(swagger))

	apiserver.RegisterHandlers(e, handler)

	go func() {
		if err := e.Start(config.ListenAddress); err != nil && errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("HTTP server shutdown %v", err)
		}
	}()
	healthServer.SetIsReady(true)

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sig
	logger.Warningf("Received a termination signal: %v", s)

	cancel()

	// nolint:mnd
	ctx, cancel = context.WithTimeout(cmd.Context(), 10*time.Second) //nolint:mnd
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Fatalf("Failed to shutdown server: %v", err)
	}
}

// Command to display the version.
func versionCommand(_ *cobra.Command, _ []string) {
	fmt.Printf("Version: %s \nCommit: %s\nBuild Time: %s",
		version.Version, version.CommitHash, version.BuildTimestamp)
}
