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

package scanresultwatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/common"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type (
	ScanResultQueue      = common.Queue[ScanResultReconcileEvent]
	ScanResultPoller     = common.Poller[ScanResultReconcileEvent]
	ScanResultReconciler = common.Reconciler[ScanResultReconcileEvent]
)

func New(c Config) *Watcher {
	return &Watcher{
		backend:          c.Backend,
		provider:         c.Provider,
		scannerConfig:    c.ScannerConfig,
		pollPeriod:       c.PollPeriod,
		reconcileTimeout: c.ReconcileTimeout,
		queue:            common.NewQueue[ScanResultReconcileEvent](),
	}
}

type Watcher struct {
	backend          *backendclient.BackendClient
	provider         provider.Provider
	scannerConfig    ScannerConfig
	pollPeriod       time.Duration
	reconcileTimeout time.Duration

	queue *ScanResultQueue
}

func (w *Watcher) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "ScanResultWatcher")
	ctx = log.SetLoggerForContext(ctx, logger)

	poller := &ScanResultPoller{
		PollPeriod: w.pollPeriod,
		Queue:      w.queue,
		GetItems:   w.GetScanResults,
	}
	poller.Start(ctx)

	reconciler := &ScanResultReconciler{
		ReconcileTimeout:  w.reconcileTimeout,
		Queue:             w.queue,
		ReconcileFunction: w.Reconcile,
	}
	reconciler.Start(ctx)
}

// nolint:cyclop
func (w *Watcher) GetScanResults(ctx context.Context) ([]ScanResultReconcileEvent, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Debugf("Fetching ScanResults which need to be reconciled")

	filter := fmt.Sprintf("status/general/state ne '%s' or status/general/state ne '%s' or resourceCleanup eq '%s'",
		models.TargetScanStateStateDONE, models.TargetScanStateStateNOTSCANNED, models.ResourceCleanupStatePENDING)
	selector := "id,scan/id,target/id"
	params := models.GetScanResultsParams{
		Filter: &filter,
		Select: &selector,
		Count:  utils.PointerTo(true),
	}
	scanResults, err := w.backend.GetScanResults(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get ScanResults: %w", err)
	}

	switch {
	case scanResults.Items == nil && scanResults.Count == nil:
		return nil, fmt.Errorf("failed to fetch ScanResults: invalid API response: %v", scanResults)
	case scanResults.Count != nil && *scanResults.Count <= 0:
		fallthrough
	case scanResults.Items != nil && len(*scanResults.Items) <= 0:
		return nil, nil
	}

	events := make([]ScanResultReconcileEvent, 0, len(*scanResults.Items))
	for _, scanResult := range *scanResults.Items {
		scanResultID, ok := scanResult.GetID()
		if !ok {
			logger.Warnf("Skipping to invalid ScanResult: ID is nil: %v", scanResult)
			continue
		}
		scanID, ok := scanResult.GetScanID()
		if !ok {
			logger.Warnf("Skipping to invalid ScanResult: Scan.ID is nil: %v", scanResult)
			continue
		}
		targetID, ok := scanResult.GetTargetID()
		if !ok {
			logger.Warnf("Skipping to invalid ScanResult: Target.ID is nil: %v", scanResult)
			continue
		}

		events = append(events, ScanResultReconcileEvent{
			ScanResultID: scanResultID,
			ScanID:       scanID,
			TargetID:     targetID,
		})
	}

	return events, nil
}

// nolint:cyclop
func (w *Watcher) Reconcile(ctx context.Context, event ScanResultReconcileEvent) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(event.ToFields())
	ctx = log.SetLoggerForContext(ctx, logger)

	logger.Infof("Reconciling ScanResult event")

	scanResult, err := w.backend.GetScanResult(ctx, event.ScanResultID, models.GetScanResultsScanResultIDParams{
		Expand: utils.PointerTo("scan,target"),
	})
	if err != nil {
		return fmt.Errorf("failed to get ScanResult with %s id: %w", event.ScanResultID, err)
	}

	state, ok := scanResult.GetGeneralState()
	if !ok {
		return fmt.Errorf("cannot determine state of ScanResult with %s id", event.ScanResultID)
	}

	logger.Tracef("Reconciling ScanResult state: %s", state)

	switch state {
	case models.TargetScanStateStateINIT:
		if err = w.reconcileInit(ctx, &scanResult); err != nil {
			return err
		}
	case models.TargetScanStateStateATTACHED, models.TargetScanStateStateINPROGRESS:
		// TODO(chrisgacsal): make sure that TargetScanResult state is set to ABORTED state once the TargetScanResult
		//                    schema is extended with timeout field and the deadline is missed.
		break
	case models.TargetScanStateStateABORTED, models.TargetScanStateStateNOTSCANNED:
		break
	case models.TargetScanStateStateDONE:
		if err = w.reconcileDone(ctx, &scanResult); err != nil {
			return err
		}
	default:
	}

	return nil
}

// nolint:cyclop
func (w *Watcher) reconcileInit(ctx context.Context, scanResult *models.TargetScanResult) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	scanResultID, ok := scanResult.GetID()
	if !ok {
		return errors.New("invalid ScanResult: ID is nil")
	}

	if scanResult.Scan == nil {
		return errors.New("invalid ScanResult: Scan is nil")
	}

	scan, err := w.backend.GetScan(ctx, scanResult.Scan.Id, models.GetScansScanIDParams{
		Select: utils.PointerTo("id,scanConfigSnapshot"),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch Scan with %s id: %w", scanResult.Scan.Id, err)
	}
	if scan == nil || scan.ScanConfigSnapshot == nil {
		return errors.New("invalid API response: Scan and/or Scan.ScanConfigSnapshot are nil")
	}
	scanConfig := scan.ScanConfigSnapshot

	// Check whether we have reached the maximum number of running scans
	// TODO(chrisgacsal): the number of concurrent scans needs to be part of the provider config and handled there
	filter := fmt.Sprintf("scan/id eq '%s' and status/general/state ne '%s' and status/general/state ne '%s' and resourceCleanup eq '%s'",
		scanResult.Scan.Id, models.TargetScanStateStateDONE, models.TargetScanStateStateINIT, models.ResourceCleanupStatePENDING)
	scanResults, err := w.backend.GetScanResults(ctx, models.GetScanResultsParams{
		Filter: utils.PointerTo(filter),
		Count:  utils.PointerTo(true),
		Top:    utils.PointerTo(0),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch in-progress TargetScanResults for Scan. ScanID=%s: %w",
			scanResult.Scan.Id, err)
	}

	var scanResultsInProgress int
	if scanResults.Count != nil {
		scanResultsInProgress = *scanResults.Count
	} else {
		return errors.New("invalid API response: Count is nil")
	}

	maxParallelScanners := scanConfig.GetMaxParallelScanners()
	if scanResultsInProgress >= maxParallelScanners {
		logger.Infof("Reconciliation is skipped as maximum number of running scans is reached: %d", maxParallelScanners)
		return nil
	}

	target, err := w.backend.GetTarget(ctx, scanResult.Target.Id, models.GetTargetsTargetIDParams{
		Select: utils.PointerTo("id,targetInfo"),
	})
	if err != nil {
		return fmt.Errorf("failed to fetch Target. TargetID=%s: %w", scanResult.Target.Id, err)
	}
	if target.TargetInfo == nil {
		return errors.New("invalid API response: TargetInfo is nil")
	}

	// Run scan for ScanResult
	jobConfig, err := newJobConfig(&jobConfigInput{
		config:     &w.scannerConfig,
		scanResult: scanResult,
		scanConfig: scanConfig,
		target:     &target,
	})
	if err != nil {
		return fmt.Errorf("failed to create ScanJobConfig for ScanResult. ScanResult=%s: %w", scanResultID, err)
	}

	err = w.provider.RunTargetScan(ctx, jobConfig)

	var fatalError provider.FatalError
	var retryableError provider.RetryableError
	switch {
	case errors.As(err, &fatalError):
		scanResult.Status.General.State = utils.PointerTo(models.TargetScanStateStateDONE)
		scanResult.Status.General.Errors = utils.PointerTo([]string{fatalError.Error()})
		scanResult.Status.General.LastTransitionTime = utils.PointerTo(time.Now().UTC())
	case errors.As(err, &retryableError):
		// nolint:wrapcheck
		return common.NewRequeueAfterError(retryableError.RetryAfter(), retryableError.Error())
	case err != nil:
		scanResult.Status.General.State = utils.PointerTo(models.TargetScanStateStateDONE)
		scanResult.Status.General.Errors = utils.PointerTo(utils.UnwrapErrorStrings(err))
		scanResult.Status.General.LastTransitionTime = utils.PointerTo(time.Now().UTC())
	default:
		scanResult.Status.General.State = utils.PointerTo(models.TargetScanStateStateATTACHED)
		scanResult.Status.General.LastTransitionTime = utils.PointerTo(time.Now().UTC())
	}

	scanResultPatch := models.TargetScanResult{
		Status: scanResult.Status,
	}
	err = w.backend.PatchScanResult(ctx, scanResultPatch, scanResultID)
	if err != nil {
		return fmt.Errorf("failed to update ScanResult. ScanResult=%s: %w", scanResultID, err)
	}

	return nil
}

func (w *Watcher) reconcileDone(ctx context.Context, scanResult *models.TargetScanResult) error {
	if scanResult.Scan == nil || scanResult.ResourceCleanup == nil {
		return errors.New("invalid ScanResult: Scan and/or ResourceCleanup are nil")
	}

	if *scanResult.ResourceCleanup != models.ResourceCleanupStatePENDING {
		return nil
	}

	if err := w.cleanupResources(ctx, scanResult); err != nil {
		return fmt.Errorf("failed to cleanup resources for ScanResults: %w", err)
	}

	return nil
}

// nolint:cyclop
func (w *Watcher) cleanupResources(ctx context.Context, scanResult *models.TargetScanResult) error {
	scanResultID, ok := scanResult.GetID()
	if !ok {
		return errors.New("invalid ScanResult: ID is nil")
	}

	isDone, ok := scanResult.IsDone()
	if !ok {
		return fmt.Errorf("invalid ScanResult: failed to determine General State. ScanResultID=%s", *scanResult.Id)
	}

	switch w.scannerConfig.DeleteJobPolicy {
	case DeleteJobPolicyNever:
		scanResult.ResourceCleanup = utils.PointerTo(models.ResourceCleanupStateSKIPPED)
	case DeleteJobPolicyOnSuccess:
		if isDone && scanResult.HasErrors() {
			scanResult.ResourceCleanup = utils.PointerTo(models.ResourceCleanupStateSKIPPED)
			break
		}
		fallthrough
	case DeleteJobPolicyAlways:
		fallthrough
	default:
		// Get Scan
		scan, err := w.backend.GetScan(ctx, scanResult.Scan.Id, models.GetScansScanIDParams{
			Select: utils.PointerTo("id,scanConfigSnapshot"),
		})
		if err != nil {
			return fmt.Errorf("failed to fetch Scan. ScanID=%s: %w", scanResult.Scan.Id, err)
		}
		if scan == nil || scan.ScanConfigSnapshot == nil {
			return errors.New("invalid API response: Scan and/or Scan.ScanConfigSnapshot are nil")
		}
		scanConfig := scan.ScanConfigSnapshot

		// Get Target
		target, err := w.backend.GetTarget(ctx, scanResult.Target.Id, models.GetTargetsTargetIDParams{
			Select: utils.PointerTo("id,targetInfo"),
		})
		if err != nil {
			return fmt.Errorf("failed to fetch Target. TargetID=%s: %w", scanResult.Target.Id, err)
		}
		if target.TargetInfo == nil {
			return errors.New("invalid API response: TargetInfo is nil")
		}

		// Create JobConfig
		jobConfig, err := newJobConfig(&jobConfigInput{
			config:     &w.scannerConfig,
			scanResult: scanResult,
			scanConfig: scanConfig,
			target:     &target,
		})
		if err != nil {
			return fmt.Errorf("failed to to create ScanJobConfigg for ScanResult. ScanResultID=%s: %w", scanResultID, err)
		}

		err = w.provider.RemoveTargetScan(ctx, jobConfig)

		var fatalError provider.FatalError
		var retryableError provider.RetryableError
		switch {
		case errors.As(err, &fatalError):
			scanResult.ResourceCleanup = utils.PointerTo(models.ResourceCleanupStateFAILED)
		case errors.As(err, &retryableError):
			// nolint:wrapcheck
			return common.NewRequeueAfterError(retryableError.RetryAfter(), retryableError.Error())
		case err != nil:
			scanResult.ResourceCleanup = utils.PointerTo(models.ResourceCleanupStateFAILED)
		default:
			scanResult.ResourceCleanup = utils.PointerTo(models.ResourceCleanupStateDONE)
		}
	}

	scanResultPatch := models.TargetScanResult{
		ResourceCleanup: scanResult.ResourceCleanup,
	}
	if err := w.backend.PatchScanResult(ctx, scanResultPatch, scanResultID); err != nil {
		return fmt.Errorf("failed to patch for ScanResult. ScanResultID=%s: %w", scanResultID, err)
	}

	return nil
}
