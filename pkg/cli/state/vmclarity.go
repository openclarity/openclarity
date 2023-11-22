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

package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/families"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

const (
	DefaultWaitForVolRetryInterval   = 15 * time.Second
	effectiveScanConfigAnnotationKey = "openclarity.io/vmclarity-scanner/config"
)

type AssetScanID = models.AssetScanID

type VMClarityState struct {
	client *backendclient.BackendClient

	assetScanID models.AssetScanID
}

// nolint:cyclop
func (v *VMClarityState) WaitForReadyState(ctx context.Context) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	timer := time.NewTicker(DefaultWaitForVolRetryInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			status, err := v.client.GetAssetScanStatus(ctx, v.assetScanID)
			if err != nil {
				logger.Errorf("failed to get asset scan status: %v", err)
				break
			}

			if status == nil || status.General == nil || status.General.State == nil {
				return errors.New("invalid API response: status or status.general or status.general.state is nil")
			}

			switch *status.General.State {
			case models.AssetScanStateStatePending, models.AssetScanStateStateScheduled:
			case models.AssetScanStateStateAborted:
				// Do nothing as WaitForAborted is responsible for handling this case
			case models.AssetScanStateStateReadyToScan, models.AssetScanStateStateInProgress:
				return nil
			case models.AssetScanStateStateDone, models.AssetScanStateStateNotScanned:
				return fmt.Errorf("failed to wait for AssetScan become ready as it is in %s state", *status.General.State)
			}
		case <-ctx.Done():
			return fmt.Errorf("waiting for volume ready was canceled: %w", ctx.Err())
		}
	}
}

func (v *VMClarityState) MarkInProgress(ctx context.Context, config *families.Config) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.General == nil {
		assetScan.Status.General = &models.AssetScanState{}
	}
	if assetScan.Stats == nil {
		assetScan.Stats = &models.AssetScanStats{}
	}
	assetScan.Stats.General = &models.AssetScanGeneralStats{
		ScanTime: &models.AssetScanScanTime{
			StartTime: utils.PointerTo(time.Now()),
		},
	}

	state := models.AssetScanStateStateInProgress
	assetScan.Status.General.State = &state
	assetScan.Status.General.LastTransitionTime = utils.PointerTo(time.Now())

	assetScan.Annotations, err = appendEffectiveScanConfigAnnotation(assetScan.Annotations, config)
	if err != nil {
		return fmt.Errorf("failed to add effective scan config annotation: %w", err)
	}

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) MarkDone(ctx context.Context, errors []error) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Status == nil {
		assetScan.Status = &models.AssetScanStatus{}
	}
	if assetScan.Status.General == nil {
		assetScan.Status.General = &models.AssetScanState{}
	}

	assetScan.Stats.General.ScanTime.EndTime = utils.PointerTo(time.Now())

	state := models.AssetScanStateStateDone
	assetScan.Status.General.State = &state
	assetScan.Status.General.LastTransitionTime = utils.PointerTo(time.Now())

	// If we had any errors running the family or exporting results add it
	// to the general errors
	if len(errors) > 0 {
		var errorStrs []string
		// Pull the errors list out so that we can append to it (if there are
		// any errors at this point I would have hoped the orcestrator wouldn't
		// have spawned the VM) but we never know.
		if assetScan.Status.General.Errors != nil {
			errorStrs = *assetScan.Status.General.Errors
		}
		for _, err := range errors {
			if err != nil {
				errorStrs = append(errorStrs, err.Error())
			}
		}
		if len(errorStrs) > 0 {
			assetScan.Status.General.Errors = &errorStrs
		}
	}

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) MarkFamilyScanInProgress(ctx context.Context, familyType types.FamilyType) error {
	var err error
	switch familyType {
	case types.SBOM:
		err = v.markSBOMScanInProgress(ctx)
	case types.Vulnerabilities:
		err = v.markVulnerabilitiesScanInProgress(ctx)
	case types.Secrets:
		err = v.markSecretsScanInProgress(ctx)
	case types.Exploits:
		err = v.markExploitsScanInProgress(ctx)
	case types.Misconfiguration:
		err = v.markMisconfigurationsScanInProgress(ctx)
	case types.Rootkits:
		err = v.markRootkitsScanInProgress(ctx)
	case types.Malware:
		err = v.markMalwareScanInProgress(ctx)
	case types.InfoFinder:
		err = v.markInfoFinderScanInProgress(ctx)
	}
	return err
}

func (v *VMClarityState) markExploitsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Exploits == nil {
		assetScan.Exploits = &models.ExploitScan{}
	}

	assetScan.Exploits.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markSecretsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Secrets == nil {
		assetScan.Secrets = &models.SecretScan{}
	}

	assetScan.Secrets.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markSBOMScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Sbom == nil {
		assetScan.Sbom = &models.SbomScan{}
	}

	assetScan.Sbom.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markVulnerabilitiesScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Vulnerabilities == nil {
		assetScan.Vulnerabilities = &models.VulnerabilityScan{}
	}

	assetScan.Vulnerabilities.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markInfoFinderScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.InfoFinder == nil {
		assetScan.InfoFinder = &models.InfoFinderScan{}
	}

	assetScan.InfoFinder.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markMalwareScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Malware == nil {
		assetScan.Malware = &models.MalwareScan{}
	}

	assetScan.Malware.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markMisconfigurationsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Misconfigurations == nil {
		assetScan.Misconfigurations = &models.MisconfigurationScan{}
	}

	assetScan.Misconfigurations.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markRootkitsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Rootkits == nil {
		assetScan.Rootkits = &models.RootkitScan{}
	}

	assetScan.Rootkits.Status = models.NewScannerStatus(
		models.ScannerStatusStateInProgress,
		models.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) IsAborted(ctx context.Context) (bool, error) {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, models.GetAssetScansAssetScanIDParams{
		Select: utils.PointerTo("id,status"),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get asset scan: %w", err)
	}

	state, ok := assetScan.GetGeneralState()
	if !ok {
		return false, errors.New("failed to get general state of asset scan")
	}

	if state == models.AssetScanStateStateAborted {
		return true, nil
	}

	return false, nil
}

func NewVMClarityState(client *backendclient.BackendClient, id AssetScanID) (*VMClarityState, error) {
	if client == nil {
		return nil, errors.New("backend client must not be nil")
	}
	return &VMClarityState{
		client:      client,
		assetScanID: id,
	}, nil
}

func appendEffectiveScanConfigAnnotation(annotations *models.Annotations, config *families.Config) (*models.Annotations, error) {
	var newAnnotations models.Annotations
	if annotations != nil {
		// Add all annotations except the effective scan config one.
		for _, annotation := range *annotations {
			if *annotation.Key == effectiveScanConfigAnnotationKey {
				continue
			}
			newAnnotations = append(newAnnotations, annotation)
		}
	}
	// Add the new effective scan config annotation
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal effective families config: %w", err)
	}
	newAnnotations = append(newAnnotations, models.Annotations{
		{
			Key:   utils.PointerTo(effectiveScanConfigAnnotationKey),
			Value: utils.PointerTo(string(configJSON)),
		},
	}...)

	return &newAnnotations, nil
}
