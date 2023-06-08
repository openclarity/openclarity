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
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/families/types"
	"github.com/openclarity/vmclarity/shared/pkg/log"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

const (
	DefaultWaitForVolRetryInterval = 15 * time.Second
)

type ScanResultID = models.ScanResultID

type VMClarityState struct {
	client *backendclient.BackendClient

	scanResultID models.ScanResultID
}

// nolint:cyclop
func (v *VMClarityState) WaitForReadyState(ctx context.Context) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	timer := time.NewTicker(DefaultWaitForVolRetryInterval)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			status, err := v.client.GetScanResultStatus(ctx, v.scanResultID)
			if err != nil {
				logger.Errorf("failed to get scan result status: %v", err)
				break
			}

			if status == nil || status.General == nil || status.General.State == nil {
				return errors.New("invalid API response: status or status.general or status.general.state is nil")
			}

			switch *status.General.State {
			case models.TargetScanStateStatePending, models.TargetScanStateStateScheduled:
			case models.TargetScanStateStateAborted:
				// Do nothing as WaitForAborted is responsible for handling this case
			case models.TargetScanStateStateReadyToScan, models.TargetScanStateStateInProgress:
				return nil
			case models.TargetScanStateStateDone, models.TargetScanStateStateNotScanned:
				return fmt.Errorf("failed to wait for ScanResult become ready as it is in %s state", *status.General.State)
			}
		case <-ctx.Done():
			return fmt.Errorf("waiting for volume ready was canceled: %w", ctx.Err())
		}
	}
}

func (v *VMClarityState) MarkInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.General == nil {
		scanResult.Status.General = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.General.State = &state
	scanResult.Status.General.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) MarkDone(ctx context.Context, errors []error) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.General == nil {
		scanResult.Status.General = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateDone
	scanResult.Status.General.State = &state
	scanResult.Status.General.LastTransitionTime = utils.PointerTo(time.Now())

	// If we had any errors running the family or exporting results add it
	// to the general errors
	if len(errors) > 0 {
		var errorStrs []string
		// Pull the errors list out so that we can append to it (if there are
		// any errors at this point I would have hoped the orcestrator wouldn't
		// have spawned the VM) but we never know.
		if scanResult.Status.General.Errors != nil {
			errorStrs = *scanResult.Status.General.Errors
		}
		for _, err := range errors {
			if err != nil {
				errorStrs = append(errorStrs, err.Error())
			}
		}
		if len(errorStrs) > 0 {
			scanResult.Status.General.Errors = &errorStrs
		}
	}

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
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
	}
	return err
}

func (v *VMClarityState) markExploitsScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Exploits == nil {
		scanResult.Status.Exploits = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Exploits.State = &state
	scanResult.Status.Exploits.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) markSecretsScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Secrets == nil {
		scanResult.Status.Secrets = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Secrets.State = &state
	scanResult.Status.Secrets.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) markSBOMScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Sbom == nil {
		scanResult.Status.Sbom = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Sbom.State = &state
	scanResult.Status.Sbom.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) markVulnerabilitiesScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Vulnerabilities == nil {
		scanResult.Status.Vulnerabilities = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Vulnerabilities.State = &state
	scanResult.Status.Vulnerabilities.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) markMalwareScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Malware == nil {
		scanResult.Status.Malware = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Malware.State = &state
	scanResult.Status.Malware.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) markMisconfigurationsScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Misconfigurations == nil {
		scanResult.Status.Misconfigurations = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Misconfigurations.State = &state
	scanResult.Status.Misconfigurations.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) markRootkitsScanInProgress(ctx context.Context) error {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status == nil {
		scanResult.Status = &models.TargetScanStatus{}
	}
	if scanResult.Status.Rootkits == nil {
		scanResult.Status.Rootkits = &models.TargetScanState{}
	}

	state := models.TargetScanStateStateInProgress
	scanResult.Status.Rootkits.State = &state
	scanResult.Status.Rootkits.LastTransitionTime = utils.PointerTo(time.Now())

	err = v.client.PatchScanResult(ctx, scanResult, v.scanResultID)
	if err != nil {
		return fmt.Errorf("failed to patch scan result: %w", err)
	}

	return nil
}

func (v *VMClarityState) IsAborted(ctx context.Context) (bool, error) {
	scanResult, err := v.client.GetScanResult(ctx, v.scanResultID, models.GetScanResultsScanResultIDParams{
		Select: utils.PointerTo("id,status"),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get scan result: %w", err)
	}

	state, ok := scanResult.GetGeneralState()
	if !ok {
		return false, errors.New("failed to get general state of scan result")
	}

	if state == models.TargetScanStateStateAborted {
		return true, nil
	}

	return false, nil
}

func NewVMClarityState(client *backendclient.BackendClient, id ScanResultID) (*VMClarityState, error) {
	if client == nil {
		return nil, errors.New("backend client must not be nil")
	}
	return &VMClarityState{
		client:       client,
		scanResultID: id,
	}, nil
}
