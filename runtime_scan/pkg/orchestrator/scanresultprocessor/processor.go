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

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/common"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
)

type ScanResultProcessor struct {
	client           *backendclient.BackendClient
	pollPeriod       time.Duration
	reconcileTimeout time.Duration
}

func New(config Config) *ScanResultProcessor {
	return &ScanResultProcessor{
		client:           config.Backend,
		pollPeriod:       config.PollPeriod,
		reconcileTimeout: config.ReconcileTimeout,
	}
}

// Returns true if TargetScanStatus.State is DONE and there are no Errors.
func statusCompletedWithNoErrors(tss *models.TargetScanState) bool {
	return tss != nil && tss.State != nil && *tss.State == models.TargetScanStateStateDONE && (tss.Errors == nil || len(*tss.Errors) == 0)
}

// nolint:cyclop
func (srp *ScanResultProcessor) Reconcile(ctx context.Context, event ScanResultReconcileEvent) error {
	// Get latest information, in case we've been sat in the reconcile
	// queue for a while
	scanResult, err := srp.client.GetScanResult(ctx, event.ScanResultID, models.GetScanResultsScanResultIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get scan result from API: %w", err)
	}

	// Re-check the findingsProcessed boolean, we might have been re-queued
	// while already being reconciled, if so we can short circuit here.
	if scanResult.FindingsProcessed != nil && *scanResult.FindingsProcessed {
		return nil
	}

	newFailedToReconcileTypeError := func(err error, t string) error {
		return fmt.Errorf("failed to reconcile scan result %s %s to findings: %w", *scanResult.Id, t, err)
	}

	// Process each of the successfully scanned (state DONE and no errors) families into findings.
	if statusCompletedWithNoErrors(scanResult.Status.Vulnerabilities) {
		if err := srp.reconcileResultVulnerabilitiesToFindings(ctx, scanResult); err != nil {
			return newFailedToReconcileTypeError(err, "vulnerabilities")
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

	// Mark post-processing completed for this scan result
	scanResult.FindingsProcessed = utils.PointerTo(true)
	err = srp.client.PatchScanResult(ctx, scanResult, *scanResult.Id)
	if err != nil {
		return fmt.Errorf("failed to update scan result %s: %w", *scanResult.Id, err)
	}

	return nil
}

type ScanResultReconcileEvent struct {
	ScanResultID string
}

func (srp *ScanResultProcessor) GetItems(ctx context.Context) ([]ScanResultReconcileEvent, error) {
	scanResults, err := srp.client.GetScanResults(ctx, models.GetScanResultsParams{
		Filter: utils.PointerTo("status/general/state eq 'DONE' and (findingsProcessed eq false or findingsProcessed eq null)"),
		Select: utils.PointerTo("id"),
	})
	if err != nil {
		return []ScanResultReconcileEvent{}, fmt.Errorf("failed to get scan results from API: %w", err)
	}

	items := make([]ScanResultReconcileEvent, len(*scanResults.Items))
	for i, res := range *scanResults.Items {
		items[i] = ScanResultReconcileEvent{*res.Id}
	}

	return items, nil
}

func (srp *ScanResultProcessor) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "ScanResultProcessor")
	ctx = log.SetLoggerForContext(ctx, logger)

	queue := common.NewQueue[ScanResultReconcileEvent]()

	poller := common.Poller[ScanResultReconcileEvent]{
		PollPeriod: srp.pollPeriod,
		GetItems:   srp.GetItems,
		Queue:      queue,
	}
	poller.Start(ctx)

	reconciler := common.Reconciler[ScanResultReconcileEvent]{
		ReconcileFunction: srp.Reconcile,
		ReconcileTimeout:  srp.reconcileTimeout,
		Queue:             queue,
	}
	reconciler.Start(ctx)
}
