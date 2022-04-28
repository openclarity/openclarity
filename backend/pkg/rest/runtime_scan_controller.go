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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/cisco-open/kubei/api/server/models"
	"github.com/cisco-open/kubei/api/server/restapi/operations"
	"github.com/cisco-open/kubei/backend/pkg/database"
	"github.com/cisco-open/kubei/backend/pkg/types"
	runtime_scan_config "github.com/cisco-open/kubei/runtime_scan/pkg/config"
	_types "github.com/cisco-open/kubei/runtime_scan/pkg/types"
	"github.com/cisco-open/kubei/shared/pkg/utils/slice"
)

/* ### Start Handlers #### */

func (s *Server) PutRuntimeScanStart(params operations.PutRuntimeScanStartParams) middleware.Responder {
	if err := s.startScan(params.Body.Namespaces); err != nil {
		log.Errorf("Failed to start scan: %v", err)
		return operations.NewPutRuntimeScanStartDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}
	s.scannedNamespaces = params.Body.Namespaces

	return operations.NewPutRuntimeScanStartCreated()
}

func (s *Server) PutRuntimeScanStop(_ operations.PutRuntimeScanStopParams) middleware.Responder {
	// stop scanner
	s.vulnerabilitiesScanner.Clear()
	// stop listening for scanner results
	s.stopCurrentScan()
	// clear scanned namespace list
	s.scannedNamespaces = []string{}

	return operations.NewPutRuntimeScanStopCreated()
}

func (s *Server) GetRuntimeScanProgress(_ operations.GetRuntimeScanProgressParams) middleware.Responder {
	status, scanned := s.getScanStatusAndScanned()

	return operations.NewGetRuntimeScanProgressOK().WithPayload(&models.Progress{
		Scanned:           &scanned,
		ScannedNamespaces: s.scannedNamespaces,
		Status:            status,
	})
}

func (s *Server) GetRuntimeScanResults(params operations.GetRuntimeScanResultsParams) middleware.Responder {
	state := s.GetState()

	failures := make([]*models.RuntimeScanFailure, len(state.runtimeScanFailures))
	for i, failure := range state.runtimeScanFailures {
		failures[i] = &models.RuntimeScanFailure{
			Message: failure,
		}
	}

	// in the case of no application IDs to look by, nothing was scanned, so we can just return
	if len(state.runtimeScanApplicationIDs) == 0 {
		return operations.NewGetRuntimeScanResultsOK().WithPayload(&models.RuntimeScanResults{
			CisDockerBenchmarkCountPerLevel: []*models.CISDockerBenchmarkLevelCount{},
			CisDockerBenchmarkCounters:      &models.CISDockerBenchmarkScanCounters{},
			Counters:                        &models.RuntimeScanCounters{},
			Failures:                        failures,
			ScannedNamespaces:               s.scannedNamespaces,
			VulnerabilityPerSeverity:        []*models.VulnerabilityCount{},
		})
	}

	counters, err := s.getRuntimeScanCounters(&database.CountFilters{
		ApplicationIDs:           state.runtimeScanApplicationIDs,
		VulnerabilitySeverityGte: params.VulnerabilitySeverityGte,
	})
	if err != nil {
		log.Errorf("Failed to get runtime counters: %v", err)
		return operations.NewGetRuntimeScanResultsDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	vulPerSeverity, err := s.dbHandler.VulnerabilityTable().CountPerSeverity(&database.CountFilters{
		ApplicationIDs: state.runtimeScanApplicationIDs,
	})
	if err != nil {
		log.Errorf("Failed to get vulnerability per severity: %v", err)
		return operations.NewGetRuntimeScanResultsDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	cisDockerBenchmarkCounters, err := s.getRuntimeScanCisDockerBenchmarkCounters(&database.CountFilters{
		ApplicationIDs:             state.runtimeScanApplicationIDs,
		CisDockerBenchmarkLevelGte: params.CisDockerBenchmarkLevelGte,
	})
	if err != nil {
		log.Errorf("Failed to get cis docker benchmark counters: %v", err)
		return operations.NewGetRuntimeScanResultsDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	cisDockerBenchmarkCountPerLevel, err := s.dbHandler.CISDockerBenchmarkResultTable().CountPerLevel(&database.CountFilters{
		ApplicationIDs: state.runtimeScanApplicationIDs,
	})
	if err != nil {
		log.Errorf("Failed to get cis docker bencmark issues per level: %v", err)
		return operations.NewGetRuntimeScanResultsDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetRuntimeScanResultsOK().WithPayload(&models.RuntimeScanResults{
		CisDockerBenchmarkCountPerLevel: cisDockerBenchmarkCountPerLevel,
		CisDockerBenchmarkCounters:      cisDockerBenchmarkCounters,
		Counters:                        counters,
		Failures:                        failures,
		ScannedNamespaces:               s.scannedNamespaces,
		VulnerabilityPerSeverity:        vulPerSeverity,
	})
}

func (s *Server) GetNamespaces(_ operations.GetNamespacesParams) middleware.Responder {
	nsList, err := s.GetNamespaceList()
	if err != nil {
		log.Errorf("Failed to get namespace list: %v", err)
		return operations.NewGetNamespacesDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	namespaces := make([]*models.Namespace, len(nsList))
	for i, ns := range nsList {
		namespaces[i] = &models.Namespace{
			Name: ns,
		}
	}

	return operations.NewGetNamespacesOK().WithPayload(namespaces)
}

func (s *Server) GetRuntimeQuickScanConfig(_ operations.GetRuntimeQuickscanConfigParams) middleware.Responder {
	conf, err := s.dbHandler.QuickScanConfigTable().Get()
	if err != nil {
		log.Errorf("Failed to get quick scan config: %v", err)
		return operations.NewGetRuntimeQuickscanConfigDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetRuntimeQuickscanConfigOK().WithPayload(conf)
}

func (s *Server) PutRuntimeQuickScanConfig(params operations.PutRuntimeQuickscanConfigParams) middleware.Responder {
	err := s.dbHandler.QuickScanConfigTable().Set(params.Body)
	if err != nil {
		log.Errorf("Failed to set quick scan config: %v", err)
		return operations.NewPutRuntimeQuickscanConfigDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewPutRuntimeQuickscanConfigCreated()
}

/* ### End Handlers #### */

func (s *Server) getRuntimeScanCounters(filters *database.CountFilters) (*models.RuntimeScanCounters, error) {
	pkgCount, err := s.dbHandler.PackageTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count packages: %v", err)
	}
	appCount, err := s.dbHandler.ApplicationTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %v", err)
	}
	resCount, err := s.dbHandler.ResourceTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count resources: %v", err)
	}
	vulCount, err := s.dbHandler.VulnerabilityTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count vulnerabilities: %v", err)
	}
	return &models.RuntimeScanCounters{
		Applications:    uint32(appCount),
		Packages:        uint32(pkgCount),
		Resources:       uint32(resCount),
		Vulnerabilities: uint32(vulCount),
	}, nil
}

func (s *Server) getRuntimeScanCisDockerBenchmarkCounters(filters *database.CountFilters) (*models.CISDockerBenchmarkScanCounters, error) {
	appCount, err := s.dbHandler.ApplicationTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %v", err)
	}
	resCount, err := s.dbHandler.ResourceTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count resources: %v", err)
	}
	return &models.CISDockerBenchmarkScanCounters{
		Applications: uint32(appCount),
		Resources:    uint32(resCount),
	}, nil
}

func (s *Server) stopCurrentScan() {
	s.lock.Lock()
	close(s.stopScanChan)
	s.stopScanChan = make(chan struct{})
	s.lock.Unlock()
}

func (s *Server) getScanStatusAndScanned() (models.RuntimeScanStatus, int64) {
	var scanned int64
	var status models.RuntimeScanStatus

	scanProgress := s.vulnerabilitiesScanner.ScanProgress()

	switch scanProgress.Status {
	case _types.Idle:
		status = models.RuntimeScanStatusNOTSTARTED
	case _types.NothingToScan, _types.ScanInitFailure:
		status = models.RuntimeScanStatusDONE
	case _types.ScanInit, _types.Scanning:
		status = models.RuntimeScanStatusINPROGRESS
	case _types.DoneScanning:
		state := s.GetState()
		if state.doneApplyingToDB {
			status = models.RuntimeScanStatusDONE
		} else {
			status = models.RuntimeScanStatusFINALIZING
		}
	default:
		log.Errorf("Unsupported status: %v", scanProgress.Status)
		status = models.RuntimeScanStatusNOTSTARTED
	}

	if scanProgress.ImagesToScan == 0 {
		return status, 0
	}

	scanned = int64((float64(scanProgress.ImagesCompletedToScan) / float64(scanProgress.ImagesToScan)) * 100) //nolint:gomnd

	return status, scanned
}

func (s *Server) startScan(namespaces []string) error {
	s.lock.Lock()
	stop := s.stopScanChan
	s.lock.Unlock()

	// clear before start
	s.SetState(&State{
		runtimeScanApplicationIDs: []string{},
		runtimeScanFailures:       []string{},
	})
	s.vulnerabilitiesScanner.Clear()

	quickScanConfig, err := s.dbHandler.QuickScanConfigTable().Get()
	if err != nil {
		return fmt.Errorf("failed to get quick scan config from db: %v", err)
	}

	// need to create scan done channel for every new scan
	done := make(chan struct{})

	if len(namespaces) == 0 {
		// Empty namespaces list should scan all namespaces.
		namespaces = []string{corev1.NamespaceAll}
	}

	err = s.vulnerabilitiesScanner.Scan(&runtime_scan_config.ScanConfig{
		MaxScanParallelism:           10, // nolint:gomnd
		TargetNamespaces:             namespaces,
		IgnoredNamespaces:            nil,
		JobResultTimeout:             10 * time.Minute, // nolint:gomnd
		DeleteJobPolicy:              runtime_scan_config.DeleteJobPolicySuccessful,
		ShouldScanCISDockerBenchmark: quickScanConfig.CisDockerBenchmarkScanEnabled,
	}, done)
	if err != nil {
		return fmt.Errorf("failed to start scan: %v", err)
	}

	go func() {
		select {
		case <-done:
			results := s.vulnerabilitiesScanner.Results()
			if results.Progress.Status == _types.ScanInitFailure {
				log.Errorf("Got scan init failure from results")
				s.SetState(&State{
					runtimeScanApplicationIDs: []string{},
					runtimeScanFailures:       []string{"Scanning initialization failed"},
				})
				return
			}

			if log.GetLevel() == log.TraceLevel {
				resultsB, _ := json.Marshal(results.ImageScanResults)
				log.Tracef("Got scan results: %s", string(resultsB))
			}

			runtimeScanApplicationIDs, runtimeScanFailures, err := s.applyRuntimeScanResults(results.ImageScanResults)
			if err != nil {
				log.Errorf("Failed to apply runtime scan results: %v", err)
				s.SetState(&State{
					runtimeScanApplicationIDs: []string{},
					runtimeScanFailures:       []string{"Failed to apply runtime scan results"},
					doneApplyingToDB:          true,
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
		case <-stop:
			log.Infof("Received a stop signal, stopping scan")
		}
	}()

	return nil
}

type podData struct {
	name      string
	namespace string
}

type scanFailData struct {
	errorMessages []string
	podData       []podData
}

func (s *Server) applyRuntimeScanResults(results []*_types.ImageScanResult) (appIDs []string, failures []string, err error) {
	appNameToScanResults := make(map[string][]*_types.ImageScanResult)
	imageNameToScanFailData := make(map[string]*scanFailData)

	for i, result := range results {
		if !result.Success {
			// Extract failed scans.
			imageNameToScanFailData = setOrUpdateScanFailDataForImage(imageNameToScanFailData, result)
		} else {
			// Unified pods scan results to a single application.
			appName := getAppNameFromResult(result)
			if appVersion := getAppVersionFromResult(result); appVersion != "" {
				appName = appName + "." + appVersion
			}
			appNameToScanResults[appName] = append(appNameToScanResults[appName], results[i])
		}
	}

	for appName, scanResults := range appNameToScanResults {
		app, created, err := s.getOrCreateDBApplicationFromRuntimeScanResults(appName, scanResults)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get or create application: %v", err)
		}

		// TODO: handleApplicationsVulnerabilityScan updates the DB - do we want to do all updates in a single DB transaction?
		// If one application failed to be inserted to the DB to we want to rollback all?
		if err = s.handleApplicationsVulnerabilityScan(app,
			types.ApplicationVulnerabilityScanFromRuntimeScan(scanResults),
			models.VulnerabilitySourceRUNTIME); err != nil {
			// TODO: Do we want to continue to the next pod?
			return nil, nil, fmt.Errorf("failed to handle vulnerability scan report: %v", err)
		}

		if created {
			log.Infof("Successfully created application. id=%v, name=%v", app.ID, app.Name)
		} else {
			log.Infof("Successfully updated application. id=%v, name=%v", app.ID, app.Name)
		}

		appIDs = append(appIDs, app.ID)
	}

	return appIDs, getFailures(imageNameToScanFailData), nil
}

const failureFormat = "Failed to scan image %q.\n" +
	"Effected pods: %s.\n" +
	"Reasons: %s."

func getFailures(imageNameToScanFailData map[string]*scanFailData) (failures []string) {
	for imageName, failData := range imageNameToScanFailData {
		failures = append(failures, fmt.Sprintf(
			failureFormat,
			imageName,
			getPodsList(failData.podData),
			strings.Join(failData.errorMessages, ", ")))
	}
	return failures
}

func getPodsList(podsData []podData) string {
	podList := make([]string, len(podsData))
	for i, p := range podsData {
		podList[i] = fmt.Sprintf("%s/%s", p.name, p.namespace)
	}
	return strings.Join(podList, ", ")
}

// Popular K8s label keys that are used for naming an application.
var appNameLabelKeys = []string{"app", "k8s-app", "app.kubernetes.io/name", "name"}

// getAppNameFromResult retrieve app name from popular K8s app name labels, if not preset use pod name.
func getAppNameFromResult(result *_types.ImageScanResult) (appName string) {
	for _, key := range appNameLabelKeys {
		appName = result.PodLabels.Get(key)
		if appName != "" {
			return appName
		}
	}

	return result.PodName
}

// getAppVersionFromResult retrieve app version from recommended labels if exists.
func getAppVersionFromResult(result *_types.ImageScanResult) (appVersion string) {
	appVersion = result.PodLabels.Get("version")
	if appVersion != "" {
		return appVersion
	}

	appVersion = result.PodLabels.Get("app.kubernetes.io/version")
	if appVersion != "" {
		return appVersion
	}

	return ""
}

func setOrUpdateScanFailDataForImage(imageNameToScanFailData map[string]*scanFailData, result *_types.ImageScanResult) map[string]*scanFailData {
	failData, ok := imageNameToScanFailData[result.ImageName]
	if !ok {
		// Create error messages only once since it's an image scan failure that not related to pod data.
		failData = &scanFailData{
			errorMessages: getErrorMessages(result.ScanErrors),
		}
	}

	// Append pod information.
	failData.podData = append(failData.podData, podData{
		name:      result.PodName,
		namespace: result.PodNamespace,
	})

	// Set new/updated fail data.
	imageNameToScanFailData[result.ImageName] = failData

	return imageNameToScanFailData
}

func getErrorMessages(scanErrors []*_types.ScanError) []string {
	errs := make([]string, len(scanErrors))
	for i, scanError := range scanErrors {
		errs[i] = scanError.ErrMsg
	}
	return errs
}

// getOrCreateDBApplicationFromRuntimeScanResults returns the DB application and a boolean whether a new application was created or not.
func (s *Server) getOrCreateDBApplicationFromRuntimeScanResults(name string, results []*_types.ImageScanResult) (*database.Application, bool, error) {
	appType := models.ApplicationTypePOD
	app := database.CreateApplication(&models.ApplicationInfo{
		Name:         &name,
		Type:         &appType,
		Labels:       getApplicationLabelsFromResults(results),
		Environments: getApplicationEnvironmentsFromResults(results),
	})

	created := false
	dbApp, err := s.dbHandler.ApplicationTable().GetDBApplication(app.ID, true)
	if err != nil {
		log.Infof("Failed to get application - trying to create: %v", err)
		// Create app if not exists
		if err := s.dbHandler.ApplicationTable().Create(app, &database.TransactionParams{}); err != nil {
			return nil, false, fmt.Errorf("failed to create application: %v", err)
		}

		dbApp = app
		created = true
	} else {
		// Application exists, only update labels and environments
		dbApp.Labels = app.Labels
		dbApp.Environments = app.Environments
	}

	return dbApp, created, nil
}

// getApplicationEnvironmentsFromResults use all pods namespaces as application environments.
func getApplicationEnvironmentsFromResults(results []*_types.ImageScanResult) []string {
	envs := make([]string, len(results))
	for i, result := range results {
		envs[i] = result.PodNamespace
	}

	return slice.RemoveStringDuplicates(envs)
}

// getApplicationLabelsFromResults use all pods labels as application labels.
func getApplicationLabelsFromResults(results []*_types.ImageScanResult) []string {
	var labels []string
	for _, result := range results {
		labels = append(labels, strings.Split(result.PodLabels.String(), ",")...)
	}

	return slice.RemoveStringDuplicates(labels)
}
