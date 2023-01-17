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
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/openclarity/kubeclarity/backend/pkg/metrics"

	"github.com/Portshift/go-utils/healthz"
	k8sutils "github.com/Portshift/go-utils/k8s"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	_config "github.com/openclarity/kubeclarity/backend/pkg/config"
	_database "github.com/openclarity/kubeclarity/backend/pkg/database"
	"github.com/openclarity/kubeclarity/backend/pkg/rest"
	runtime_scan_models "github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
	runtime_scan_config "github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/fake"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/orchestrator"
)

type Backend struct {
	dbHandler *_database.Handler
}

func CreateBackend(dbHandler *_database.Handler) *Backend {
	return &Backend{
		dbHandler: dbHandler,
	}
}

func createDatabaseConfig(config *_config.Config) *_database.DBConfig {
	return &_database.DBConfig{
		DriverType:                config.DatabaseDriver,
		EnableInfoLogs:            config.EnableDBInfoLogs,
		DBPassword:                config.DBPassword,
		DBUser:                    config.DBUser,
		DBHost:                    config.DBHost,
		DBPort:                    config.DBPort,
		DBName:                    config.DBName,
		ViewRefreshIntervalSecond: config.ViewRefreshIntervalSecond,
	}
}

const defaultChanSize = 100

func Run() {
	config, err := _config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	errChan := make(chan struct{}, defaultChanSize)

	healthServer := healthz.NewHealthServer(config.HealthCheckAddress)
	healthServer.Start()

	healthServer.SetIsReady(false)

	globalCtx, globalCancel := context.WithCancel(context.Background())
	defer globalCancel()

	log.Info("KubeClarity backend is running")

	dbConfig := createDatabaseConfig(config)
	dbHandler := _database.Init(dbConfig)

	go dbHandler.RefreshMaterializedViews()

	if config.EnableFakeData {
		go dbHandler.CreateFakeData()
	}

	backend := CreateBackend(dbHandler)

	var k8sClientset *kubernetes.Clientset
	if !config.EnableFakeRuntimeScanner {
		k8sClientset, _, err = k8sutils.CreateK8sClientset(nil, k8sutils.KubeOptions{})
		if err != nil {
			log.Fatalf("Failed to create K8s clientset: %v", err)
		}
	}

	orc, err := createRuntimeScanOrchestrator(config, k8sClientset, backend.handleImageContentAnalysis)
	if err != nil {
		log.Fatalf("Failed to create runtime scan orchestrator: %v", err)
	}

	restServer, err := rest.CreateRESTServer(config.BackendRestPort, backend.dbHandler, orc, k8sClientset)
	if err != nil {
		log.Fatalf("Failed to create REST server: %v", err)
	}
	restServer.Start(errChan)
	defer restServer.Stop()

	if config.PrometheusRefreshIntervalSeconds > 0 {
		metrics.CreateMetrics(dbHandler, config.PrometheusRefreshIntervalSeconds).StartRecordingMetrics(globalCtx)
	}

	healthServer.SetIsReady(true)
	log.Info("KubeClarity backend is ready")

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

func createRuntimeScanOrchestrator(config *_config.Config, clientset kubernetes.Interface,
	imageContentAnalysisHandler orchestrator.ImageContentAnalysisHandlerCallback,
) (orchestrator.VulnerabilitiesScanner, error) {
	if config.EnableFakeRuntimeScanner {
		return fake.Create(), nil
	}

	orcConfig, err := runtime_scan_config.LoadConfig(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to load runtime scan orchestrator config: %v", err)
	}

	orc, err := orchestrator.Create(orcConfig, clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime scan orchestrator: %v", err)
	}

	orc.SetImageContentAnalysisHandlerCallback(imageContentAnalysisHandler)

	return orc, nil
}

func (b *Backend) handleImageContentAnalysis(contentAnalysis *runtime_scan_models.ImageContentAnalysis) error {
	log.Infof("Handling image content analysis. image-id=%v", contentAnalysis.ImageID)
	if log.GetLevel() == log.TraceLevel {
		contentAnalysisB, err := json.Marshal(contentAnalysis)
		if err == nil {
			log.Tracef("Handling image content analysis. contentAnalysis=%s", contentAnalysisB)
		} else {
			log.Errorf("Failed to marshal content alanysis: %v", err)
		}
	}

	// Create a new resource tree.
	transactionParams := &_database.TransactionParams{
		Analyzers: make(map[_database.ResourcePkgID][]string), // will be populated during object creation
	}

	resource := _database.CreateResourceFromRuntimeContentAnalysis(contentAnalysis.ResourceContentAnalysis, transactionParams)
	// Update new resource information.
	// Since it's a content analysis report we should NOT update PackageVulnerabilities relationships.
	if err := b.dbHandler.ObjectTree().SetResource(resource, transactionParams, false); err != nil {
		return fmt.Errorf("failed to update resource: %v", err)
	}

	return nil
}
