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

package scanner

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anchore/syft/syft/source"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	kubeclarityConfig "github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	familiesSbom "github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	familiesVulnerabilities "github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
)

// TODO this code is taken from KubeClarity, we can make improvements base on the discussions here: https://github.com/openclarity/vmclarity/pull/3

const TrivyTimeout = 300

// run jobs.
// nolint:cyclop
func (s *Scanner) jobBatchManagement(ctx context.Context) {
	s.Lock()
	targetIDToScanData := s.targetIDToScanData
	numberOfWorkers := 2 // TODO: create this in the API
	s.Unlock()

	// queue of scan data
	q := make(chan *scanData)
	// done channel takes the result of the job
	done := make(chan bool)

	fullScanDone := make(chan bool)

	// spawn workers
	for i := 0; i < numberOfWorkers; i++ {
		go s.worker(ctx, q, i, done, s.killSignal)
	}

	// wait until scan of all instances is done. once all done, notify on fullScanDone chan
	go func() {
		for c := 0; c < len(targetIDToScanData); c++ {
			select {
			case <-done:
			case <-s.killSignal:
				log.WithFields(s.logFields).Debugf("Scan process was canceled - stop waiting for finished jobs")
				return
			}
		}

		fullScanDone <- true
	}()

	// send all scan data on scan data queue, for workers to pick it up.
	for _, data := range targetIDToScanData {
		go func(data *scanData, ks chan bool) {
			select {
			case q <- data:
			case <-ks:
				log.WithFields(s.logFields).Debugf("Scan process was canceled. targetID=%v, scanID=%v", data.targetInstance.TargetID, s.scanID)
				return
			}
		}(data, s.killSignal)
	}

	// wait for killSignal or fullScanDone
	select {
	case <-s.killSignal:
		log.WithFields(s.logFields).Info("Scan process was canceled")
	case <-fullScanDone:
		log.WithFields(s.logFields).Infof("All jobs has finished")
	}
	if err := s.patchScanEndTime(ctx, time.Now()); err != nil {
		log.WithFields(s.logFields).Errorf("Failed to set end time of the scan ID=%s: %v", s.scanID, err)
	}
}

// worker waits for data on the queue, runs a scan job and waits for results from that scan job. Upon completion, done is notified to the caller.
func (s *Scanner) worker(ctx context.Context, queue chan *scanData, workNumber int, done, ks chan bool) {
	for {
		select {
		case data := <-queue:
			// TODO: Run the job only if that target scan status is in init phase, else go to wait
			job, err := s.runJob(ctx, data)
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("failed to run job: %v", err)
				s.Lock()
				data.success = false
				data.completed = true
				s.Unlock()
			} else {
				s.waitForResult(ctx, data, ks)
				// if there was a timeout, concat errMsg
				if data.timeout {
					errMsg = "job has timed out"
				}
			}

			if errMsg != "" {
				log.WithFields(s.logFields).Error(errMsg)
				err := s.SetTargetScanStatusCompletionError(ctx, data.scanResultID, errMsg)
				if err != nil {
					log.WithFields(s.logFields).Errorf("Couldn't set completion error for target scan status. targetID=%v, scanID=%v: %v", data.targetInstance.TargetID, s.scanID, err)
					// TODO: Should we retry?
				}
			}

			s.deleteJobIfNeeded(ctx, &job, data.success, data.completed)

			select {
			case done <- true:
			case <-ks:
				log.WithFields(s.logFields).Infof("Instance scan was canceled. targetID=%v", data.targetInstance.TargetID)
			}
		case <-ks:
			log.WithFields(s.logFields).Debugf("worker #%v halted", workNumber)
			return
		}
	}
}

func (s *Scanner) waitForResult(ctx context.Context, data *scanData, ks chan bool) {
	log.WithFields(s.logFields).Infof("Waiting for result. targetID=%+v", data.targetInstance.TargetID)
	ticker := time.NewTicker(s.config.JobResultsPollingInterval)
	timeout := time.After(s.config.JobResultTimeout)
	for {
		select {
		case <-ticker.C:
			log.WithFields(s.logFields).Infof("Polling scan results for targetID=%v with scanID=%v", data.targetInstance.TargetID, s.scanID)
			// Get scan results from backend
			instanceScanResults, err := s.GetTargetScanStatus(ctx, data.scanResultID)
			if err != nil {
				log.WithFields(s.logFields).Errorf("Failed to get target scan status. scanID=%v, targetID=%s: %v", s.scanID, data.targetInstance.TargetID, err)
				continue
			}
			if *instanceScanResults.General.State != models.DONE {
				log.WithFields(s.logFields).Infof("Scan is not done. scan id=%v, targetID=%s, state=%v", s.scanID, data.targetInstance.TargetID, *instanceScanResults.General.State)
				continue
			}

			s.Lock()
			data.success = !scanStatusHasErrors(instanceScanResults)
			data.completed = true
			s.Unlock()
			return
		case <-timeout:
			log.WithFields(s.logFields).Infof("Job has timed out. targetID=%v", data.targetInstance.TargetID)
			s.Lock()
			data.success = false
			data.completed = true
			data.timeout = true
			s.Unlock()
			return
		case <-ks:
			log.WithFields(s.logFields).Infof("Instance scan was canceled. targetID=%v", data.targetInstance.TargetID)
			return
		}
	}
}

func scanStatusHasErrors(status *models.TargetScanStatus) bool {
	if status.General.Errors != nil && len(*status.General.Errors) > 0 {
		return true
	}

	return false
}

// TODO: need to understand how to destroy the job in case the scanner dies until it gets the results
// We can put the targetID on the scanner VM for easy deletion.
func (s *Scanner) runJob(ctx context.Context, data *scanData) (types.Job, error) {
	var launchInstance types.Instance
	var launchSnapshot types.Snapshot
	var cpySnapshot types.Snapshot
	var snapshot types.Snapshot
	var job types.Job
	var err error

	instanceToScan := data.targetInstance.Instance

	// cleanup in case of an error
	defer func() {
		if err != nil {
			s.deleteJob(ctx, &job)
		}
	}()

	volume, err := instanceToScan.GetRootVolume(ctx)
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to get root volume of an instance %v: %v", instanceToScan.GetID(), err)
	}

	snapshot, err = volume.TakeSnapshot(ctx)
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to take snapshot of a volume: %v", err)
	}
	job.SrcSnapshot = snapshot
	launchSnapshot = snapshot
	if err = snapshot.WaitForReady(ctx); err != nil {
		return types.Job{}, fmt.Errorf("failed to wait for snapshot to be ready. snapshotID=%v: %v", snapshot.GetID(), err)
	}

	if s.config.Region != snapshot.GetRegion() {
		cpySnapshot, err = snapshot.Copy(ctx, s.config.Region)
		if err != nil {
			return types.Job{}, fmt.Errorf("failed to copy snapshot. snapshotID=%v: %v", snapshot.GetID(), err)
		}
		job.DstSnapshot = cpySnapshot
		launchSnapshot = cpySnapshot
		if err = cpySnapshot.WaitForReady(ctx); err != nil {
			return types.Job{}, fmt.Errorf("failed to wait for snapshot to be ready. snapshotID=%v: %v", cpySnapshot.GetID(), err)
		}
	}

	// Scanner job picks the path where the VM volume should be mounted so
	// that the VMClarity CLI config and the provider are synced up on the
	// expected location of the volume on disk.
	volumeMountDirectory := "/vmToBeScanned"
	familiesConfiguration, err := s.generateFamiliesConfigurationYaml(volumeMountDirectory)
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to generate scanner configuration yaml: %w", err)
	}

	scanningJobConfig := provider.ScanningJobConfig{
		ScannerImage:         s.config.ScannerImage,
		ScannerCLIConfig:     familiesConfiguration,
		VolumeMountDirectory: volumeMountDirectory,
		VMClarityAddress:     s.config.ScannerBackendAddress,
		ScanResultID:         data.scanResultID,
	}
	launchInstance, err = s.providerClient.RunScanningJob(ctx, launchSnapshot, scanningJobConfig)
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to launch a new instance: %v", err)
	}
	job.Instance = launchInstance

	return job, nil
}

func (s *Scanner) generateFamiliesConfigurationYaml(scanRootDirectory string) (string, error) {
	famConfig := families.Config{
		SBOM:            userSBOMConfigToFamiliesSbomConfig(s.scanConfig.ScanFamiliesConfig.Sbom, scanRootDirectory),
		Vulnerabilities: userVulnConfigToFamiliesVulnConfig(s.scanConfig.ScanFamiliesConfig.Vulnerabilities),
		// TODO(sambetts) Configure other families once we've got the known working ones working e2e
	}

	famConfigYaml, err := yaml.Marshal(famConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal families config to yaml: %w", err)
	}

	return string(famConfigYaml), nil
}

func userSBOMConfigToFamiliesSbomConfig(sbomConfig *models.SBOMConfig, scanRootDirectory string) familiesSbom.Config {
	if sbomConfig == nil || sbomConfig.Enabled == nil || !*sbomConfig.Enabled {
		return familiesSbom.Config{}
	}
	return familiesSbom.Config{
		Enabled: true,
		// TODO(sambetts) This choice should come from the user's configuration
		AnalyzersList: []string{"syft", "trivy"},
		Inputs: []familiesSbom.Input{
			{
				Input:     scanRootDirectory,
				InputType: string(utils.ROOTFS),
			},
		},
		AnalyzersConfig: &kubeclarityConfig.Config{
			// TODO(sambetts) The user needs to be able to provide this configuration
			Registry: &kubeclarityConfig.Registry{},
			Analyzer: &kubeclarityConfig.Analyzer{
				OutputFormat: "cyclonedx",
				TrivyConfig: kubeclarityConfig.AnalyzerTrivyConfig{
					Timeout: TrivyTimeout,
				},
			},
		},
	}
}

func userVulnConfigToFamiliesVulnConfig(vulnerabilitiesConfig *models.VulnerabilitiesConfig) familiesVulnerabilities.Config {
	if vulnerabilitiesConfig == nil || vulnerabilitiesConfig.Enabled == nil || !*vulnerabilitiesConfig.Enabled {
		return familiesVulnerabilities.Config{}
	}
	return familiesVulnerabilities.Config{
		Enabled: true,
		// TODO(sambetts) This choice should come from the user's configuration
		ScannersList:  []string{"grype", "trivy"},
		InputFromSbom: true,
		ScannersConfig: &kubeclarityConfig.Config{
			// TODO(sambetts) The user needs to be able to provide this configuration
			Registry: &kubeclarityConfig.Registry{},
			Scanner: &kubeclarityConfig.Scanner{
				GrypeConfig: kubeclarityConfig.GrypeConfig{
					// TODO(sambetts) Should run grype in remote mode eventually
					Mode: kubeclarityConfig.ModeLocal,
					LocalGrypeConfig: kubeclarityConfig.LocalGrypeConfig{
						UpdateDB:   true,
						DBRootDir:  "/tmp/",
						ListingURL: "https://toolbox-data.anchore.io/grype/databases/listing.json",
						Scope:      source.SquashedScope,
					},
				},
				TrivyConfig: kubeclarityConfig.ScannerTrivyConfig{
					Timeout: TrivyTimeout,
				},
			},
		},
	}
}

func (s *Scanner) deleteJobIfNeeded(ctx context.Context, job *types.Job, isSuccessfulJob, isCompletedJob bool) {
	if job == nil {
		return
	}

	// delete uncompleted jobs - scan process was canceled
	if !isCompletedJob {
		s.deleteJob(ctx, job)
		return
	}

	switch s.config.DeleteJobPolicy {
	case config.DeleteJobPolicyNever:
		// do nothing
	case config.DeleteJobPolicyAll:
		s.deleteJob(ctx, job)
	case config.DeleteJobPolicySuccessful:
		if isSuccessfulJob {
			s.deleteJob(ctx, job)
		}
	}
}

func (s *Scanner) deleteJob(ctx context.Context, job *types.Job) {
	if job.Instance != nil {
		if err := job.Instance.Delete(ctx); err != nil {
			log.Errorf("Failed to delete instance. instanceID=%v: %v", job.Instance.GetID(), err)
		}
	}
	if job.SrcSnapshot != nil {
		if err := job.SrcSnapshot.Delete(ctx); err != nil {
			log.Errorf("Failed to delete source snapshot. snapshotID=%v: %v", job.SrcSnapshot.GetID(), err)
		}
	}
	if job.DstSnapshot != nil {
		if err := job.DstSnapshot.Delete(ctx); err != nil {
			log.Errorf("Failed to delete destination snapshot. snapshotID=%v: %v", job.DstSnapshot.GetID(), err)
		}
	}
}

// nolint:cyclop
func (s *Scanner) createInitTargetScanStatus(ctx context.Context, scanID, targetID string) (string, error) {
	initScanStatus := &models.TargetScanStatus{
		Exploits: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Exploits.Enabled),
		},
		General: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(true),
		},
		Malware: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Malware.Enabled),
		},
		Misconfigurations: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Misconfigurations.Enabled),
		},
		Rootkits: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Rootkits.Enabled),
		},
		Sbom: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Sbom.Enabled),
		},
		Secrets: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Secrets.Enabled),
		},
		Vulnerabilities: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusStateFromEnabled(*s.scanConfig.ScanFamiliesConfig.Vulnerabilities.Enabled),
		},
	}
	scanResult := models.TargetScanResult{
		ScanId:   scanID,
		Status:   initScanStatus,
		TargetId: targetID,
	}
	resp, err := s.backendClient.PostScanResultsWithResponse(ctx, scanResult)
	if err != nil {
		return "", fmt.Errorf("failed to post scan status: %v", err)
	}
	switch resp.StatusCode() {
	case http.StatusCreated:
		if resp.JSON201 == nil {
			return "", fmt.Errorf("failed to create a scan status, empty body")
		}
		if resp.JSON201.Id == nil {
			return "", fmt.Errorf("failed to create a scan status, missing id")
		}
		return *resp.JSON201.Id, nil
	case http.StatusConflict:
		if resp.JSON409 == nil {
			return "", fmt.Errorf("failed to create a scan status, empty body on conflict")
		}
		if resp.JSON409.Id == nil {
			return "", fmt.Errorf("failed to create a scan status, missing id")
		}
		log.Infof("Scan results already exist with id %v.", *resp.JSON409.Id)
		return *resp.JSON409.Id, nil
	default:
		if resp.JSONDefault != nil && resp.JSONDefault.Message != nil {
			return "", fmt.Errorf("failed to create a scan status. status code=%v: %v", resp.StatusCode(), resp.JSONDefault.Message)
		}
		return "", fmt.Errorf("failed to create a scan status. status code=%v", resp.StatusCode())
	}
}

func getInitScanStatusStateFromEnabled(enabled bool) *models.TargetScanStateState {
	if enabled {
		initState := models.INIT
		return &initState
	}

	notScannedState := models.NOTSCANNED
	return &notScannedState
}
