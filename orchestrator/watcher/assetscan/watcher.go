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

package assetscan

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/orchestrator/common"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/scanner/utils"
)

type (
	AssetScanQueue      = common.Queue[AssetScanReconcileEvent]
	AssetScanPoller     = common.Poller[AssetScanReconcileEvent]
	AssetScanReconciler = common.Reconciler[AssetScanReconcileEvent]
)

func New(c Config) *Watcher {
	return &Watcher{
		client:           c.Client,
		provider:         c.Provider,
		scannerConfig:    c.ScannerConfig,
		pollPeriod:       c.PollPeriod,
		reconcileTimeout: c.ReconcileTimeout,
		deleteJobPolicy:  c.DeleteJobPolicy,
		abortTimeout:     c.AbortTimeout,
		queue:            common.NewQueue[AssetScanReconcileEvent](),
	}
}

type Watcher struct {
	client           *apiclient.Client
	provider         provider.Provider
	scannerConfig    ScannerConfig
	pollPeriod       time.Duration
	reconcileTimeout time.Duration
	deleteJobPolicy  DeleteJobPolicyType
	abortTimeout     time.Duration

	queue *AssetScanQueue
}

func (w *Watcher) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "AssetScanWatcher")
	ctx = log.SetLoggerForContext(ctx, logger)

	poller := &AssetScanPoller{
		PollPeriod: w.pollPeriod,
		Queue:      w.queue,
		GetItems:   w.GetAssetScans,
	}
	poller.Start(ctx)

	reconciler := &AssetScanReconciler{
		ReconcileTimeout:  w.reconcileTimeout,
		Queue:             w.queue,
		ReconcileFunction: w.Reconcile,
	}
	reconciler.Start(ctx)
}

// nolint:cyclop
func (w *Watcher) GetAssetScans(ctx context.Context) ([]AssetScanReconcileEvent, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Debugf("Fetching AssetScans which need to be reconciled")

	filter := fmt.Sprintf("(status/state ne '%s' and status/state ne '%s') or resourceCleanupStatus/state eq '%s'",
		apitypes.AssetScanStatusStateDone, apitypes.AssetScanStatusStateFailed, apitypes.ResourceCleanupStatusStatePending)
	selector := "id,scan/id,asset/id"
	params := apitypes.GetAssetScansParams{
		Filter: &filter,
		Select: &selector,
		Count:  to.Ptr(true),
	}
	assetScans, err := w.client.GetAssetScans(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetScans: %w", err)
	}

	switch {
	case assetScans.Items == nil && assetScans.Count == nil:
		return nil, fmt.Errorf("failed to fetch AssetScans: invalid API response: %v", assetScans)
	case assetScans.Count != nil && *assetScans.Count <= 0:
		fallthrough
	case assetScans.Items != nil && len(*assetScans.Items) <= 0:
		return nil, nil
	}

	events := make([]AssetScanReconcileEvent, 0, len(*assetScans.Items))
	for _, assetScan := range *assetScans.Items {
		assetScanID, ok := assetScan.GetID()
		if !ok {
			logger.Warnf("Skipping to invalid AssetScan: ID is nil: %v", assetScan)
			continue
		}
		scanID, ok := assetScan.GetScanID()
		if !ok {
			logger.Warnf("Skipping to invalid AssetScan: Scan.ID is nil: %v", assetScan)
			continue
		}
		assetID, ok := assetScan.GetAssetID()
		if !ok {
			logger.Warnf("Skipping to invalid AssetScan: Asset.ID is nil: %v", assetScan)
			continue
		}

		events = append(events, AssetScanReconcileEvent{
			AssetScanID: assetScanID,
			ScanID:      scanID,
			AssetID:     assetID,
		})
	}

	return events, nil
}

// nolint:cyclop
func (w *Watcher) Reconcile(ctx context.Context, event AssetScanReconcileEvent) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(event.ToFields())
	ctx = log.SetLoggerForContext(ctx, logger)

	assetScan, err := w.client.GetAssetScan(ctx, event.AssetScanID, apitypes.GetAssetScansAssetScanIDParams{
		Expand: to.Ptr("scan,asset"),
	})
	if err != nil {
		return fmt.Errorf("failed to get AssetScan with %s id: %w", event.AssetScanID, err)
	}

	status, ok := assetScan.GetStatus()
	if !ok {
		return fmt.Errorf("cannot determine status of AssetScan with %s id", event.AssetScanID)
	}

	logger.Tracef("Reconciling AssetScan state: %s", status.State)

	switch status.State {
	case apitypes.AssetScanStatusStatePending:
		if err = w.reconcilePending(ctx, &assetScan); err != nil {
			return err
		}
	case apitypes.AssetScanStatusStateScheduled:
		if err = w.reconcileScheduled(ctx, &assetScan); err != nil {
			return err
		}
	case apitypes.AssetScanStatusStateReadyToScan, apitypes.AssetScanStatusStateInProgress:
		// TODO(chrisgacsal): make sure that AssetScan state is set to ABORTED state once the AssetScan
		//                    schema is extended with timeout field and the deadline is missed.
		break
	case apitypes.AssetScanStatusStateAborted:
		if err = w.reconcileAborted(ctx, &assetScan); err != nil {
			return err
		}
	case apitypes.AssetScanStatusStateDone, apitypes.AssetScanStatusStateFailed:
		if err = w.reconcileDoneAndFailed(ctx, &assetScan); err != nil {
			return err
		}
	default:
	}

	return nil
}

// nolint:cyclop
func (w *Watcher) reconcilePending(ctx context.Context, assetScan *apitypes.AssetScan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	assetScanID, ok := assetScan.GetID()
	if !ok {
		return errors.New("invalid AssetScan: ID is nil")
	}

	// If this asset scan has a scan associated with it, and its configured
	// to only allow max parallel scanners, then check that scenario.
	// TODO(sambetts) Replace this with provider level scheduling.
	if assetScan.Scan != nil && assetScan.Scan.Id != "" && assetScan.Scan.MaxParallelScanners != nil {
		// Check whether we have reached the maximum number of running scans
		// TODO(chrisgacsal): the number of concurrent scans needs to be part of the provider config and handled there
		filter := fmt.Sprintf("scan/id eq '%s' and status/state ne '%s' and status/state ne '%s' and status/state ne '%s' and resourceCleanupStatus/state eq '%s'",
			assetScan.Scan.Id, apitypes.AssetScanStatusStateDone, apitypes.AssetScanStatusStateFailed, apitypes.AssetScanStatusStatePending, apitypes.ResourceCleanupStatusStatePending)
		assetScans, err := w.client.GetAssetScans(ctx, apitypes.GetAssetScansParams{
			Filter: to.Ptr(filter),
			Count:  to.Ptr(true),
			Top:    to.Ptr(0),
		})
		if err != nil {
			return fmt.Errorf("failed to fetch in-progress AssetScans for Scan. ScanID=%s: %w",
				assetScan.Scan.Id, err)
		}

		var assetScansInProgress int
		if assetScans.Count != nil {
			assetScansInProgress = *assetScans.Count
		} else {
			return errors.New("invalid API response: Count is nil")
		}

		maxParallelScanners := assetScan.Scan.GetMaxParallelScanners()
		if assetScansInProgress >= maxParallelScanners {
			logger.Infof("Reconciliation is skipped as maximum number of running scans is reached: %d", maxParallelScanners)
			return nil
		}
	}

	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateScheduled,
		apitypes.AssetScanStatusReasonProvisioning,
		nil,
	)

	assetScanPatch := apitypes.AssetScan{
		Status: assetScan.Status,
	}
	err := w.client.PatchAssetScan(ctx, assetScanPatch, assetScanID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScan. AssetScan=%s: %w", assetScanID, err)
	}

	// nolint:wrapcheck
	return common.NewRequeueAfterError(time.Second,
		"AssetScan state moved to Scheduled. Skip waiting for another reconcile cycle. AssetScanID="+assetScanID)
}

// nolint:cyclop
func (w *Watcher) reconcileScheduled(ctx context.Context, assetScan *apitypes.AssetScan) error {
	assetScanID, ok := assetScan.GetID()
	if !ok {
		return errors.New("invalid AssetScan: ID is nil")
	}

	if assetScan.Asset == nil || assetScan.Asset.AssetInfo == nil {
		return errors.New("invalid AssetScan: Asset or AssetInfo is nil")
	}
	asset := &apitypes.Asset{
		Id:         to.Ptr(assetScan.Asset.Id),
		Revision:   assetScan.Asset.Revision,
		ScansCount: assetScan.Asset.ScansCount,
		Summary:    assetScan.Asset.Summary,
		AssetInfo:  assetScan.Asset.AssetInfo,
	}

	// Run scan for AssetScan
	jobConfig, err := newJobConfig(&jobConfigInput{
		config:    &w.scannerConfig,
		assetScan: assetScan,
		asset:     asset,
	})
	if err != nil {
		return fmt.Errorf("failed to create ScanJobConfig for AssetScan. AssetScan=%s: %w", assetScanID, err)
	}

	err = w.provider.RunAssetScan(ctx, jobConfig)

	var fatalError provider.FatalError
	var retryableError provider.RetryableError
	switch {
	case errors.As(err, &fatalError):
		assetScan.Status = apitypes.NewAssetScanStatus(
			apitypes.AssetScanStatusStateFailed,
			apitypes.AssetScanStatusReasonError,
			to.Ptr(fatalError.Error()),
		)
	case errors.As(err, &retryableError):
		// nolint:wrapcheck
		return common.NewRequeueAfterError(retryableError.RetryAfter(), retryableError.Error())
	case err != nil:
		assetScan.Status = apitypes.NewAssetScanStatus(
			apitypes.AssetScanStatusStateFailed,
			apitypes.AssetScanStatusReasonError,
			to.Ptr(errors.Join(utils.UnwrapErrors(err)...).Error()),
		)
	default:
		assetScan.Status = apitypes.NewAssetScanStatus(
			apitypes.AssetScanStatusStateReadyToScan,
			apitypes.AssetScanStatusReasonResourcesReady,
			nil,
		)
	}

	assetScanPatch := apitypes.AssetScan{
		Status: assetScan.Status,
	}
	err = w.client.PatchAssetScan(ctx, assetScanPatch, assetScanID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScan. AssetScan=%s: %w", assetScanID, err)
	}

	return nil
}

func (w *Watcher) reconcileDoneAndFailed(ctx context.Context, assetScan *apitypes.AssetScan) error {
	if assetScan.Scan == nil || assetScan.ResourceCleanupStatus == nil {
		return errors.New("invalid AssetScan: Scan and/or ResourceCleanupStatus are nil")
	}

	if assetScan.ResourceCleanupStatus.State != apitypes.ResourceCleanupStatusStatePending {
		return nil
	}

	if err := w.cleanupResources(ctx, assetScan); err != nil {
		return fmt.Errorf("failed to cleanup resources for AssetScans: %w", err)
	}

	return nil
}

// nolint:cyclop
func (w *Watcher) cleanupResources(ctx context.Context, assetScan *apitypes.AssetScan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	assetScanID, ok := assetScan.GetID()
	if !ok {
		return errors.New("invalid AssetScan: ID is nil")
	}

	switch w.deleteJobPolicy {
	case DeleteJobPolicyNever:
		assetScan.ResourceCleanupStatus = apitypes.NewResourceCleanupStatus(
			apitypes.ResourceCleanupStatusStateSkipped,
			apitypes.ResourceCleanupStatusReasonDeletePolicy,
			nil,
		)
	case DeleteJobPolicyOnSuccess:
		if assetScan.Status.State == apitypes.AssetScanStatusStateFailed {
			assetScan.ResourceCleanupStatus = apitypes.NewResourceCleanupStatus(
				apitypes.ResourceCleanupStatusStateSkipped,
				apitypes.ResourceCleanupStatusReasonDeletePolicy,
				nil,
			)
			break
		}
		fallthrough
	case DeleteJobPolicyAlways:
		fallthrough
	default:
		// Get Asset
		asset, err := w.client.GetAsset(ctx, assetScan.Asset.Id, apitypes.GetAssetsAssetIDParams{
			Select: to.Ptr("id,assetInfo"),
		})
		if err != nil {
			return fmt.Errorf("failed to fetch Asset. AssetID=%s: %w", assetScan.Asset.Id, err)
		}
		if asset.AssetInfo == nil {
			return errors.New("invalid API response: AssetInfo is nil")
		}

		// Create JobConfig
		jobConfig, err := newJobConfig(&jobConfigInput{
			config:    &w.scannerConfig,
			assetScan: assetScan,
			asset:     &asset,
		})
		if err != nil {
			return fmt.Errorf("failed to create ScanJobConfig for AssetScan. AssetScanID=%s: %w", assetScanID, err)
		}

		err = w.provider.RemoveAssetScan(ctx, jobConfig)

		var fatalError provider.FatalError
		var retryableError provider.RetryableError
		switch {
		case errors.As(err, &fatalError):
			assetScan.ResourceCleanupStatus = apitypes.NewResourceCleanupStatus(
				apitypes.ResourceCleanupStatusStateFailed,
				apitypes.ResourceCleanupStatusReasonProviderError,
				to.Ptr(fatalError.Error()),
			)
			logger.Errorf("resource cleanup failed: %v", fatalError)
		case errors.As(err, &retryableError):
			// nolint:wrapcheck
			return common.NewRequeueAfterError(retryableError.RetryAfter(), retryableError.Error())
		case err != nil:
			assetScan.ResourceCleanupStatus = apitypes.NewResourceCleanupStatus(
				apitypes.ResourceCleanupStatusStateFailed,
				apitypes.ResourceCleanupStatusReasonProviderError,
				to.Ptr(err.Error()),
			)
			logger.Errorf("resource cleanup failed: %v", err)
		default:
			assetScan.ResourceCleanupStatus = apitypes.NewResourceCleanupStatus(
				apitypes.ResourceCleanupStatusStateDone,
				apitypes.ResourceCleanupStatusReasonSuccess,
				nil,
			)
		}
	}

	assetScanPatch := apitypes.AssetScan{
		ResourceCleanupStatus: assetScan.ResourceCleanupStatus,
	}
	if err := w.client.PatchAssetScan(ctx, assetScanPatch, assetScanID); err != nil {
		return fmt.Errorf("failed to patch for AssetScan. AssetScanID=%s: %w", assetScanID, err)
	}

	return nil
}

// nolint:cyclop
func (w *Watcher) reconcileAborted(ctx context.Context, assetScan *apitypes.AssetScan) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	assetScanID, ok := assetScan.GetID()
	if !ok {
		return errors.New("invalid AssetScan: ID is nil")
	}

	// Check if AssetScan is in aborted state for more time than the timeout allows
	if assetScan.Status == nil {
		return errors.New("invalid AssetScan: Status is nil")
	}

	transitionTimeToAbort := assetScan.Status.LastTransitionTime
	logger.Debugf("AssetScan moved to aborted state: %s", transitionTimeToAbort)

	now := time.Now()
	abortTimedOut := now.After(transitionTimeToAbort.Add(w.abortTimeout))
	if !abortTimedOut {
		logger.Tracef("AssetScan in aborted state is not timed out yet. TransitionTime=%s Timeout=%s",
			transitionTimeToAbort, w.abortTimeout)
		return nil
	}
	logger.Tracef("AssetScan in aborted state is timed out. TransitionTime=%s Timeout=%s",
		transitionTimeToAbort, w.abortTimeout)

	assetScan.Status = apitypes.NewAssetScanStatus(
		apitypes.AssetScanStatusStateFailed,
		apitypes.AssetScanStatusReasonAbortTimeout,
		to.Ptr(fmt.Sprintf("failed to wait for scanner to finish graceful shutdown on abort after: %s", w.abortTimeout)),
	)

	assetScanPatch := apitypes.AssetScan{
		Status: assetScan.Status,
	}
	err := w.client.PatchAssetScan(ctx, assetScanPatch, assetScanID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScan. AssetScan=%s: %w", assetScanID, err)
	}

	return nil
}
