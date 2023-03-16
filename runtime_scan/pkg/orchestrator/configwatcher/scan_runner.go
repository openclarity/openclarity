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
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	_scanner "github.com/openclarity/vmclarity/runtime_scan/pkg/scanner"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

func (scw *ScanConfigWatcher) scan(ctx context.Context, scanConfig *models.ScanConfig) error {
	// TODO: check if existing scan or a new scan
	targetInstances, scanID, err := scw.initNewScan(ctx, scanConfig)
	if err != nil {
		return fmt.Errorf("failed to init new scan: %v", err)
	}

	scanner := _scanner.CreateScanner(scw.scannerConfig, scw.providerClient, scw.backendClient, scanConfig, targetInstances, scanID)
	if err := scanner.Scan(ctx); err != nil {
		return fmt.Errorf("failed to scan: %v", err)
	}

	return nil
}

// initNewScan Initialized a new scan, returns target instances and scan ID.
func (scw *ScanConfigWatcher) initNewScan(ctx context.Context, scanConfig *models.ScanConfig) ([]*types.TargetInstance, string, error) {
	// Create scan in pending
	now := time.Now().UTC()
	scan := &models.Scan{
		ScanConfig: &models.ScanConfigRelationship{
			Id: *scanConfig.Id,
		},
		ScanConfigSnapshot: &models.ScanConfigData{
			MaxParallelScanners: scanConfig.MaxParallelScanners,
			Name:                scanConfig.Name,
			ScanFamiliesConfig:  scanConfig.ScanFamiliesConfig,
			Scheduled:           scanConfig.Scheduled,
			Scope:               scanConfig.Scope,
		},
		StartTime: &now,
		State:     utils.PointerTo(models.ScanStatePending),
		Summary:   createInitScanSummary(),
	}
	var scanID string
	createdScan, err := scw.backendClient.PostScan(ctx, *scan)
	if err != nil {
		var conErr backendclient.ScanConflictError
		if errors.As(err, &conErr) {
			log.Infof("Scan already exist. scan id=%v.", *conErr.ConflictingScan.Id)
			scanID = *conErr.ConflictingScan.Id
		} else {
			return nil, "", fmt.Errorf("failed to post scan: %v", err)
		}
	} else {
		scanID = *createdScan.Id
	}

	// Do discovery of targets
	instances, err := scw.providerClient.DiscoverInstances(ctx, scan.ScanConfigSnapshot.Scope)
	if err != nil {
		return nil, "", fmt.Errorf("failed to discover instances to scan: %v", err)
	}
	targetInstances, err := scw.createTargetInstances(ctx, instances)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get or create targets: %v", err)
	}

	// Move scan to discovered and add the discovered targets.
	targetIds := getTargetIDs(targetInstances)
	scan = &models.Scan{
		TargetIDs:    targetIds,
		State:        utils.PointerTo(models.ScanStateDiscovered),
		StateMessage: utils.PointerTo("Targets for scan successfully discovered"),
	}
	err = scw.backendClient.PatchScan(ctx, scanID, scan)
	if err != nil {
		return nil, "", fmt.Errorf("failed to update scan: %v", err)
	}

	return targetInstances, scanID, nil
}

func createInitScanSummary() *models.ScanSummary {
	return &models.ScanSummary{
		JobsCompleted:          utils.PointerTo(0),
		JobsLeftToRun:          utils.PointerTo(0),
		TotalExploits:          utils.PointerTo(0),
		TotalMalware:           utils.PointerTo(0),
		TotalMisconfigurations: utils.PointerTo(0),
		TotalPackages:          utils.PointerTo(0),
		TotalRootkits:          utils.PointerTo(0),
		TotalSecrets:           utils.PointerTo(0),
		TotalVulnerabilities: &models.VulnerabilityScanSummary{
			TotalCriticalVulnerabilities:   utils.PointerTo(0),
			TotalHighVulnerabilities:       utils.PointerTo(0),
			TotalLowVulnerabilities:        utils.PointerTo(0),
			TotalMediumVulnerabilities:     utils.PointerTo(0),
			TotalNegligibleVulnerabilities: utils.PointerTo(0),
		},
	}
}

func getTargetIDs(targetInstances []*types.TargetInstance) *[]string {
	ret := make([]string, len(targetInstances))
	for i, targetInstance := range targetInstances {
		ret[i] = targetInstance.TargetID
	}

	return &ret
}

func (scw *ScanConfigWatcher) createTargetInstances(ctx context.Context, instances []types.Instance) ([]*types.TargetInstance, error) {
	targetInstances := make([]*types.TargetInstance, 0, len(instances))
	for i, instance := range instances {
		targetID, err := scw.createTarget(ctx, instance)
		if err != nil {
			return nil, fmt.Errorf("failed to create target. instanceID=%v: %v", instance.GetID(), err)
		}
		targetInstances = append(targetInstances, &types.TargetInstance{
			TargetID: targetID,
			Instance: instances[i],
		})
	}

	return targetInstances, nil
}

func (scw *ScanConfigWatcher) createTarget(ctx context.Context, instance types.Instance) (string, error) {
	info := models.TargetType{}
	instanceProvider := models.AWS
	err := info.FromVMInfo(models.VMInfo{
		InstanceID:       instance.GetID(),
		InstanceProvider: &instanceProvider,
		Location:         instance.GetLocation(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create VMInfo: %v", err)
	}
	createdTarget, err := scw.backendClient.PostTarget(ctx, models.Target{
		TargetInfo: &info,
	})
	if err != nil {
		var conErr backendclient.TargetConflictError
		if errors.As(err, &conErr) {
			log.Infof("Target already exist. target id=%v.", *conErr.ConflictingTarget.Id)
			return *conErr.ConflictingTarget.Id, nil
		}
		return "", fmt.Errorf("failed to post target: %v", err)
	}
	return *createdTarget.Id, nil
}
