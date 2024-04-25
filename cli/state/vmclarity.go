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

	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/types"
)

const (
	DefaultWaitForVolRetryInterval   = 15 * time.Second
	effectiveScanConfigAnnotationKey = "openclarity.io/vmclarity-scanner/config"
)

type AssetScanID = apitypes.AssetScanID

type VMClarityState struct {
	client *apiclient.Client

	assetScanID apitypes.AssetScanID
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

			if status == nil {
				return errors.New("invalid API response: status is nil")
			}

			switch status.State {
			case apitypes.AssetScanStatusStatePending, apitypes.AssetScanStatusStateScheduled:
			case apitypes.AssetScanStatusStateAborted:
				// Do nothing as WaitForAborted is responsible for handling this case
			case apitypes.AssetScanStatusStateReadyToScan, apitypes.AssetScanStatusStateInProgress:
				return nil
			case apitypes.AssetScanStatusStateDone, apitypes.AssetScanStatusStateFailed:
				return fmt.Errorf("failed to wait for AssetScan become ready as it is in %s state", status.State)
			}
		case <-ctx.Done():
			return fmt.Errorf("waiting for volume ready was canceled: %w", ctx.Err())
		}
	}
}

func (v *VMClarityState) MarkInProgress(ctx context.Context, config *families.Config) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Stats == nil {
		assetScan.Stats = &apitypes.AssetScanStats{}
	}
	assetScan.Stats.General = &apitypes.AssetScanGeneralStats{
		ScanTime: &apitypes.AssetScanScanTime{
			StartTime: to.Ptr(time.Now()),
		},
	}

	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateInProgress,
		apitypes.AssetScanStatusReasonScannerIsRunning,
		nil,
	)

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

func (v *VMClarityState) MarkDone(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	assetScan.Stats.General.ScanTime.EndTime = to.Ptr(time.Now())
	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateDone,
		apitypes.AssetScanStatusReasonSuccess,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) MarkFailed(ctx context.Context, errorMessage string) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	assetScan.Stats.General.ScanTime.EndTime = to.Ptr(time.Now())
	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateFailed,
		apitypes.AssetScanStatusReasonError,
		to.Ptr(errorMessage),
	)

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
	case types.Plugins:
		err = v.markPluginsScanInProgress(ctx)
	}
	return err
}

func (v *VMClarityState) markExploitsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Exploits == nil {
		assetScan.Exploits = &apitypes.ExploitScan{}
	}

	assetScan.Exploits.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markSecretsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Secrets == nil {
		assetScan.Secrets = &apitypes.SecretScan{}
	}

	assetScan.Secrets.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markSBOMScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Sbom == nil {
		assetScan.Sbom = &apitypes.SbomScan{}
	}

	assetScan.Sbom.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markVulnerabilitiesScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Vulnerabilities == nil {
		assetScan.Vulnerabilities = &apitypes.VulnerabilityScan{}
	}

	assetScan.Vulnerabilities.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markInfoFinderScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.InfoFinder == nil {
		assetScan.InfoFinder = &apitypes.InfoFinderScan{}
	}

	assetScan.InfoFinder.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markMalwareScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Malware == nil {
		assetScan.Malware = &apitypes.MalwareScan{}
	}

	assetScan.Malware.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markMisconfigurationsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Misconfigurations == nil {
		assetScan.Misconfigurations = &apitypes.MisconfigurationScan{}
	}

	assetScan.Misconfigurations.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markRootkitsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Rootkits == nil {
		assetScan.Rootkits = &apitypes.RootkitScan{}
	}

	assetScan.Rootkits.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) markPluginsScanInProgress(ctx context.Context) error {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	if assetScan.Plugins == nil {
		assetScan.Plugins = &apitypes.PluginScan{}
	}

	assetScan.Plugins.Status = apitypes.NewScannerStatus(
		apitypes.ScannerStatusStateInProgress,
		apitypes.ScannerStatusReasonScanning,
		nil,
	)

	err = v.client.PatchAssetScan(ctx, assetScan, v.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (v *VMClarityState) IsAborted(ctx context.Context) (bool, error) {
	assetScan, err := v.client.GetAssetScan(ctx, v.assetScanID, apitypes.GetAssetScansAssetScanIDParams{
		Select: to.Ptr("id,status"),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get asset scan: %w", err)
	}

	status, ok := assetScan.GetStatus()
	if !ok {
		return false, errors.New("failed to get status of asset scan")
	}

	if status.State == apitypes.AssetScanStatusStateAborted {
		return true, nil
	}

	return false, nil
}

func NewVMClarityState(client *apiclient.Client, id AssetScanID) (*VMClarityState, error) {
	if client == nil {
		return nil, errors.New("API client must not be nil")
	}
	return &VMClarityState{
		client:      client,
		assetScanID: id,
	}, nil
}

func appendEffectiveScanConfigAnnotation(annotations *apitypes.Annotations, config *families.Config) (*apitypes.Annotations, error) {
	var newAnnotations apitypes.Annotations
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
	newAnnotations = append(newAnnotations, apitypes.Annotations{
		{
			Key:   to.Ptr(effectiveScanConfigAnnotationKey),
			Value: to.Ptr(string(configJSON)),
		},
	}...)

	return &newAnnotations, nil
}
