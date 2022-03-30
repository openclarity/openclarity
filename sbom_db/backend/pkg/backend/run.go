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

package backend

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Portshift/go-utils/healthz"
	log "github.com/sirupsen/logrus"

	_config "github.com/cisco-open/kubei/sbom_db/backend/pkg/config"
	"github.com/cisco-open/kubei/sbom_db/backend/pkg/database"
)

const defaultChanSize = 100

func Run() {
	config, err := _config.LoadConfig()
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		return
	}
	errChan := make(chan struct{}, defaultChanSize)

	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start()

	healthServer.SetIsReady(false)

	_, globalCancel := context.WithCancel(context.Background())
	defer globalCancel()

	log.Info("KubeClarity SBOM DB backend is running")

	dbHandler := database.InitDataBase()

	if config.EnableFakeData {
		go dbHandler.CreateFakeData()
	}

	restServer, err := CreateRESTServer(config.BackendRestPort, dbHandler)
	if err != nil {
		log.Fatalf("Failed to create REST server: %v", err)
	}
	restServer.Start(errChan)
	defer restServer.Stop()

	healthServer.SetIsReady(true)
	log.Info("KubeClarity SBOM DB backend is ready")

	// Wait for deactivation
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-errChan:
		log.Errorf("Received an error - shutting down")
	case s := <-sig:
		log.Warningf("Received a termination signal: %v", s)
	}
}
