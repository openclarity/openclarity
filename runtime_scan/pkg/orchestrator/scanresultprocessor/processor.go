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

package scanresultprocessor

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/common"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

type ScanResultProcessor struct {
	logger *log.Entry
	client *backendclient.BackendClient
}

func NewScanResultProcessor(client *backendclient.BackendClient) *ScanResultProcessor {
	logger := log.WithFields(log.Fields{"controller": "ScanResultProcessor"})

	return &ScanResultProcessor{
		logger: logger,
		client: client,
	}
}

// Returns true if TargetScanStatus.State is DONE and there are no Errors.
func statusCompletedWithNoErrors(tss *models.TargetScanState) bool {
	return tss != nil && tss.State != nil && *tss.State == models.DONE && (tss.Errors == nil || len(*tss.Errors) == 0)
}

// nolint:cyclop
func (srp *ScanResultProcessor) Reconcile(ctx context.Context, scanResult models.TargetScanResult) error {
	newFailedToReconcileTypeError := func(err error, t string) error {
		return fmt.Errorf("failed to reconcile scan result %s %s to findings: %w", *scanResult.Id, t, err)
	}

	// Get latest information, in case we've been sat in the reconcile
	// queue for a while
	scanResult, err := srp.client.GetScanResult(ctx, *scanResult.Id, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result from API: %w", err)
	}

	// Re-check the findingsProcessed boolean, we might have been re-queued
	// while already being reconciled, if so we can short circuit here.
	if scanResult.FindingsProcessed != nil && *scanResult.FindingsProcessed {
		return nil
	}

	// Process each of the successfully scanned (state DONE and no errors) families into findings.
	if statusCompletedWithNoErrors(scanResult.Status.Vulnerabilities) {
		if err := srp.reconcileResultVulnerabilitiesToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "vulnerabilties")
		}
	}

	if statusCompletedWithNoErrors(scanResult.Status.Sbom) {
		if err := srp.reconcileResultPackagesToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "sbom")
		}
	}

	if statusCompletedWithNoErrors(scanResult.Status.Exploits) {
		if err := srp.reconcileResultExploitsToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "exploits")
		}
	}

	if statusCompletedWithNoErrors(scanResult.Status.Secrets) {
		if err := srp.reconcileResultSecretsToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "secrets")
		}
	}

	if statusCompletedWithNoErrors(scanResult.Status.Malware) {
		if err := srp.reconcileResultMalwareToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "malware")
		}
	}

	if statusCompletedWithNoErrors(scanResult.Status.Rootkits) {
		if err := srp.reconcileResultRootkitsToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "rootkits")
		}
	}

	if statusCompletedWithNoErrors(scanResult.Status.Misconfigurations) {
		if err := srp.reconcileResultMisconfigurationsToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "misconfigurations")
		}
	}

	// Mark post processing completed for this scan result
	scanResult.FindingsProcessed = utils.PointerTo(true)
	err = srp.client.PatchScanResult(ctx, scanResult, *scanResult.Id)
	if err != nil {
		return fmt.Errorf("failed to update scan result %s: %w", *scanResult.Id, err)
	}

	return nil
}

func (srp *ScanResultProcessor) GetItems(ctx context.Context) ([]models.TargetScanResult, error) {
	scanResults, err := srp.client.GetScanResults(ctx, models.GetScanResultsParams{
		Filter: utils.PointerTo("status/general/state eq 'DONE' and (findingsProcessed eq false or findingsProcessed eq null)"),
	})
	if err != nil {
		return []models.TargetScanResult{}, fmt.Errorf("failed to get scan results from API: %w", err)
	}
	return *scanResults.Items, nil
}

const (
	reconcileTimeoutSeconds = 120
	pollPeriodSeconds       = 60
)

func (srp *ScanResultProcessor) Start(ctx context.Context) {
	poller := common.Poller[models.TargetScanResult]{
		Logger:     srp.logger,
		PollPeriod: pollPeriodSeconds * time.Second,
		GetItems:   srp.GetItems,
	}
	eventChan := poller.Start(ctx)

	reconciler := common.Reconciler[models.TargetScanResult]{
		Logger:            srp.logger,
		EventChan:         eventChan,
		ReconcileFunction: srp.Reconcile,
		ReconcileTimeout:  reconcileTimeoutSeconds * time.Second,
	}
	reconciler.Start(ctx)
}
