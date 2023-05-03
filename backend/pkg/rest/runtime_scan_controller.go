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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/api/server/restapi/operations"
	"github.com/openclarity/kubeclarity/backend/pkg/database"
	runtimescanner "github.com/openclarity/kubeclarity/backend/pkg/runtime_scanner"
	"github.com/openclarity/kubeclarity/backend/pkg/scheduler"
	"github.com/openclarity/kubeclarity/backend/pkg/types"
	_types "github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/slice"
)

/* ### Start Handlers #### */

func (s *Server) PutRuntimeScanStart(params operations.PutRuntimeScanStartParams) middleware.Responder {
	scanConfig, err := s.createScanConfigFromQuickScan(params.Body.Namespaces)
	if err != nil {
		log.Errorf("Failed to create scan config from quick scan: %v", err)
		return operations.NewPutRuntimeScanStartDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	// Return a 409 conflict if the user trys to start a scan when there is
	// already one running, this will prevent us returning a 500 error
	// because the scanChan will not be ready to accept a new
	// configuration.
	status, _ := s.getScanStatusAndScanned()
	if status == models.RuntimeScanStatusINPROGRESS {
		return operations.NewPutRuntimeScanStartDefault(http.StatusConflict).
			WithPayload(&models.APIResponse{
				Message: "Scan already running, stop existing scan before starting a new one.",
			})
	}

	select {
	case s.scanChan <- scanConfig:
	default:
		log.Errorf("Failed to send scan config to channel")
		return operations.NewPutRuntimeScanStartDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewPutRuntimeScanStartCreated()
}

func (s *Server) PutRuntimeScanStop(_ operations.PutRuntimeScanStopParams) middleware.Responder {
	// stop any currently running scan
	s.runtimeScanner.StopCurrentScan()

	return operations.NewPutRuntimeScanStopCreated()
}

func (s *Server) GetRuntimeScanProgress(_ operations.GetRuntimeScanProgressParams) middleware.Responder {
	status, scanned := s.getScanStatusAndScanned()
	scanConfig := s.runtimeScanner.GetScanConfig()
	startTime := s.runtimeScanner.GetLastScanStartTime()

	return operations.NewGetRuntimeScanProgressOK().WithPayload(&models.Progress{
		ScanType:          scanConfig.ScanType,
		Scanned:           &scanned,
		ScannedNamespaces: s.runtimeScanner.GetScannedNamespaces(),
		StartTime:         strfmt.DateTime(startTime),
		Status:            status,
	})
}

func (s *Server) GetRuntimeScanResults(params operations.GetRuntimeScanResultsParams) middleware.Responder {
	state := s.GetState()
	// Fetch last scan config from state (not from DB).
	scanConfig := s.runtimeScanner.GetScanConfig()

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
			CisDockerBenchmarkScanEnabled:   scanConfig.CisDockerBenchmarkScanEnabled,
			Counters:                        &models.RuntimeScanCounters{},
			Failures:                        failures,
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
		CisDockerBenchmarkScanEnabled:   scanConfig.CisDockerBenchmarkScanEnabled,
		Counters:                        counters,
		EndTime:                         strfmt.DateTime(s.runtimeScanner.GetLastScanEndTime()),
		Failures:                        failures,
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

func (s *Server) GetRuntimeScheduleScanConfig(_ operations.GetRuntimeScheduleScanConfigParams) middleware.Responder {
	ret := models.RuntimeScheduleScanConfig{}
	schedulerConf, err := s.dbHandler.SchedulerTable().Get()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return operations.NewGetRuntimeScheduleScanConfigOK().WithPayload(&ret)
		}
		log.Errorf("Failed to get scheduler table from db: %v", err)
		return operations.NewGetRuntimeScheduleScanConfigDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	if err := json.Unmarshal([]byte(schedulerConf.Config), &ret); err != nil {
		log.Errorf("Failed to unmarshal scheduler config: %v. %v", schedulerConf.Config, err)
		return operations.NewGetRuntimeScheduleScanConfigDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewGetRuntimeScheduleScanConfigOK().WithPayload(&ret)
}

func (s *Server) PutRuntimeScheduleScanConfig(params operations.PutRuntimeScheduleScanConfigParams) middleware.Responder {
	if err := s.handleNewScheduleScanConfig(params.Body); err != nil {
		log.Errorf("Failed to handle a new scheduled scan config: %v", err)
		return operations.NewPutRuntimeScheduleScanConfigDefault(http.StatusInternalServerError).
			WithPayload(oopsResponse)
	}

	return operations.NewPutRuntimeScheduleScanConfigCreated()
}

/* ### End Handlers #### */

func (s *Server) saveSchedulerConfigToDB(config *models.RuntimeScheduleScanConfig, startTime time.Time, interval time.Duration) error {
	configB, err := config.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal scheduled config: %v", err)
	}
	if err := s.dbHandler.SchedulerTable().Set(&database.Scheduler{
		ID:           "1", // we want to keep one scheduler config in db.
		NextScanTime: startTime.Format(time.RFC3339),
		Config:       string(configB),
		Interval:     int64(interval),
	}); err != nil {
		return fmt.Errorf("failed to set new scheduler config to db: %v", err)
	}
	return nil
}

func (s *Server) createScanConfigFromQuickScan(namespaces []string) (*runtimescanner.ScanConfig, error) {
	quickScanConfig, err := s.dbHandler.QuickScanConfigTable().Get()
	if err != nil {
		return nil, fmt.Errorf("failed to get quick scan config from db")
	}
	return &runtimescanner.ScanConfig{
		ScanType:                      models.ScanTypeQUICK,
		CisDockerBenchmarkScanEnabled: quickScanConfig.CisDockerBenchmarkScanEnabled,
		MaxScanParallelism:            quickScanConfig.MaxScanParallelism,
		Namespaces:                    namespaces,
	}, nil
}

const (
	secondsInHour = 60 * 60
	secondsInDay  = 24 * secondsInHour
	secondsInWeek = 7 * secondsInDay
)

func getIntervalAndStartTimeFromByDaysScheduleScanConfig(timeNow time.Time, scanConfig *models.ByDaysScheduleScanConfig) (time.Duration, time.Time) {
	interval := time.Duration(scanConfig.DaysInterval*secondsInDay) * time.Second
	hour := int(*scanConfig.TimeOfDay.Hour)
	minute := int(*scanConfig.TimeOfDay.Minute)
	year, month, day := timeNow.Date()

	startTime := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)

	return interval, startTime
}

func getIntervalAndStartTimeFromByHoursScheduleScanConfig(timeNow time.Time, scanConfig *models.ByHoursScheduleScanConfig) (time.Duration, time.Time) {
	interval := time.Duration(scanConfig.HoursInterval*secondsInHour) * time.Second

	return interval, timeNow
}

func getIntervalAndStartTimeFromWeeklyScheduleScanConfig(timeNow time.Time, scanConfig *models.WeeklyScheduleScanConfig) (time.Duration, time.Time) {
	interval := time.Duration(secondsInWeek) * time.Second

	currentDay := timeNow.Weekday() + 1
	diffDays := scanConfig.DayInWeek - int64(currentDay)

	hour := int(*scanConfig.TimeOfDay.Hour)
	minute := int(*scanConfig.TimeOfDay.Minute)
	year, month, day := timeNow.Add(time.Duration(diffDays*secondsInDay) * time.Second).Date()

	startTime := time.Date(year, month, day, hour, minute, 0, 0, time.UTC)

	return interval, startTime
}

func (s *Server) handleNewScheduleScanConfig(config *models.RuntimeScheduleScanConfig) error {
	var interval time.Duration
	var startTime time.Time
	singleScan := false

	timeNow := time.Now().UTC()

	switch config.ScanConfigType().ScheduleScanConfigType() {
	case scheduler.ByDaysScheduleScanConfig:
		// nolint:forcetypeassert
		scanConfig := config.ScanConfigType().(*models.ByDaysScheduleScanConfig)
		interval, startTime = getIntervalAndStartTimeFromByDaysScheduleScanConfig(timeNow, scanConfig)
	case scheduler.ByHoursScheduleScanConfig:
		// nolint:forcetypeassert
		scanConfig := config.ScanConfigType().(*models.ByHoursScheduleScanConfig)
		interval, startTime = getIntervalAndStartTimeFromByHoursScheduleScanConfig(timeNow, scanConfig)
	case scheduler.SingleScheduleScanConfig:
		var err error
		singleScan = true
		// nolint:forcetypeassert
		scanConfig := config.ScanConfigType().(*models.SingleScheduleScanConfig)
		startTime, err = time.Parse(time.RFC3339, scanConfig.OperationTime.String())
		if err != nil {
			return fmt.Errorf("failed to parse operation time: %v. %v", scanConfig.OperationTime.String(), err)
		}
		// set interval to a positive value so we will not crash when starting ticker in Scheduler.spin. This will not be used.
		interval = 1
	case scheduler.WeeklyScheduleScanConfig:
		// nolint:forcetypeassert
		scanConfig := config.ScanConfigType().(*models.WeeklyScheduleScanConfig)
		interval, startTime = getIntervalAndStartTimeFromWeeklyScheduleScanConfig(timeNow, scanConfig)
	default:
		return fmt.Errorf("unsupported schedule config type: %v", config.ScanConfigType().ScheduleScanConfigType())
	}

	if interval <= 0 {
		return fmt.Errorf("parameters validation failed. Interval=%v", interval)
	}
	if err := s.saveSchedulerConfigToDB(config, startTime, interval); err != nil {
		return fmt.Errorf("failed to save scheduler config to DB: %v", err)
	}

	schedParams := &scheduler.Params{
		Namespaces:                    config.Namespaces,
		CisDockerBenchmarkScanEnabled: config.CisDockerBenchmarkScanEnabled,
		MaxScanParallelism:            config.MaxScanParallelism,
		Interval:                      interval,
		StartTime:                     startTime,
		SingleScan:                    singleScan,
	}
	s.scheduler.Schedule(schedParams)
	return nil
}

func (s *Server) getRuntimeScanCounters(filters *database.CountFilters) (*models.RuntimeScanCounters, error) {
	pkgCount, err := s.dbHandler.PackageTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count packages: %v", err)
	}
	appCount, err := s.dbHandler.ApplicationTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %v", err)
	}
	resourceCount, err := s.dbHandler.ResourceTable().Count(filters)
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
		Resources:       uint32(resourceCount),
		Vulnerabilities: uint32(vulCount),
	}, nil
}

func (s *Server) getRuntimeScanCisDockerBenchmarkCounters(filters *database.CountFilters) (*models.CISDockerBenchmarkScanCounters, error) {
	appCount, err := s.dbHandler.ApplicationTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count applications: %v", err)
	}
	resourceCount, err := s.dbHandler.ResourceTable().Count(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count resources: %v", err)
	}
	return &models.CISDockerBenchmarkScanCounters{
		Applications: uint32(appCount),
		Resources:    uint32(resourceCount),
	}, nil
}

func (s *Server) getScanStatusAndScanned() (models.RuntimeScanStatus, int64) {
	var scanned int64
	var status models.RuntimeScanStatus

	scanProgress := s.runtimeScanner.ScanProgress()

	switch scanProgress.Status {
	case _types.Idle:
		status = models.RuntimeScanStatusNOTSTARTED
	case _types.NothingToScan, _types.ScanInitFailure, _types.ScanAborted:
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

type podData struct {
	name      string
	namespace string
}

type scanFailData struct {
	errorMessages []string
	podData       []podData
}

func (s *Server) applyRuntimeScanResults(results []*_types.ImageScanResult) ([]string, []string, error) {
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

	appIDs := make([]string, 0)
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

func getFailures(imageNameToScanFailData map[string]*scanFailData) []string {
	failures := make([]string, 0)
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
func getAppNameFromResult(result *_types.ImageScanResult) string {
	for _, key := range appNameLabelKeys {
		appName := result.PodLabels.Get(key)
		if appName != "" {
			return appName
		}
	}

	return result.PodName
}

// getAppVersionFromResult retrieve app version from recommended labels if exists.
func getAppVersionFromResult(result *_types.ImageScanResult) string {
	appVersion := result.PodLabels.Get("version")
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
