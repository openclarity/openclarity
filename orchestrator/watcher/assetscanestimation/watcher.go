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

package assetscanestimation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/client"
	"github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/cli/pkg/utils"
	"github.com/openclarity/vmclarity/orchestrator/common"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/utils/log"
)

type (
	AssetScanEstimationQueue      = common.Queue[AssetScanEstimationReconcileEvent]
	AssetScanEstimationPoller     = common.Poller[AssetScanEstimationReconcileEvent]
	AssetScanEstimationReconciler = common.Reconciler[AssetScanEstimationReconcileEvent]
)

func New(c Config) *Watcher {
	return &Watcher{
		backend:          c.Backend,
		provider:         c.Provider,
		pollPeriod:       c.PollPeriod,
		reconcileTimeout: c.ReconcileTimeout,
		queue:            common.NewQueue[AssetScanEstimationReconcileEvent](),
	}
}

type Watcher struct {
	backend          *client.BackendClient
	provider         provider.Provider
	pollPeriod       time.Duration
	reconcileTimeout time.Duration

	queue *AssetScanEstimationQueue
}

func (w *Watcher) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "AssetScanEstimationWatcher")
	ctx = log.SetLoggerForContext(ctx, logger)

	poller := &AssetScanEstimationPoller{
		PollPeriod: w.pollPeriod,
		Queue:      w.queue,
		GetItems:   w.GetAssetScanEstimations,
	}
	poller.Start(ctx)

	reconciler := &AssetScanEstimationReconciler{
		ReconcileTimeout:  w.reconcileTimeout,
		Queue:             w.queue,
		ReconcileFunction: w.Reconcile,
	}
	reconciler.Start(ctx)
}

// nolint:cyclop
func (w *Watcher) GetAssetScanEstimations(ctx context.Context) ([]AssetScanEstimationReconcileEvent, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Debugf("Fetching AssetScanEstimations which need to be reconciled")

	filter := fmt.Sprintf("(status/state ne '%s' and status/state ne '%s') or (deleteAfter eq null or deleteAfter lt %s)",
		types.AssetScanEstimationStatusStateDone, types.AssetScanEstimationStatusStateFailed, time.Now().Format(time.RFC3339))
	selector := "id,scanEstimation/id,asset/id"
	params := types.GetAssetScanEstimationsParams{
		Filter: &filter,
		Select: &selector,
		Count:  utils.PointerTo(true),
	}
	assetScanEstimations, err := w.backend.GetAssetScanEstimations(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get AssetScanEstimations: %w", err)
	}

	switch {
	case assetScanEstimations.Items == nil && assetScanEstimations.Count == nil:
		return nil, fmt.Errorf("failed to fetch AssetScanEstimations: invalid API response: %v", assetScanEstimations)
	case assetScanEstimations.Count != nil && *assetScanEstimations.Count <= 0:
		fallthrough
	case assetScanEstimations.Items != nil && len(*assetScanEstimations.Items) <= 0:
		return nil, nil
	}

	events := make([]AssetScanEstimationReconcileEvent, 0, len(*assetScanEstimations.Items))
	for _, assetScanEstimation := range *assetScanEstimations.Items {
		assetScanEstimationID, ok := assetScanEstimation.GetID()
		if !ok {
			logger.Warnf("Skipping due to invalid AssetScanEstimation: ID is nil: %v", assetScanEstimation)
			continue
		}
		scanEstimationID, ok := assetScanEstimation.GetScanEstimationID()
		if !ok {
			// AssetScanEstimations should be usable by themselves and not part of a scanEstimation.
			logger.Debugf("ScanEstimation.ID is nil: %v", assetScanEstimation)
		}
		assetID, ok := assetScanEstimation.GetAssetID()
		if !ok {
			logger.Warnf("Skipping due to invalid AssetScanEstimation: Asset.ID is nil: %v", assetScanEstimation)
			continue
		}

		events = append(events, AssetScanEstimationReconcileEvent{
			AssetScanEstimationID: assetScanEstimationID,
			ScanEstimationID:      scanEstimationID,
			AssetID:               assetID,
		})
	}

	return events, nil
}

// nolint:cyclop
func (w *Watcher) Reconcile(ctx context.Context, event AssetScanEstimationReconcileEvent) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(event.ToFields())
	ctx = log.SetLoggerForContext(ctx, logger)

	assetScanEstimation, err := w.backend.GetAssetScanEstimation(ctx, event.AssetScanEstimationID, types.GetAssetScanEstimationsAssetScanEstimationIDParams{
		Expand: utils.PointerTo("asset"),
	})
	if err != nil {
		return fmt.Errorf("failed to get AssetScanEstimation with %s id: %w", event.AssetScanEstimationID, err)
	}

	status, ok := assetScanEstimation.GetStatus()
	if !ok {
		if err = w.reconcileNoState(ctx, &assetScanEstimation); err != nil {
			return err
		}
		return nil
	}

	logger.Tracef("Reconciling AssetScanEstimation state: %s", status.State)

	switch status.State {
	case types.AssetScanEstimationStatusStatePending:
		if err = w.reconcilePending(ctx, &assetScanEstimation); err != nil {
			return err
		}
	case types.AssetScanEstimationStatusStateAborted:
		if err = w.reconcileAborted(ctx, &assetScanEstimation); err != nil {
			return err
		}
	case types.AssetScanEstimationStatusStateFailed, types.AssetScanEstimationStatusStateDone:
		if err = w.reconcileDone(ctx, &assetScanEstimation); err != nil {
			return err
		}
	default:
	}

	return nil
}

func (w *Watcher) reconcileDone(ctx context.Context, assetScanEstimation *types.AssetScanEstimation) error {
	if assetScanEstimation.EndTime == nil {
		assetScanEstimation.EndTime = utils.PointerTo(time.Now())
	}
	if assetScanEstimation.TTLSecondsAfterFinished == nil {
		assetScanEstimation.TTLSecondsAfterFinished = utils.PointerTo(DefaultAssetScanEstimationTTLSeconds)
	}

	endTime := *assetScanEstimation.EndTime
	ttl := *assetScanEstimation.TTLSecondsAfterFinished

	assetScanEstimationID, ok := assetScanEstimation.GetID()
	if !ok {
		return errors.New("invalid AssetScanEstimation: ID is nil")
	}

	timeNow := time.Now()

	if assetScanEstimation.DeleteAfter == nil {
		assetScanEstimation.DeleteAfter = utils.PointerTo(endTime.Add(time.Duration(ttl) * time.Second))
		// if delete time has already pass, no need to patch the object, just delete it.
		if !timeNow.After(*assetScanEstimation.DeleteAfter) {
			assetScanEstimationPatch := types.AssetScanEstimation{
				DeleteAfter:             assetScanEstimation.DeleteAfter,
				EndTime:                 assetScanEstimation.EndTime,
				TTLSecondsAfterFinished: assetScanEstimation.TTLSecondsAfterFinished,
			}
			err := w.backend.PatchAssetScanEstimation(ctx, assetScanEstimationPatch, assetScanEstimationID)
			if err != nil {
				return fmt.Errorf("failed to patch AssetScanEstimation. AssetScanEstimationID=%v: %w", assetScanEstimationID, err)
			}
			return nil
		}
	}

	if timeNow.After(*assetScanEstimation.DeleteAfter) {
		err := w.backend.DeleteAssetScanEstimation(ctx, assetScanEstimationID)
		if err != nil {
			return fmt.Errorf("failed to delete AssetScanEstimation. AssetScanEstimationID=%v: %w", assetScanEstimationID, err)
		}
	}

	return nil
}

func (w *Watcher) reconcileNoState(ctx context.Context, assetScanEstimation *types.AssetScanEstimation) error {
	assetScanEstimationID, ok := assetScanEstimation.GetID()
	if !ok {
		return errors.New("invalid AssetScanEstimation: ID is nil")
	}

	assetScanEstimationPatch := types.AssetScanEstimation{
		Status: types.NewAssetScanEstimationStatus(
			types.AssetScanEstimationStatusStatePending,
			types.AssetScanEstimationStatusReasonCreated,
			nil,
		),
	}

	err := w.backend.PatchAssetScanEstimation(ctx, assetScanEstimationPatch, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScanEstimation. AssetScanEstimationID=%v: %w", assetScanEstimationID, err)
	}
	return nil
}

// nolint:cyclop
func (w *Watcher) reconcilePending(ctx context.Context, assetScanEstimation *types.AssetScanEstimation) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	logger.Debugf("Reconciling pending asset scan estimations")

	assetScanEstimationID, ok := assetScanEstimation.GetID()
	if !ok {
		return errors.New("invalid AssetScanEstimation: ID is nil")
	}

	if assetScanEstimation.Asset == nil || assetScanEstimation.Asset.AssetInfo == nil {
		return errors.New("invalid AssetScanEstimation: Asset or AssetInfo is nil")
	}

	asset, err := assetScanEstimation.Asset.ToAsset()
	if err != nil {
		return fmt.Errorf("failed to convert assetRelationship to asset: %w", err)
	}

	stats := w.getLatestAssetScanStats(ctx, asset)
	startTime := time.Now()

	estimation, err := w.provider.Estimate(ctx, stats, asset, assetScanEstimation.AssetScanTemplate)

	endTime := time.Now()

	var fatalError provider.FatalError
	var retryableError provider.RetryableError
	switch {
	case errors.As(err, &fatalError):
		logger.Errorf("Fatal error while estimating asset scan: %v", err)
		assetScanEstimation.Status = types.NewAssetScanEstimationStatus(
			types.AssetScanEstimationStatusStateFailed,
			types.AssetScanEstimationStatusReasonError,
			utils.PointerTo(fatalError.Error()),
		)
	case errors.As(err, &retryableError):
		// nolint:wrapcheck
		return common.NewRequeueAfterError(retryableError.RetryAfter(), retryableError.Error())
	case err != nil:
		return fmt.Errorf("failed to estimate asset scan: %w", err)
	default:
		logger.Infof("Asset scan estimation completed successfully. Estimation=%v", estimation)
		assetScanEstimation.Status = types.NewAssetScanEstimationStatus(
			types.AssetScanEstimationStatusStateDone,
			types.AssetScanEstimationStatusReasonSuccess,
			nil,
		)
		assetScanEstimation.Estimation = &types.Estimation{
			Cost:          estimation.Cost,
			Size:          estimation.Size,
			Duration:      estimation.Duration,
			CostBreakdown: estimation.CostBreakdown,
		}
	}

	// Set default ttl if not set.
	if assetScanEstimation.TTLSecondsAfterFinished == nil {
		assetScanEstimation.TTLSecondsAfterFinished = utils.PointerTo(DefaultAssetScanEstimationTTLSeconds)
	}

	assetScanEstimationPatch := types.AssetScanEstimation{
		StartTime:               utils.PointerTo(startTime),
		EndTime:                 utils.PointerTo(endTime),
		DeleteAfter:             utils.PointerTo(endTime.Add(time.Duration(*assetScanEstimation.TTLSecondsAfterFinished) * time.Second)),
		Status:                  assetScanEstimation.Status,
		Estimation:              assetScanEstimation.Estimation,
		TTLSecondsAfterFinished: assetScanEstimation.TTLSecondsAfterFinished,
	}
	err = w.backend.PatchAssetScanEstimation(ctx, assetScanEstimationPatch, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScanEstimation. AssetScanEstimationID=%v: %w", assetScanEstimationID, err)
	}
	return nil
}

// nolint:cyclop
func (w *Watcher) reconcileAborted(ctx context.Context, assetScanEstimation *types.AssetScanEstimation) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	assetScanEstimationID, ok := assetScanEstimation.GetID()
	if !ok {
		return errors.New("invalid AssetScanEstimation: ID is nil")
	}

	if assetScanEstimation.Status == nil {
		return errors.New("invalid AssetScanEstimation: State is nil")
	}

	assetScanEstimationPatch := types.AssetScanEstimation{
		Status: types.NewAssetScanEstimationStatus(
			types.AssetScanEstimationStatusStateFailed,
			types.AssetScanEstimationStatusReasonAborted,
			utils.PointerTo("asset scan estimation was aborted"),
		),
	}

	err := w.backend.PatchAssetScanEstimation(ctx, assetScanEstimationPatch, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScanEstimation. AssetScanEstimation=%s: %w", assetScanEstimationID, err)
	}

	logger.Infof("AssetScanEstimation successfully aborted. AssetScanEstimationID=%v", assetScanEstimationID)

	return nil
}
