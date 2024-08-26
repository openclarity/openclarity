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

	"github.com/openclarity/openclarity/scanner"
	"github.com/openclarity/openclarity/scanner/families"

	apiclient "github.com/openclarity/openclarity/api/client"
	apitypes "github.com/openclarity/openclarity/api/types"
	"github.com/openclarity/openclarity/core/log"
	"github.com/openclarity/openclarity/core/to"
)

const (
	DefaultWaitForVolRetryInterval             = 15 * time.Second
	effectiveScanConfigAnnotationKey           = "openclarity.io/openclarity-scanner/config"
	deprecatedeffectiveScanConfigAnnotationKey = "openclarity.io/vmclarity-scanner/config"
)

type AssetScanID = apitypes.AssetScanID

type OpenClarityState struct {
	client *apiclient.Client

	assetScanID apitypes.AssetScanID
}

func NewOpenClarityState(client *apiclient.Client, id AssetScanID) (*OpenClarityState, error) {
	if client == nil {
		return nil, errors.New("API client must not be nil")
	}
	return &OpenClarityState{
		client:      client,
		assetScanID: id,
	}, nil
}

// nolint:cyclop
func (o *OpenClarityState) WaitForReadyState(ctx context.Context) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	timer := time.NewTicker(DefaultWaitForVolRetryInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			status, err := o.client.GetAssetScanStatus(ctx, o.assetScanID)
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

func (o *OpenClarityState) MarkInProgress(ctx context.Context, config *scanner.Config) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) MarkDone(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	assetScan.Stats.General.ScanTime.EndTime = to.Ptr(time.Now())
	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateDone,
		apitypes.AssetScanStatusReasonSuccess,
		nil,
	)

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) MarkFailed(ctx context.Context, errorMessage string) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan: %w", err)
	}

	assetScan.Stats.General.ScanTime.EndTime = to.Ptr(time.Now())
	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateFailed,
		apitypes.AssetScanStatusReasonError,
		to.Ptr(errorMessage),
	)

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) MarkFamilyScanInProgress(ctx context.Context, familyType families.FamilyType) error {
	var err error
	switch familyType {
	case families.SBOM:
		err = o.markSBOMScanInProgress(ctx)
	case families.Vulnerabilities:
		err = o.markVulnerabilitiesScanInProgress(ctx)
	case families.Secrets:
		err = o.markSecretsScanInProgress(ctx)
	case families.Exploits:
		err = o.markExploitsScanInProgress(ctx)
	case families.Misconfiguration:
		err = o.markMisconfigurationsScanInProgress(ctx)
	case families.Rootkits:
		err = o.markRootkitsScanInProgress(ctx)
	case families.Malware:
		err = o.markMalwareScanInProgress(ctx)
	case families.InfoFinder:
		err = o.markInfoFinderScanInProgress(ctx)
	case families.Plugins:
		err = o.markPluginsScanInProgress(ctx)
	}
	return err
}

func (o *OpenClarityState) markExploitsScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markSecretsScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markSBOMScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markVulnerabilitiesScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markInfoFinderScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markMalwareScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markMisconfigurationsScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markRootkitsScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) markPluginsScanInProgress(ctx context.Context) error {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{})
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

	err = o.client.PatchAssetScan(ctx, assetScan, o.assetScanID)
	if err != nil {
		return fmt.Errorf("failed to patch asset scan: %w", err)
	}

	return nil
}

func (o *OpenClarityState) IsAborted(ctx context.Context) (bool, error) {
	assetScan, err := o.client.GetAssetScan(ctx, o.assetScanID, apitypes.GetAssetScansAssetScanIDParams{
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

func appendEffectiveScanConfigAnnotation(annotations *apitypes.Annotations, config *scanner.Config) (*apitypes.Annotations, error) {
	var newAnnotations apitypes.Annotations
	if annotations != nil {
		// Add all annotations except the effective scan config one.
		for _, annotation := range *annotations {
			if *annotation.Key == effectiveScanConfigAnnotationKey {
				continue
			} else if *annotation.Key == deprecatedeffectiveScanConfigAnnotationKey {
				// change the key to the new one
				annotation.Key = to.Ptr(effectiveScanConfigAnnotationKey)
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
