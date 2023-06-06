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

package scanwatcher

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/common"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/provider"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
	"github.com/openclarity/vmclarity/shared/pkg/log"
	"github.com/openclarity/vmclarity/shared/pkg/utils"
)

type (
	ScanQueue      = common.Queue[ScanReconcileEvent]
	ScanPoller     = common.Poller[ScanReconcileEvent]
	ScanReconciler = common.Reconciler[ScanReconcileEvent]
)

func New(c Config) *Watcher {
	return &Watcher{
		backend:          c.Backend,
		provider:         c.Provider,
		pollPeriod:       c.PollPeriod,
		reconcileTimeout: c.ReconcileTimeout,
		scanTimeout:      c.ScanTimeout,
		queue:            common.NewQueue[ScanReconcileEvent](),
	}
}

type Watcher struct {
	backend          *backendclient.BackendClient
	provider         provider.Provider
	pollPeriod       time.Duration
	reconcileTimeout time.Duration
	scanTimeout      time.Duration

	queue *ScanQueue
}

func (w *Watcher) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "ScanWatcher")
	ctx = log.SetLoggerForContext(ctx, logger)

	poller := &ScanPoller{
		PollPeriod: w.pollPeriod,
		Queue:      w.queue,
		GetItems:   w.GetRunningScans,
	}
	poller.Start(ctx)

	reconciler := &ScanReconciler{
		ReconcileTimeout:  w.reconcileTimeout,
		Queue:             w.queue,
		ReconcileFunction: w.Reconcile,
	}
	reconciler.Start(ctx)
}

// nolint:cyclop
func (w *Watcher) GetRunningScans(ctx context.Context) ([]ScanReconcileEvent, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Debugf("Fetching running Scans")

	filter := fmt.Sprintf("state ne '%s' and state ne '%s'", models.ScanStateDone, models.ScanStateFailed)
	selector := "id"
	params := models.GetScansParams{
		Filter: &filter,
		Select: &selector,
		Count:  utils.PointerTo(true),
	}
	scans, err := w.backend.GetScans(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get running scans: %v", err)
	}

	switch {
	case scans.Items == nil && scans.Count == nil:
		return nil, fmt.Errorf("failed to fetch running Scans: invalid API response: %v", scans)
	case scans.Count != nil && *scans.Count <= 0:
		fallthrough
	case scans.Items != nil && len(*scans.Items) <= 0:
		return nil, nil
	}

	events := make([]ScanReconcileEvent, 0, *scans.Count)
	for _, scan := range *scans.Items {
		scanID, ok := scan.GetID()
		if !ok {
			logger.Warnf("Skipping to invalid Scan: ID is nil: %v", scan)
			continue
		}

		events = append(events, ScanReconcileEvent{
			ScanID: scanID,
		})
	}

	return events, nil
}

// nolint:cyclop
func (w *Watcher) Reconcile(ctx context.Context, event ScanReconcileEvent) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(event.ToFields())
	ctx = log.SetLoggerForContext(ctx, logger)

	logger.Infof("Reconciling Scan event")

	params := models.GetScansScanIDParams{
		Expand: utils.PointerTo("scanConfig"),
	}
	scan, err := w.backend.GetScan(ctx, event.ScanID, params)
	if err != nil || scan == nil {
		return fmt.Errorf("failed to fetch Scan. ScanID=%s: %w", event.ScanID, err)
	}

	if isScanTimedOut(scan, w.scanTimeout) {
		scan.State = utils.PointerTo(models.ScanStateAborted)
		scan.StateMessage = utils.PointerTo("Scan has been timed out")
		scan.StateReason = utils.PointerTo(models.ScanStateReasonTimedOut)

		err = w.backend.PatchScan(ctx, *scan.Id, &models.Scan{
			State:        scan.State,
			StateMessage: scan.StateMessage,
			StateReason:  scan.StateReason,
		})
		if err != nil {
			return fmt.Errorf("failed to patch Scan. ScanID=%s: %w", event.ScanID, err)
		}
	}

	state, ok := scan.GetState()
	if !ok {
		return fmt.Errorf("failed to determine state of Scan. ScanID=%s", event.ScanID)
	}
	logger.Tracef("Reconciling Scan state: %s", state)

	switch state {
	case models.ScanStatePending:
		if err = w.reconcilePending(ctx, scan); err != nil {
			return err
		}
	case models.ScanStateDiscovered:
		if err = w.reconcileDiscovered(ctx, scan); err != nil {
			return err
		}
	case models.ScanStateInProgress:
		if err = w.reconcileInProgress(ctx, scan); err != nil {
			return err
		}
	case models.ScanStateAborted:
		if err = w.reconcileAborted(ctx, scan); err != nil {
			return err
		}
	case models.ScanStateDone, models.ScanStateFailed:
		logger.Debug("Reconciling Scan is skipped as it is already finished.")
		fallthrough
	default:
		return nil
	}

	return nil
}

func (w *Watcher) reconcilePending(ctx context.Context, scan *models.Scan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	if scan == nil {
		return errors.New("invalid Scan: object is nil")
	}

	scanID, ok := scan.GetID()
	if !ok {
		return errors.New("invalid Scan: Id is nil")
	}

	scope, ok := scan.GetScanConfigScope()
	if !ok {
		return fmt.Errorf("invalid Scan: Scope is nil. ScanID=%s", scanID)
	}

	targets, err := w.provider.DiscoverTargets(ctx, &scope)
	if err != nil {
		return fmt.Errorf("failed to discover Targets for Scan. ScanID=%s: %w", scanID, err)
	}
	numOfTargets := len(targets)

	if numOfTargets > 0 {
		if err = w.createTargets(ctx, scan, targets); err != nil {
			return fmt.Errorf("failed to create Targets for Scan. ScanID=%s: %w", scanID, err)
		}
		scan.State = utils.PointerTo(models.ScanStateDiscovered)
		scan.StateMessage = utils.PointerTo("Targets for Scan are successfully discovered")
	} else {
		scan.State = utils.PointerTo(models.ScanStateDone)
		scan.StateReason = utils.PointerTo(models.ScanStateReasonNothingToScan)
		scan.StateMessage = utils.PointerTo("No instances found in scope for Scan")
	}
	logger.Debugf("%d Target(s) have been created for Scan", numOfTargets)

	scanPatch := &models.Scan{
		TargetIDs:    scan.TargetIDs,
		State:        scan.State,
		StateReason:  scan.StateReason,
		StateMessage: scan.StateMessage,
	}

	if err = w.backend.PatchScan(ctx, scanID, scanPatch); err != nil {
		return fmt.Errorf("failed to patch Scan. ScanID=%s: %w", scanID, err)
	}

	return nil
}

func (w *Watcher) createTargets(ctx context.Context, scan *models.Scan, targetTypes []models.TargetType) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	var creatingTargetsFailed bool
	var wg sync.WaitGroup

	results := make(chan string, len(targetTypes))
	for _, t := range targetTypes {
		targetType := t

		wg.Add(1)
		go func() {
			defer wg.Done()

			targetID, err := w.createTarget(ctx, targetType)
			if err != nil {
				creatingTargetsFailed = true
				return
			}

			logger.WithField("TargetID", targetID).Trace("Pushing Target to channel")
			results <- targetID
		}()
	}
	logger.Trace("Waiting until all Target(s) are created")
	wg.Wait()
	close(results)

	if creatingTargetsFailed {
		return fmt.Errorf("failed to create Target(s) for Scan. ScanID=%s", *scan.Id)
	}

	targetIDs := make([]string, 0)
	for targetID := range results {
		targetIDs = append(targetIDs, targetID)
	}
	scan.TargetIDs = &targetIDs

	logger.Tracef("Created Target(s): %v", targetIDs)

	return nil
}

func (w *Watcher) createTarget(ctx context.Context, targetType models.TargetType) (string, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	target, err := w.backend.PostTarget(ctx, models.Target{
		TargetInfo: &targetType,
	})
	if err != nil {
		var conErr backendclient.TargetConflictError
		if errors.As(err, &conErr) {
			logger.WithField("TargetID", *conErr.ConflictingTarget.Id).Trace("Target already exist")
			return *conErr.ConflictingTarget.Id, nil
		}
		return "", fmt.Errorf("failed to post Target: %w", err)
	}
	logger.WithField("TargetID", *target.Id).Debug("Target object created")

	return *target.Id, nil
}

func (w *Watcher) reconcileDiscovered(ctx context.Context, scan *models.Scan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	if scan == nil {
		return errors.New("invalid Scan: object is nil")
	}

	scanID, ok := scan.GetID()
	if !ok {
		return errors.New("invalid Scan: Id is nil")
	}

	if err := w.createScanResultsForScan(ctx, scan); err != nil {
		return fmt.Errorf("failed to creates ScanResult(s) for Scan. ScanID=%s: %w", scanID, err)
	}
	scan.State = utils.PointerTo(models.ScanStateInProgress)

	scanPatch := &models.Scan{
		State:     scan.State,
		Summary:   scan.Summary,
		TargetIDs: scan.TargetIDs,
	}
	err := w.backend.PatchScan(ctx, scanID, scanPatch)
	if err != nil {
		return fmt.Errorf("failed to update Scan. ScanID=%s: %w", scanID, err)
	}

	logger.Infof("Total %d unique targets for Scan", len(*scan.TargetIDs))

	return nil
}

func (w *Watcher) createScanResultsForScan(ctx context.Context, scan *models.Scan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	if scan.TargetIDs == nil || *scan.TargetIDs == nil {
		return nil
	}
	numOfTargets := len(*scan.TargetIDs)

	errs := make(chan error, numOfTargets)
	var wg sync.WaitGroup
	for _, id := range *scan.TargetIDs {
		wg.Add(1)
		targetID := id
		go func() {
			defer wg.Done()

			err := w.createScanResultForTarget(ctx, scan, targetID)
			if err != nil {
				logger.WithField("TargetID", targetID).Errorf("Failed to create TargetScanResult: %v", err)
				errs <- err

				return
			}
		}()
	}
	wg.Wait()
	close(errs)

	targetErrs := make([]error, 0, numOfTargets)
	for err := range errs {
		targetErrs = append(targetErrs, err)
	}
	numOfErrs := len(targetErrs)

	if numOfErrs > 0 {
		return fmt.Errorf("failed to create %d ScanResult(s) for Scan. ScanID=%s: %w", numOfErrs, *scan.Id, targetErrs[0])
	}

	scan.Summary.JobsLeftToRun = utils.PointerTo(numOfTargets)

	return nil
}

func (w *Watcher) createScanResultForTarget(ctx context.Context, scan *models.Scan, targetID string) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	scanResultData, err := newScanResultFromScan(scan, targetID)
	if err != nil {
		return fmt.Errorf("failed to generate new ScanResult for Scan. ScanID=%s, TargetID=%s: %w", *scan.Id, targetID, err)
	}

	_, err = w.backend.PostScanResult(ctx, *scanResultData)
	if err != nil {
		var conErr backendclient.ScanResultConflictError
		if errors.As(err, &conErr) {
			scanResultID := *conErr.ConflictingScanResult.Id
			logger.WithField("ScanResultID", scanResultID).Debug("ScanResult already exist.")
			return nil
		}
		return fmt.Errorf("failed to post ScanResult to backend API: %w", err)
	}
	return nil
}

// nolint:cyclop
func (w *Watcher) reconcileInProgress(ctx context.Context, scan *models.Scan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	if scan == nil {
		return errors.New("invalid Scan: object is nil")
	}

	scanID, ok := scan.GetID()
	if !ok {
		return errors.New("invalid Scan: ID is nil")
	}

	// FIXME(chrisgacsal):a add pagination to API queries in poller/reconciler logic by using Top/Skip
	filter := fmt.Sprintf("scan/id eq '%s'", scanID)
	selector := "id,status/general,summary"
	targetScanResults, err := w.backend.GetScanResults(ctx, models.GetScanResultsParams{
		Filter: &filter,
		Select: &selector,
		Count:  utils.PointerTo(true),
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve TargetScans for Scan. ScanID=%s: %w", scanID, err)
	}

	if targetScanResults.Count == nil || targetScanResults.Items == nil {
		return fmt.Errorf("invalid response for getting TargetScans for Scan. ScanID=%s: Count and/or Items parameters are nil", scanID)
	}

	// Reset Scan Summary as it is going to be recalculated
	scan.Summary = newScanSummary()

	var targetScanResultsWithErr int
	for _, targetScanResult := range *targetScanResults.Items {
		scanResultID, ok := targetScanResult.GetID()
		if !ok {
			return errors.New("invalid ScanResult: ID is nil")
		}

		if err := updateScanSummaryFromScanResult(scan, targetScanResult); err != nil {
			return fmt.Errorf("failed to update Scan Summary from ScanResult. ScanID=%s ScanResultID=%s: %w",
				scanID, scanResultID, err)
		}

		errs := targetScanResult.GetGeneralErrors()
		if len(errs) > 0 {
			targetScanResultsWithErr++
		}
	}
	logger.Tracef("Scan Summary updated. JobCompleted=%d JobLeftToRun=%d", *scan.Summary.JobsCompleted,
		*scan.Summary.JobsLeftToRun)

	if *scan.Summary.JobsLeftToRun <= 0 {
		if targetScanResultsWithErr > 0 {
			scan.State = utils.PointerTo(models.ScanStateFailed)
			scan.StateReason = utils.PointerTo(models.ScanStateReasonOneOrMoreTargetFailedToScan)
		} else {
			scan.State = utils.PointerTo(models.ScanStateDone)
			scan.StateReason = utils.PointerTo(models.ScanStateReasonSuccess)
		}
		scan.StateMessage = utils.PointerTo(fmt.Sprintf("%d succeeded, %d failed out of %d total target scans",
			*targetScanResults.Count-targetScanResultsWithErr, targetScanResultsWithErr, *targetScanResults.Count))

		scan.EndTime = utils.PointerTo(time.Now().UTC())
	}

	scanPatch := &models.Scan{
		State:        scan.State,
		Summary:      scan.Summary,
		StateMessage: scan.StateMessage,
		EndTime:      scan.EndTime,
		TargetIDs:    scan.TargetIDs,
	}
	err = w.backend.PatchScan(ctx, scanID, scanPatch)
	if err != nil {
		return fmt.Errorf("failed to patch Scan. ScanID=%s: %w", scanID, err)
	}

	return nil
}

// nolint:cyclop
func (w *Watcher) reconcileAborted(ctx context.Context, scan *models.Scan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	if scan == nil {
		return fmt.Errorf("scan must not be nil: %v", scan)
	}

	scanID, ok := scan.GetID()
	if !ok {
		return fmt.Errorf("scan id must not be nil: %v", scan)
	}

	filter := fmt.Sprintf("scan/id eq '%s' and status/general/state ne '%s' and status/general/state ne '%s'",
		scanID, models.TargetScanStateStateABORTED, models.TargetScanStateStateDONE)
	selector := "id,status"
	params := models.GetScanResultsParams{
		Filter: &filter,
		Select: &selector,
	}

	scanResults, err := w.backend.GetScanResults(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to get ScanResult(s) for Scan with %s id: %w", scanID, err)
	}

	if scanResults.Items != nil && len(*scanResults.Items) > 0 {
		var reconciliationFailed bool
		var wg sync.WaitGroup

		for _, scanResult := range *scanResults.Items {
			if scanResult.Id == nil {
				continue
			}
			scanResultID := *scanResult.Id

			wg.Add(1)
			go func() {
				defer wg.Done()
				sr := models.TargetScanResult{
					Status: &models.TargetScanStatus{
						General: &models.TargetScanState{
							State: utils.PointerTo(models.TargetScanStateStateABORTED),
						},
					},
				}

				err = w.backend.PatchScanResult(ctx, sr, scanResultID)
				if err != nil {
					logger.WithField("ScanResultID", scanResultID).Error("Failed to patch ScanResult")
					reconciliationFailed = true
					return
				}
			}()
		}
		wg.Wait()

		// NOTE: reconciliationFailed is used to track errors returned by patching ScanResults
		//       as setting the state of Scan to models.ScanStateFailed must be skipped in case
		//       even a single error occurred to allow reconciling re-running for this Scan.
		if reconciliationFailed {
			return errors.New("updating one or more ScanResults failed")
		}
	}

	scan.EndTime = utils.PointerTo(time.Now().UTC())
	scan.State = utils.PointerTo(models.ScanStateFailed)

	err = w.backend.PatchScan(ctx, scanID, scan)
	if err != nil {
		return fmt.Errorf("failed to patch Scan with %s id: %w", scanID, err)
	}

	return nil
}
