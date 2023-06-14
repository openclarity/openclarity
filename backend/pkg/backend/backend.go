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

package backend

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Portshift/go-utils/healthz"

	_config "github.com/openclarity/vmclarity/backend/pkg/config"
	"github.com/openclarity/vmclarity/backend/pkg/database"
	databaseTypes "github.com/openclarity/vmclarity/backend/pkg/database/types"
	"github.com/openclarity/vmclarity/backend/pkg/rest"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
	uibackend "github.com/openclarity/vmclarity/ui_backend/pkg/rest"
)

func createDatabaseConfig(config *_config.Config) databaseTypes.DBConfig {
	return databaseTypes.DBConfig{
		DriverType:     config.DatabaseDriver,
		EnableInfoLogs: config.EnableDBInfoLogs,
		DBPassword:     config.DBPassword,
		DBUser:         config.DBUser,
		DBHost:         config.DBHost,
		DBPort:         config.DBPort,
		DBName:         config.DBName,
		LocalDBPath:    config.LocalDBPath,
	}
}

const defaultChanSize = 100

func Run(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	config, err := _config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	errChan := make(chan struct{}, defaultChanSize)

	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start()

	healthServer.SetIsReady(false)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger.Info("VMClarity backend is running")

	dbConfig := createDatabaseConfig(config)
	dbHandler, err := database.InitializeDatabase(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to initialise database: %v", err)
	}

	if config.EnableFakeData {
		go database.CreateDemoData(ctx, dbHandler)
	}

	backendAddress := fmt.Sprintf("http://%s%s", net.JoinHostPort(config.BackendRestHost, strconv.Itoa(config.BackendRestPort)), rest.BaseURL)
	backendClient, err := backendclient.Create(backendAddress)
	if err != nil {
		logger.Fatalf("Failed to create a backend client: %v", err)
	}

	uiBackendServer := uibackend.CreateUIBackedServer(backendClient)

	// nolint:contextcheck
	restServer, err := rest.CreateRESTServer(config.BackendRestPort, dbHandler, config.UISitePath, uiBackendServer)
	if err != nil {
		logger.Fatalf("Failed to create REST server: %v", err)
	}
	restServer.Start(ctx, errChan)
	defer restServer.Stop(ctx)

	if config.DisableOrchestrator {
		logger.Infof("Runtime orchestrator is disabled")
	} else {
		if err = startOrchestrator(ctx, config, backendClient); err != nil {
			logger.Fatalf("Failed to start orchestrator: %v", err)
		}
	}

	// Background processing must start after rest server was started.
	uiBackendServer.StartBackgroundProcessing(ctx)

	healthServer.SetIsReady(true)
	logger.Info("VMClarity backend is ready")

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-errChan:
		cancel()
		logger.Errorf("Received an error - shutting down")
	case s := <-sig:
		cancel()
		logger.Warningf("Received a termination signal: %v", s)
	}
}

func startOrchestrator(ctx context.Context, config *_config.Config, client *backendclient.BackendClient) error {
	orchestratorConfig, err := orchestrator.LoadConfig(config.BackendRestHost, config.BackendRestPort, rest.BaseURL)
	if err != nil {
		return fmt.Errorf("failed to load Orchestrator config: %w", err)
	}

	o, err := orchestrator.New(ctx, orchestratorConfig, client)
	if err != nil {
		return fmt.Errorf("failed to initialize Orchestrator: %w", err)
	}

	o.Start(ctx)

	return nil
}
