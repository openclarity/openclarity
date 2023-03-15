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

package configwatcher

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	_config "github.com/openclarity/vmclarity/runtime_scan/pkg/config"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

const (
	timeWindow = 5 * time.Minute
)

type ScanConfigWatcher struct {
	backendClient  *backendclient.BackendClient
	providerClient provider.Client
	scannerConfig  *_config.ScannerConfig
}

func CreateScanConfigWatcher(
	backendClient *backendclient.BackendClient,
	providerClient provider.Client,
	scannerConfig _config.ScannerConfig,
) *ScanConfigWatcher {
	return &ScanConfigWatcher{
		backendClient:  backendClient,
		providerClient: providerClient,
		scannerConfig:  &scannerConfig,
	}
}

func (scw *ScanConfigWatcher) hasRunningScansByScanConfigIDAndOperationTime(scanConfigID string, operationTime time.Time) (bool, error) {
	// TODO(sambetts) Once we can validate that gte/eq works with times
	// correctly then we can add them to the filter like:
	//
	//	scanConfig/id eq '%s' and (endTime eq null or startTime gte '%s')
	//
	filter := fmt.Sprintf("scanConfig/id eq '%s'", scanConfigID)
	scans, err := scw.backendClient.GetScans(context.TODO(), models.GetScansParams{
		Filter: &filter,
	})
	if err != nil {
		return false, fmt.Errorf("failed to get a scans: %v", err)
	}

	return anyScansRunningOrCompleted(scans, operationTime), nil
}

func (scw *ScanConfigWatcher) getScanConfigsToScan() ([]models.ScanConfig, error) {
	scanConfigsToScan := make([]models.ScanConfig, 0)
	scanConfigs, err := scw.backendClient.GetScanConfigs(context.TODO(), models.GetScanConfigsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to check new scan configs: %v", err)
	}

	log.Infof("Found %d ScanConfigs from the backend", len(*scanConfigs.Items))

	now := time.Now()
	for _, scanConfig := range *scanConfigs.Items {
		// Check only the SingleScheduledScanConfigs at the moment
		shouldScan := false
		scheduled, err := scanConfig.Scheduled.ValueByDiscriminator()
		if err != nil {
			log.Errorf("failed to determine scheduled scan config type: %v", err)
			continue
		}
		switch singleSchedule := scheduled.(type) {
		case models.SingleScheduleScanConfig:
			shouldScan, err = scw.shouldStartSingleScheduleScanConfig(*scanConfig.Id, singleSchedule, now)
			if err != nil {
				log.Errorf("Failed to get scans for scan config ID=%s: %v", *scanConfig.Id, err)
				continue
			}
		default:
		}

		if shouldScan {
			log.Debugf("ScanConfig %s should start to scan", *scanConfig.Id)
			scanConfigsToScan = append(scanConfigsToScan, scanConfig)
		} else {
			log.Debugf("ScanConfig %s should not start to scan", *scanConfig.Id)
		}
	}
	return scanConfigsToScan, nil
}

func anyScansRunningOrCompleted(scans *models.Scans, operationTime time.Time) bool {
	if scans.Items == nil {
		return false
	}
	for _, scan := range *scans.Items {
		if scan.EndTime == nil {
			// There is a running scan for this scanConfig
			return true
		}

		// If StartTime isn't set on a Scan then it is assumed that its
		// not started, this should never happen.
		if scan.StartTime != nil && scan.StartTime.After(operationTime) {
			// There is a completed scan that started after the operation time
			return true
		}
	}
	return false
}

// isWithinTheWindow checks if `checkTime` is within the window (after `now` and before `now + window`).
func isWithinTheWindow(checkTime, now time.Time, window time.Duration) bool {
	if checkTime.Before(now) {
		return false
	}

	endWindowTime := now.Add(window)
	return checkTime.Before(endWindowTime)
}

func (scw *ScanConfigWatcher) shouldStartSingleScheduleScanConfig(scanConfigID string, schedule models.SingleScheduleScanConfig, now time.Time) (bool, error) {
	// Skip processing ScanConfig because its operationTime is not within the start window
	if !isWithinTheWindow(schedule.OperationTime, now, timeWindow) {
		log.Debugf("ScanConfig %s start time %v outside of the start window %v - %v", scanConfigID, schedule.OperationTime.Format(time.RFC3339), now.Format(time.RFC3339), now.Add(timeWindow).Format(time.RFC3339))
		return false, nil
	}
	// Check running or completed scan for specific scan config
	hasRunningOrCompletedScan, err := scw.hasRunningScansByScanConfigIDAndOperationTime(scanConfigID, schedule.OperationTime)
	if err != nil {
		return false, fmt.Errorf("failed to get scans: %v", err)
	}
	if hasRunningOrCompletedScan {
		log.Debugf("ScanConfig %s has a running or completed scan", scanConfigID)
	}
	return !hasRunningOrCompletedScan, nil
}

func (scw *ScanConfigWatcher) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-time.After(scw.scannerConfig.ScanConfigWatchInterval):
				log.Debug("Looking for ScanConfigs to Scan")
				// nolint:contextcheck
				scanConfigsToScan, err := scw.getScanConfigsToScan()
				if err != nil {
					log.Warnf("Failed to get scan configs to scan: %v", err)
					break
				}
				log.Debugf("Found %d ScanConfigs that need to start scanning.", len(scanConfigsToScan))
				if len(scanConfigsToScan) > 0 {
					scw.runNewScans(ctx, scanConfigsToScan)
				}
			case <-ctx.Done():
				log.Infof("Stop watching scan configs.")
				return
			}
		}
	}()
}
