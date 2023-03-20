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

package scanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/anchore/syft/syft/source"
	kubeclarityConfig "github.com/openclarity/kubeclarity/shared/pkg/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	runtimeScanUtils "github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/families"
	familiesExploits "github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	exploitsCommon "github.com/openclarity/vmclarity/shared/pkg/families/exploits/common"
	exploitdbConfig "github.com/openclarity/vmclarity/shared/pkg/families/exploits/exploitdb/config"
	misconfigurationTypes "github.com/openclarity/vmclarity/shared/pkg/families/misconfiguration/types"
	familiesSbom "github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets/common"
	gitleaksconfig "github.com/openclarity/vmclarity/shared/pkg/families/secrets/gitleaks/config"
	familiesVulnerabilities "github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
)

// TODO this code is taken from KubeClarity, we can make improvements base on the discussions here: https://github.com/openclarity/vmclarity/pull/3

const (
	TrivyTimeout = 300

	SnapshotCreationTimeout = 3 * time.Minute
	SnapshotCopyTimeout     = 15 * time.Minute
)

// run jobs.
// nolint:cyclop
func (s *Scanner) jobBatchManagement(ctx context.Context) {
	s.Lock()
	targetIDToScanData := s.targetIDToScanData
	// Since this value has a default in the API, I assume it is safe to dereference it.
	numberOfWorkers := *s.scanConfig.MaxParallelScanners
	s.Unlock()

	// queue of scan data
	q := make(chan *scanData)
	// done channel takes the result of the job
	done := make(chan string)

	// spawn workers
	for i := 0; i < numberOfWorkers; i++ {
		go s.worker(ctx, q, i, done, s.killSignal)
	}

	// send all scan data on scan data queue, for workers to pick it up.
	go func() {
		for _, data := range targetIDToScanData {
			select {
			case q <- data:
			case <-s.killSignal:
				log.WithFields(s.logFields).Debugf("Scan process was canceled. targetID=%v, scanID=%v", data.targetInstance.TargetID, s.scanID)
				return
			}
		}
	}()

	anyJobsFailed := false
	numberOfCompletedJobs := 0
	scanComplete := false
	for !scanComplete {
		var scan *models.Scan
		var err error
		select {
		case targetID := <-done:
			numberOfCompletedJobs = numberOfCompletedJobs + 1
			data := targetIDToScanData[targetID]
			if !data.success {
				anyJobsFailed = true
			}

			scan, err = s.createScanWithUpdatedSummary(ctx, *data)
			if err != nil {
				log.WithFields(s.logFields).Errorf("Failed to create a scan with updated summary: %v", err)
				scan = &models.Scan{}
			}

			if numberOfCompletedJobs == len(targetIDToScanData) {
				scanComplete = true

				state := models.ScanStateDone
				stateMessage := "All scan jobs completed"
				stateReason := models.ScanStateReasonSuccess
				if anyJobsFailed {
					state = models.ScanStateFailed
					stateMessage = "One or more ScanJobs failed"
					stateReason = models.ScanStateReasonOneOrMoreTargetFailedToScan
				}
				t := time.Now()
				scan.EndTime = &t
				scan.State = &state
				scan.StateMessage = &stateMessage
				scan.StateReason = &stateReason
			}
		case <-s.killSignal:
			t := time.Now()
			reason := models.ScanStateReasonTimedOut
			scan = &models.Scan{
				EndTime:      &t,
				State:        runtimeScanUtils.PointerTo(models.ScanStateFailed),
				StateMessage: runtimeScanUtils.StringPtr("Scan was canceled or timed out"),
				StateReason:  &reason,
			}
			scanComplete = true
			log.WithFields(s.logFields).Debugf("Scan process was canceled - stop waiting for finished jobs")
		}

		// regardless of success or failure we need to patch the scan status
		err = s.backendClient.PatchScan(ctx, s.scanID, scan)
		if err != nil {
			log.WithFields(s.logFields).Errorf("failed to patch the scan ID=%s: %v", s.scanID, err)
		}
	}
}

func (s *Scanner) createScanWithUpdatedSummary(ctx context.Context, data scanData) (*models.Scan, error) {
	scan, err := s.backendClient.GetScan(ctx, s.scanID, models.GetScansScanIDParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get scan to update status: %v", err)
	}

	scanResultSummary, err := s.backendClient.GetScanResultSummary(ctx, data.scanResultID)
	if err != nil {
		return nil, fmt.Errorf("failed to get result summary to update status: %v", err)
	}

	// Update the scan summary with the summary from the completed scan result
	scan.Summary.JobsCompleted = runtimeScanUtils.IntPtr(*scan.Summary.JobsCompleted + 1)
	scan.Summary.JobsLeftToRun = runtimeScanUtils.IntPtr(*scan.Summary.JobsLeftToRun - 1)
	scan.Summary.TotalExploits = runtimeScanUtils.IntPtr(*scan.Summary.TotalExploits + *scanResultSummary.TotalExploits)
	scan.Summary.TotalMalware = runtimeScanUtils.IntPtr(*scan.Summary.TotalMalware + *scanResultSummary.TotalMalware)
	scan.Summary.TotalMisconfigurations = runtimeScanUtils.IntPtr(*scan.Summary.TotalMisconfigurations + *scanResultSummary.TotalMisconfigurations)
	scan.Summary.TotalPackages = runtimeScanUtils.IntPtr(*scan.Summary.TotalPackages + *scanResultSummary.TotalPackages)
	scan.Summary.TotalRootkits = runtimeScanUtils.IntPtr(*scan.Summary.TotalRootkits + *scanResultSummary.TotalRootkits)
	scan.Summary.TotalSecrets = runtimeScanUtils.IntPtr(*scan.Summary.TotalSecrets + *scanResultSummary.TotalSecrets)
	scan.Summary.TotalVulnerabilities = &models.VulnerabilityScanSummary{
		TotalCriticalVulnerabilities:   runtimeScanUtils.IntPtr(*scan.Summary.TotalVulnerabilities.TotalCriticalVulnerabilities + *scanResultSummary.TotalVulnerabilities.TotalCriticalVulnerabilities),
		TotalHighVulnerabilities:       runtimeScanUtils.IntPtr(*scan.Summary.TotalVulnerabilities.TotalHighVulnerabilities + *scanResultSummary.TotalVulnerabilities.TotalHighVulnerabilities),
		TotalLowVulnerabilities:        runtimeScanUtils.IntPtr(*scan.Summary.TotalVulnerabilities.TotalLowVulnerabilities + *scanResultSummary.TotalVulnerabilities.TotalLowVulnerabilities),
		TotalMediumVulnerabilities:     runtimeScanUtils.IntPtr(*scan.Summary.TotalVulnerabilities.TotalMediumVulnerabilities + *scanResultSummary.TotalVulnerabilities.TotalMediumVulnerabilities),
		TotalNegligibleVulnerabilities: runtimeScanUtils.IntPtr(*scan.Summary.TotalVulnerabilities.TotalCriticalVulnerabilities + *scanResultSummary.TotalVulnerabilities.TotalNegligibleVulnerabilities),
	}

	return scan, nil
}

// worker waits for data on the queue, runs a scan job and waits for results from that scan job. Upon completion, done is notified to the caller.
func (s *Scanner) worker(ctx context.Context, queue chan *scanData, workNumber int, done chan string, ks chan bool) {
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
			case done <- data.targetInstance.TargetID:
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
			instanceScanResults, err := s.backendClient.GetScanResultStatus(ctx, data.scanResultID)
			if err != nil {
				log.WithFields(s.logFields).Errorf("Failed to get target scan status. scanID=%v, targetID=%s: %v", s.scanID, data.targetInstance.TargetID, err)
				continue
			}
			if *instanceScanResults.General.State != models.DONE {
				log.WithFields(s.logFields).Infof("Scan is not done. scan result id=%v, scan id=%v, targetID=%s, state=%v", data.scanResultID,
					s.scanID, data.targetInstance.TargetID, *instanceScanResults.General.State)
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
// nolint:cyclop
func (s *Scanner) runJob(ctx context.Context, data *scanData) (types.Job, error) {
	var launchInstance types.Instance
	var launchSnapshot types.Snapshot
	var cpySnapshot types.Snapshot
	var snapshot types.Snapshot
	var job types.Job
	var err error

	instanceToScan := data.targetInstance.Instance
	log.WithFields(s.logFields).Infof("Running scanner job for instance id %v", instanceToScan.GetID())

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

	waitContext, waitCancel := context.WithTimeout(ctx, SnapshotCreationTimeout)
	defer waitCancel()
	if err = snapshot.WaitForReady(waitContext); err != nil {
		return types.Job{}, fmt.Errorf("failed to wait for snapshot to be ready. snapshotID=%v: %v", snapshot.GetID(), err)
	}

	// we need the snapshot to be in the scanner region in order to create
	// a volume and attach it.
	if s.config.Region != snapshot.GetRegion() {
		cpySnapshot, err = snapshot.Copy(ctx, s.config.Region)
		if err != nil {
			return types.Job{}, fmt.Errorf("failed to copy snapshot. snapshotID=%v: %v", snapshot.GetID(), err)
		}
		job.DstSnapshot = cpySnapshot
		launchSnapshot = cpySnapshot

		// Copying snapshots between regions can take much longer than
		// creating a snapshot normally
		waitContext, waitCancel := context.WithTimeout(ctx, SnapshotCopyTimeout)
		defer waitCancel()
		if err = cpySnapshot.WaitForReady(waitContext); err != nil {
			return types.Job{}, fmt.Errorf("failed to wait for snapshot to be ready. snapshotID=%v: %v", cpySnapshot.GetID(), err)
		}
	}

	familiesConfiguration, err := s.generateFamiliesConfigurationYaml()
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to generate scanner configuration yaml: %w", err)
	}

	scanningJobConfig := provider.ScanningJobConfig{
		ScannerImage:                  s.config.ScannerImage,
		ScannerCLIConfig:              familiesConfiguration,
		VMClarityAddress:              s.config.ScannerBackendAddress,
		ScanResultID:                  data.scanResultID,
		KeyPairName:                   s.config.ScannerKeyPairName,
		ScannerInstanceCreationConfig: s.scanConfig.ScannerInstanceCreationConfig,
	}
	launchInstance, err = s.providerClient.RunScanningJob(ctx, launchSnapshot.GetRegion(), launchSnapshot.GetID(), scanningJobConfig)
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to launch a new instance: %v", err)
	}
	job.Instance = launchInstance

	// create a volume from the snapshot.
	newVolume, err := launchSnapshot.CreateVolume(ctx, launchInstance.GetAvailabilityZone())
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to create volume: %v", err)
	}

	// wait for instance to be in a running state.
	if err := job.Instance.WaitForReady(ctx); err != nil {
		return types.Job{}, fmt.Errorf("failed to wait for instance ready: %v", err)
	}

	// wait for volume to be available.
	if err := newVolume.WaitForReady(ctx); err != nil {
		return types.Job{}, fmt.Errorf("failed to wait for volume to be ready: %v", err)
	}

	// attach the volume to the scanning job instance.
	err = launchInstance.AttachVolume(ctx, newVolume, s.config.DeviceName)
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to attach volume: %v", err)
	}

	// wait for the volume to be attached.
	if err := newVolume.WaitForAttached(ctx); err != nil {
		return types.Job{}, fmt.Errorf("failed to wait for volume attached: %v", err)
	}

	// mark attached state in the backend.
	err = s.backendClient.PatchTargetScanStatus(ctx, data.scanResultID, &models.TargetScanStatus{
		General: &models.TargetScanState{
			State: runtimeScanUtils.PointerTo(models.ATTACHED),
		},
	})
	if err != nil {
		return types.Job{}, fmt.Errorf("failed to patch target scan status: %v", err)
	}

	return job, nil
}

func (s *Scanner) generateFamiliesConfigurationYaml() (string, error) {
	famConfig := families.Config{
		SBOM:             userSBOMConfigToFamiliesSbomConfig(s.scanConfig.ScanFamiliesConfig.Sbom),
		Vulnerabilities:  userVulnConfigToFamiliesVulnConfig(s.scanConfig.ScanFamiliesConfig.Vulnerabilities),
		Secrets:          userSecretsConfigToFamiliesSecretsConfig(s.scanConfig.ScanFamiliesConfig.Secrets, s.config.GitleaksBinaryPath),
		Exploits:         userExploitsConfigToFamiliesExploitsConfig(s.scanConfig.ScanFamiliesConfig.Exploits, s.config.ExploitsDBAddress),
		Misconfiguration: userMisconfigurationConfigToFamiliesMisconfigurationConfig(s.scanConfig.ScanFamiliesConfig.Misconfigurations),
		// TODO(sambetts) Configure other families once we've got the known working ones working e2e
	}

	famConfigYaml, err := yaml.Marshal(famConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal families config to yaml: %w", err)
	}

	return string(famConfigYaml), nil
}

func userSecretsConfigToFamiliesSecretsConfig(secretsConfig *models.SecretsConfig, gitleaksBinaryPath string) secrets.Config {
	if secretsConfig == nil || secretsConfig.Enabled == nil || !*secretsConfig.Enabled {
		return secrets.Config{}
	}
	return secrets.Config{
		Enabled: true,
		// TODO(idanf) This choice should come from the user's configuration
		ScannersList: []string{"gitleaks"},
		Inputs:       nil, // rootfs directory will be determined by the CLI after mount.
		ScannersConfig: &common.ScannersConfig{
			Gitleaks: gitleaksconfig.Config{
				BinaryPath: gitleaksBinaryPath,
			},
		},
	}
}

func userSBOMConfigToFamiliesSbomConfig(sbomConfig *models.SBOMConfig) familiesSbom.Config {
	if sbomConfig == nil || sbomConfig.Enabled == nil || !*sbomConfig.Enabled {
		return familiesSbom.Config{}
	}
	return familiesSbom.Config{
		Enabled: true,
		// TODO(sambetts) This choice should come from the user's configuration
		AnalyzersList: []string{"syft", "trivy"},
		Inputs:        nil, // rootfs directory will be determined by the CLI after mount.
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

func userMisconfigurationConfigToFamiliesMisconfigurationConfig(misconfigurationConfig *models.MisconfigurationsConfig) misconfigurationTypes.Config {
	if misconfigurationConfig == nil || misconfigurationConfig.Enabled == nil || !*misconfigurationConfig.Enabled {
		return misconfigurationTypes.Config{}
	}
	return misconfigurationTypes.Config{
		Enabled: true,
		// TODO(sambetts) This choice should come from the user's configuration
		ScannersList:   []string{"fake"},
		Inputs:         nil, // rootfs directory will be determined by the CLI after mount.
		ScannersConfig: misconfigurationTypes.ScannersConfig{
			// TODO(sambetts) Add scanner configurations here as we add them like Lynis
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
		InputFromSbom: false, // will be determined by the CLI.
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

func userExploitsConfigToFamiliesExploitsConfig(exploitsConfig *models.ExploitsConfig, baseURL string) familiesExploits.Config {
	if exploitsConfig == nil || exploitsConfig.Enabled == nil || !*exploitsConfig.Enabled {
		return familiesExploits.Config{}
	}
	// TODO(erezf) Some choices should come from the user's configuration
	return familiesExploits.Config{
		Enabled:       true,
		ScannersList:  []string{"exploitdb"},
		InputFromVuln: true,
		ScannersConfig: &exploitsCommon.ScannersConfig{
			ExploitDB: exploitdbConfig.Config{
				BaseURL: baseURL,
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
	if job.Volume != nil {
		if err := job.Volume.Delete(ctx); err != nil {
			log.Errorf("Failed to delete volume. volumeID=%v: %v", job.Volume.GetID(), err)
		}
	}
}

// nolint:cyclop
func (s *Scanner) createInitTargetScanStatus(ctx context.Context, scanID, targetID string) (string, error) {
	initScanStatus := &models.TargetScanStatus{
		Exploits: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusExploitsStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Exploits),
		},
		General: &models.TargetScanState{
			Errors: nil,
			State:  stateToPointer(models.INIT),
		},
		Malware: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusMalwareStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Malware),
		},
		Misconfigurations: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusMisconfigurationsStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Misconfigurations),
		},
		Rootkits: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusRootkitsStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Rootkits),
		},
		Sbom: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusSbomStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Sbom),
		},
		Secrets: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusSecretsStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Secrets),
		},
		Vulnerabilities: &models.TargetScanState{
			Errors: nil,
			State:  getInitScanStatusVulnerabilitiesStateFromEnabled(s.scanConfig.ScanFamiliesConfig.Vulnerabilities),
		},
	}
	scanResult := models.TargetScanResult{
		Summary: createInitScanResultSummary(),
		Scan: &models.ScanRelationship{
			Id: scanID,
		},
		Status: initScanStatus,
		Target: &models.TargetRelationship{
			Id: targetID,
		},
	}
	createdScanResult, err := s.backendClient.PostScanResult(ctx, scanResult)
	if err != nil {
		var conErr backendclient.ScanResultConflictError
		if errors.As(err, &conErr) {
			log.Infof("Scan results already exist. scan result id=%v.", *conErr.ConflictingScanResult.Id)
			return *conErr.ConflictingScanResult.Id, nil
		}
		return "", fmt.Errorf("failed to post scan result: %v", err)
	}
	return *createdScanResult.Id, nil
}

func createInitScanResultSummary() *models.ScanFindingsSummary {
	return &models.ScanFindingsSummary{
		TotalExploits:          runtimeScanUtils.PointerTo[int](0),
		TotalMalware:           runtimeScanUtils.PointerTo[int](0),
		TotalMisconfigurations: runtimeScanUtils.PointerTo[int](0),
		TotalPackages:          runtimeScanUtils.PointerTo[int](0),
		TotalRootkits:          runtimeScanUtils.PointerTo[int](0),
		TotalSecrets:           runtimeScanUtils.PointerTo[int](0),
		TotalVulnerabilities: &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   runtimeScanUtils.PointerTo[int](0),
			TotalHighVulnerabilities:       runtimeScanUtils.PointerTo[int](0),
			TotalMediumVulnerabilities:     runtimeScanUtils.PointerTo[int](0),
			TotalLowVulnerabilities:        runtimeScanUtils.PointerTo[int](0),
			TotalNegligibleVulnerabilities: runtimeScanUtils.PointerTo[int](0),
		},
	}
}

func getInitScanStatusVulnerabilitiesStateFromEnabled(config *models.VulnerabilitiesConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func getInitScanStatusSecretsStateFromEnabled(config *models.SecretsConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func getInitScanStatusSbomStateFromEnabled(config *models.SBOMConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func getInitScanStatusRootkitsStateFromEnabled(config *models.RootkitsConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func getInitScanStatusMisconfigurationsStateFromEnabled(config *models.MisconfigurationsConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func getInitScanStatusMalwareStateFromEnabled(config *models.MalwareConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func getInitScanStatusExploitsStateFromEnabled(config *models.ExploitsConfig) *models.TargetScanStateState {
	if config == nil || config.Enabled == nil || !*config.Enabled {
		return stateToPointer(models.NOTSCANNED)
	}

	return stateToPointer(models.INIT)
}

func stateToPointer(state models.TargetScanStateState) *models.TargetScanStateState {
	return &state
}
