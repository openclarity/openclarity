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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/kubeclarity/api/server/restapi"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/common"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
	runtimescanner "github.com/openclarity/kubeclarity/backend/pkg/runtime_scanner"
	"github.com/openclarity/kubeclarity/backend/pkg/scheduler"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/orchestrator"
	_types "github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
)

type Server struct {
	server         *restapi.Server
	dbHandler      database.Database
	clientset      kubernetes.Interface
	runtimeScanner runtimescanner.Scanner
	scheduler      *scheduler.Scheduler
	// Channel to initiate the actual scan according to schedule.
	scanChan chan *runtimescanner.ScanConfig
	// Terminate all go routines.
	stopChan chan struct{}
	// Scan results will be received through this channel.
	resultsChan chan *_types.ScanResults
	State
	lock sync.RWMutex
}

type State struct {
	// List of application IDs that were affected by the last runtime scan.
	runtimeScanApplicationIDs []string
	// List of image scan failures details that were affected by the last runtime scan.
	runtimeScanFailures []string
	doneApplyingToDB    bool
}

func (s *Server) GetState() *State {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return &State{
		runtimeScanApplicationIDs: s.runtimeScanApplicationIDs,
		runtimeScanFailures:       s.runtimeScanFailures,
		doneApplyingToDB:          s.doneApplyingToDB,
	}
}

func (s *Server) SetState(state *State) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.runtimeScanFailures = state.runtimeScanFailures
	s.runtimeScanApplicationIDs = state.runtimeScanApplicationIDs
	s.doneApplyingToDB = state.doneApplyingToDB
}

func (s *Server) GetNamespaceList() ([]string, error) {
	nsList, err := s.clientset.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	ret := make([]string, len(nsList.Items))
	for i, ns := range nsList.Items {
		ret[i] = ns.Name
	}
	return ret, nil
}

func CreateRESTServer(port int, dbHandler *database.Handler, scanner orchestrator.VulnerabilitiesScanner,
	clientset kubernetes.Interface,
) (*Server, error) {
	scanChan := make(chan *runtimescanner.ScanConfig)
	resultsChan := make(chan *_types.ScanResults)
	s := &Server{
		dbHandler:      dbHandler,
		clientset:      clientset,
		runtimeScanner: runtimescanner.CreateRuntimeScanner(scanner, scanChan, resultsChan),
		resultsChan:    resultsChan,
		scanChan:       scanChan,
		stopChan:       make(chan struct{}),
		lock:           sync.RWMutex{},
		scheduler:      scheduler.CreateScheduler(scanChan, dbHandler),
	}
	s.scheduler.Init()

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load swagger spec: %v", err)
	}

	api := operations.NewKubeClarityAPIsAPI(swaggerSpec)

	api.GetApplicationResourcesHandler = operations.GetApplicationResourcesHandlerFunc(func(params operations.GetApplicationResourcesParams) middleware.Responder {
		return s.GetApplicationResources(params)
	})

	api.GetApplicationResourcesIDHandler = operations.GetApplicationResourcesIDHandlerFunc(func(params operations.GetApplicationResourcesIDParams) middleware.Responder {
		return s.GetApplicationResource(params.ID)
	})

	api.GetApplicationsHandler = operations.GetApplicationsHandlerFunc(func(params operations.GetApplicationsParams) middleware.Responder {
		return s.GetApplications(params)
	})

	api.PostApplicationsHandler = operations.PostApplicationsHandlerFunc(func(params operations.PostApplicationsParams) middleware.Responder {
		return s.CreateApplication(params)
	})

	api.GetApplicationsIDHandler = operations.GetApplicationsIDHandlerFunc(func(params operations.GetApplicationsIDParams) middleware.Responder {
		return s.GetApplication(params.ID)
	})

	api.DeleteApplicationsIDHandler = operations.DeleteApplicationsIDHandlerFunc(func(params operations.DeleteApplicationsIDParams) middleware.Responder {
		return s.DeleteApplication(params.ID)
	})

	api.PutApplicationsIDHandler = operations.PutApplicationsIDHandlerFunc(func(params operations.PutApplicationsIDParams) middleware.Responder {
		return s.UpdateApplication(params)
	})

	api.GetPackagesHandler = operations.GetPackagesHandlerFunc(func(params operations.GetPackagesParams) middleware.Responder {
		return s.GetPackages(params)
	})

	api.GetPackagesIDHandler = operations.GetPackagesIDHandlerFunc(func(params operations.GetPackagesIDParams) middleware.Responder {
		return s.GetPackage(params.ID)
	})

	api.GetPackagesIDApplicationResourcesHandler = operations.GetPackagesIDApplicationResourcesHandlerFunc(func(params operations.GetPackagesIDApplicationResourcesParams) middleware.Responder {
		return s.GetPackageApplicationResources(params)
	})

	api.GetVulnerabilitiesHandler = operations.GetVulnerabilitiesHandlerFunc(func(params operations.GetVulnerabilitiesParams) middleware.Responder {
		return s.GetVulnerabilities(params)
	})

	api.GetVulnerabilitiesVulIDPkgIDHandler = operations.GetVulnerabilitiesVulIDPkgIDHandlerFunc(func(params operations.GetVulnerabilitiesVulIDPkgIDParams) middleware.Responder {
		return s.GetVulnerabilitiesVulIDPkgID(params.VulID, params.PkgID)
	})

	api.GetDashboardVulnerabilitiesWithFixHandler = operations.GetDashboardVulnerabilitiesWithFixHandlerFunc(func(params operations.GetDashboardVulnerabilitiesWithFixParams) middleware.Responder {
		return s.GetDashboardVulnerabilitiesWithFix(params)
	})

	api.GetDashboardPackagesPerLanguageHandler = operations.GetDashboardPackagesPerLanguageHandlerFunc(func(params operations.GetDashboardPackagesPerLanguageParams) middleware.Responder {
		return s.GetDashboardPackagesPerLanguage(params)
	})

	api.GetDashboardPackagesPerLicenseHandler = operations.GetDashboardPackagesPerLicenseHandlerFunc(func(params operations.GetDashboardPackagesPerLicenseParams) middleware.Responder {
		return s.GetDashboardPackagesPerLicense(params)
	})

	api.GetDashboardCountersHandler = operations.GetDashboardCountersHandlerFunc(func(params operations.GetDashboardCountersParams) middleware.Responder {
		return s.GetDashboardCounters(params)
	})

	api.GetDashboardMostVulnerableHandler = operations.GetDashboardMostVulnerableHandlerFunc(func(params operations.GetDashboardMostVulnerableParams) middleware.Responder {
		return s.GetDashboardMostVulnerable(params)
	})

	api.GetDashboardTrendsVulnerabilitiesHandler = operations.GetDashboardTrendsVulnerabilitiesHandlerFunc(func(params operations.GetDashboardTrendsVulnerabilitiesParams) middleware.Responder {
		return s.GetDashboardTrendsVulnerabilities(params)
	})

	api.PostApplicationsVulnerabilityScanIDHandler = operations.PostApplicationsVulnerabilityScanIDHandlerFunc(func(params operations.PostApplicationsVulnerabilityScanIDParams) middleware.Responder {
		return s.PostApplicationsVulnerabilityScan(params)
	})

	api.PostApplicationsContentAnalysisIDHandler = operations.PostApplicationsContentAnalysisIDHandlerFunc(func(params operations.PostApplicationsContentAnalysisIDParams) middleware.Responder {
		return s.PostApplicationsContentAnalysis(params)
	})

	api.PutRuntimeScanStartHandler = operations.PutRuntimeScanStartHandlerFunc(func(params operations.PutRuntimeScanStartParams) middleware.Responder {
		return s.PutRuntimeScanStart(params)
	})

	api.PutRuntimeScanStopHandler = operations.PutRuntimeScanStopHandlerFunc(func(params operations.PutRuntimeScanStopParams) middleware.Responder {
		return s.PutRuntimeScanStop(params)
	})

	api.GetRuntimeScanProgressHandler = operations.GetRuntimeScanProgressHandlerFunc(func(params operations.GetRuntimeScanProgressParams) middleware.Responder {
		return s.GetRuntimeScanProgress(params)
	})

	api.GetRuntimeScanResultsHandler = operations.GetRuntimeScanResultsHandlerFunc(func(params operations.GetRuntimeScanResultsParams) middleware.Responder {
		return s.GetRuntimeScanResults(params)
	})

	api.GetNamespacesHandler = operations.GetNamespacesHandlerFunc(func(params operations.GetNamespacesParams) middleware.Responder {
		return s.GetNamespaces(params)
	})

	api.GetRuntimeQuickscanConfigHandler = operations.GetRuntimeQuickscanConfigHandlerFunc(func(params operations.GetRuntimeQuickscanConfigParams) middleware.Responder {
		return s.GetRuntimeQuickScanConfig(params)
	})

	api.PutRuntimeQuickscanConfigHandler = operations.PutRuntimeQuickscanConfigHandlerFunc(func(params operations.PutRuntimeQuickscanConfigParams) middleware.Responder {
		return s.PutRuntimeQuickScanConfig(params)
	})

	api.GetRuntimeScheduleScanConfigHandler = operations.GetRuntimeScheduleScanConfigHandlerFunc(func(params operations.GetRuntimeScheduleScanConfigParams) middleware.Responder {
		return s.GetRuntimeScheduleScanConfig(params)
	})

	api.PutRuntimeScheduleScanConfigHandler = operations.PutRuntimeScheduleScanConfigHandlerFunc(func(params operations.PutRuntimeScheduleScanConfigParams) middleware.Responder {
		return s.PutRuntimeScheduleScanConfig(params)
	})

	api.GetCisdockerbenchmarkresultsIDHandler = operations.GetCisdockerbenchmarkresultsIDHandlerFunc(func(params operations.GetCisdockerbenchmarkresultsIDParams) middleware.Responder {
		return s.GetCISDockerBenchmarkResults(params)
	})

	server := restapi.NewServer(api)

	server.ConfigureFlags()
	server.ConfigureAPI()
	server.Port = port

	s.server = server

	return s, nil
}

func (s *Server) Start(errChan chan struct{}) {
	s.runtimeScanner.Start(s.stopChan)
	s.startListeningForScanResults(s.stopChan)

	log.Infof("Starting REST server")
	go func() {
		if err := s.server.Serve(); err != nil {
			log.Errorf("Failed to serve REST server: %v", err)
			errChan <- common.Empty
		}
	}()
	log.Infof("REST server is running")
}

func (s *Server) Stop() {
	close(s.stopChan)

	log.Infof("Stopping REST server")
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			log.Errorf("Failed to shutdown REST server: %v", err)
		}
	}
}

func (s *Server) startListeningForScanResults(stopChan chan struct{}) {
	go func() {
		for {
			select {
			case results := <-s.resultsChan:
				s.handleScanResults(results)
			case <-stopChan:
				log.Info("Received stop event, stop listening to results")
				return
			}
		}
	}()
}

func (s *Server) handleScanResults(results *_types.ScanResults) {
	// clear before start
	s.SetState(&State{
		runtimeScanApplicationIDs: []string{},
		runtimeScanFailures:       []string{},
		doneApplyingToDB:          false,
	})

	if results.Progress.Status == _types.ScanInitFailure {
		log.Errorf("Got scan init failure from results")
		s.SetState(&State{
			runtimeScanFailures: []string{"Scanning initialization failed"},
		})
		return
	}

	if results.Progress.Status == _types.ScanAborted {
		log.Infof("Got scan aborted from results")
		s.SetState(&State{
			runtimeScanFailures: []string{"Scan aborted"},
		})
		return
	}

	if log.GetLevel() == log.TraceLevel {
		resultsB, err := json.Marshal(results.ImageScanResults)
		if err == nil {
			log.Tracef("Got scan results: %s", string(resultsB))
		} else {
			log.Errorf("Failed to marshal image scan results: %v", err)
		}
	}

	runtimeScanApplicationIDs, runtimeScanFailures, err := s.applyRuntimeScanResults(results.ImageScanResults)
	if err != nil {
		log.Errorf("Failed to apply runtime scan results: %v", err)
		s.SetState(&State{
			runtimeScanFailures: []string{"Failed to apply runtime scan results"},
			doneApplyingToDB:    true,
		})
		return
	}

	// Set new runtime scan application IDs and scan failures details on state.
	s.SetState(&State{
		runtimeScanApplicationIDs: runtimeScanApplicationIDs,
		runtimeScanFailures:       runtimeScanFailures,
		doneApplyingToDB:          true,
	})
	log.Infof("Succeeded to apply runtime scan results. app ids=%+v, failures=%+v",
		runtimeScanApplicationIDs, runtimeScanFailures)
}
