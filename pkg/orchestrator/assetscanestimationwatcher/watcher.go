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

package assetscanestimationwatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/common"
	"github.com/openclarity/vmclarity/pkg/orchestrator/provider"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
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
	backend          *backendclient.BackendClient
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

	filter := fmt.Sprintf("(state/state ne '%s' and state/state ne '%s') or (deleteAfter eq null or deleteAfter lt %s)",
		models.AssetScanEstimationStateStateDone, models.AssetScanEstimationStateStateFailed, time.Now().Format(time.RFC3339))
	selector := "id,scanEstimation/id,asset/id"
	params := models.GetAssetScanEstimationsParams{
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

	assetScanEstimation, err := w.backend.GetAssetScanEstimation(ctx, event.AssetScanEstimationID, models.GetAssetScanEstimationsAssetScanEstimationIDParams{
		Expand: utils.PointerTo("asset"),
	})
	if err != nil {
		return fmt.Errorf("failed to get AssetScanEstimation with %s id: %w", event.AssetScanEstimationID, err)
	}

	state, ok := assetScanEstimation.GetState()
	if !ok {
		if err = w.reconcileNoState(ctx, &assetScanEstimation); err != nil {
			return err
		}
		return nil
	}

	logger.Tracef("Reconciling AssetScanEstimation state: %s", state)

	switch state {
	case models.AssetScanEstimationStateStatePending:
		if err = w.reconcilePending(ctx, &assetScanEstimation); err != nil {
			return err
		}
	case models.AssetScanEstimationStateStateAborted:
		if err = w.reconcileAborted(ctx, &assetScanEstimation); err != nil {
			return err
		}
	case models.AssetScanEstimationStateStateFailed, models.AssetScanEstimationStateStateDone:
		if err = w.reconcileDone(ctx, &assetScanEstimation); err != nil {
			return err
		}
	default:
	}

	return nil
}

func (w *Watcher) reconcileDone(ctx context.Context, assetScanEstimation *models.AssetScanEstimation) error {
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
			assetScanEstimationPatch := models.AssetScanEstimation{
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

func (w *Watcher) reconcileNoState(ctx context.Context, assetScanEstimation *models.AssetScanEstimation) error {
	assetScanEstimationID, ok := assetScanEstimation.GetID()
	if !ok {
		return errors.New("invalid AssetScanEstimation: ID is nil")
	}

	assetScanEstimation.State.State = utils.PointerTo(models.AssetScanEstimationStateStatePending)
	assetScanEstimation.State.LastTransitionTime = utils.PointerTo(time.Now())

	assetScanEstimationPatch := models.AssetScanEstimation{
		State: assetScanEstimation.State,
	}
	err := w.backend.PatchAssetScanEstimation(ctx, assetScanEstimationPatch, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScanEstimation. AssetScanEstimationID=%v: %w", assetScanEstimationID, err)
	}
	return nil
}

// nolint:cyclop
func (w *Watcher) reconcilePending(ctx context.Context, assetScanEstimation *models.AssetScanEstimation) error {
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
		assetScanEstimation.State.State = utils.PointerTo(models.AssetScanEstimationStateStateFailed)
		assetScanEstimation.State.StateMessage = utils.PointerTo(fatalError.Error())
		assetScanEstimation.State.StateReason = utils.PointerTo(models.AssetScanEstimationStateStateReasonUnexpected)
	case errors.As(err, &retryableError):
		// nolint:wrapcheck
		return common.NewRequeueAfterError(retryableError.RetryAfter(), retryableError.Error())
	case err != nil:
		return fmt.Errorf("failed to estimate asset scan: %w", err)
	default:
		logger.Infof("Asset scan estimation completed successfully. Estimation=%v", estimation)
		assetScanEstimation.State.State = utils.PointerTo(models.AssetScanEstimationStateStateDone)
		assetScanEstimation.State.StateReason = utils.PointerTo(models.AssetScanEstimationStateStateReasonSuccess)
		assetScanEstimation.Estimation = &models.Estimation{
			Cost:          estimation.Cost,
			Size:          estimation.Size,
			Duration:      estimation.Duration,
			CostBreakdown: estimation.CostBreakdown,
		}
	}

	assetScanEstimation.State.LastTransitionTime = utils.PointerTo(time.Now())

	// Set default ttl if not set.
	if assetScanEstimation.TTLSecondsAfterFinished == nil {
		assetScanEstimation.TTLSecondsAfterFinished = utils.PointerTo(DefaultAssetScanEstimationTTLSeconds)
	}

	assetScanEstimationPatch := models.AssetScanEstimation{
		StartTime:               utils.PointerTo(startTime),
		EndTime:                 utils.PointerTo(endTime),
		DeleteAfter:             utils.PointerTo(endTime.Add(time.Duration(*assetScanEstimation.TTLSecondsAfterFinished) * time.Second)),
		State:                   assetScanEstimation.State,
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
func (w *Watcher) reconcileAborted(ctx context.Context, assetScanEstimation *models.AssetScanEstimation) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	assetScanEstimationID, ok := assetScanEstimation.GetID()
	if !ok {
		return errors.New("invalid AssetScanEstimation: ID is nil")
	}

	if assetScanEstimation.State == nil || assetScanEstimation.State.State == nil {
		return errors.New("invalid AssetScanEstimation: State is nil")
	}

	now := time.Now()

	assetScanEstimation.State.State = utils.PointerTo(models.AssetScanEstimationStateStateFailed)
	assetScanEstimation.State.LastTransitionTime = utils.PointerTo(now)
	assetScanEstimation.State.StateMessage = utils.PointerTo("asset scan estimation was aborted")
	assetScanEstimation.State.StateReason = utils.PointerTo(models.AssetScanEstimationStateStateReasonAborted)

	assetScanEstimationPatch := models.AssetScanEstimation{
		State: assetScanEstimation.State,
	}
	err := w.backend.PatchAssetScanEstimation(ctx, assetScanEstimationPatch, assetScanEstimationID)
	if err != nil {
		return fmt.Errorf("failed to update AssetScanEstimation. AssetScanEstimation=%s: %w", assetScanEstimationID, err)
	}

	logger.Infof("AssetScanEstimation successfully aborted. AssetScanEstimationID=%v", assetScanEstimationID)

	return nil
}
