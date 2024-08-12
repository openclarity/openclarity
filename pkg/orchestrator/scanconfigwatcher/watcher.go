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

package scanconfigwatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/orchestrator/common"
	"github.com/openclarity/vmclarity/pkg/shared/backendclient"
	"github.com/openclarity/vmclarity/pkg/shared/log"
	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

type (
	ScanConfigQueue      = common.Queue[ScanConfigReconcileEvent]
	ScanConfigPoller     = common.Poller[ScanConfigReconcileEvent]
	ScanConfigReconciler = common.Reconciler[ScanConfigReconcileEvent]
)

func New(c Config) *Watcher {
	return &Watcher{
		backend:          c.Backend,
		pollPeriod:       c.PollPeriod,
		reconcileTimeout: c.ReconcileTimeout,
		queue:            common.NewQueue[ScanConfigReconcileEvent](),
	}
}

type Watcher struct {
	backend          *backendclient.BackendClient
	pollPeriod       time.Duration
	reconcileTimeout time.Duration

	queue *ScanConfigQueue
}

func (w *Watcher) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "ScanConfigWatcher")
	ctx = log.SetLoggerForContext(ctx, logger)

	poller := &ScanConfigPoller{
		PollPeriod: w.pollPeriod,
		Queue:      w.queue,
		GetItems:   w.GetScanConfigs,
	}
	poller.Start(ctx)

	reconciler := &ScanConfigReconciler{
		ReconcileTimeout:  w.reconcileTimeout,
		Queue:             w.queue,
		ReconcileFunction: w.Reconcile,
	}
	reconciler.Start(ctx)
}

func (w *Watcher) GetScanConfigs(ctx context.Context) ([]ScanConfigReconcileEvent, error) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)
	logger.Debugf("Fetching enabled ScanConfigs")

	params := models.GetScanConfigsParams{
		Filter: utils.PointerTo("disabled eq null or disabled eq false"),
		Select: utils.PointerTo("id"),
	}
	scanConfigs, err := w.backend.GetScanConfigs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled ScanConfigs: %w", err)
	}

	switch {
	case scanConfigs.Items == nil && scanConfigs.Count == nil:
		return nil, fmt.Errorf("failed to fetch enabled ScanConfigs: invalid API response: %v", scanConfigs)
	case scanConfigs.Items != nil && len(*scanConfigs.Items) <= 0:
		return nil, nil
	}

	events := make([]ScanConfigReconcileEvent, 0)
	for _, scanConfig := range *scanConfigs.Items {
		scanConfigID, ok := scanConfig.GetID()
		if !ok {
			logger.Warnf("Skipping to invalid ScanConfig: ID is nil: %v", scanConfig)
			continue
		}

		events = append(events, ScanConfigReconcileEvent{
			ScanConfigID: scanConfigID,
		})
	}

	return events, nil
}

// nolint:cyclop
func (w *Watcher) Reconcile(ctx context.Context, event ScanConfigReconcileEvent) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithFields(event.ToFields())
	ctx = log.SetLoggerForContext(ctx, logger)

	scanConfig, err := w.backend.GetScanConfig(ctx, event.ScanConfigID, models.GetScanConfigsScanConfigIDParams{})
	if err != nil || scanConfig == nil {
		return fmt.Errorf("failed to fetch ScanConfig. Event=%s: %w", event, err)
	}

	_, ok := scanConfig.GetID()
	if !ok {
		return fmt.Errorf("invalid ScanConfig: ID is nil. Event=%s", event)
	}

	// nolint:gomnd
	scheduleWindowSize := w.pollPeriod * 2
	scheduleWindow := NewScheduleWindow(time.Now(), scheduleWindowSize)

	scanConfigSchedule, err := NewScanConfigSchedule(scanConfig, scheduleWindow)
	if err != nil {
		return fmt.Errorf("failed to create new ScanConfig schedule: %w", err)
	}

	switch scanConfigSchedule.State {
	case ScheduleStateDisabled:
		logger.Debug("Skipping ScanConfig as it is disabled")
	case ScheduleStateNotDue:
		logger.Debugf("Skipping ScanConfig due to schedule: %s", scanConfigSchedule)
	case ScheduleStateUnscheduled:
		logger.Debug("Disable unscheduled ScanConfig")
		if err = w.reconcileUnscheduled(ctx, scanConfig); err != nil {
			return fmt.Errorf("failed to disable unscheduled ScanConfig: %w", err)
		}
	case ScheduleStateDue:
		logger.Debug("Run new Scan for ScanConfig")
		if err = w.reconcileDue(ctx, scanConfig, scanConfigSchedule); err != nil {
			return fmt.Errorf("failed to create new Scan for ScanConfig: %w", err)
		}
	case ScheduleStateOverdue:
		logger.Debug("Reschedule overdue ScanConfig")
		if err = w.reconcileOverdue(ctx, scanConfig, scanConfigSchedule); err != nil {
			return fmt.Errorf("failed to reschedule ScanConfig: %w", err)
		}
	}

	return nil
}

func (w *Watcher) reconcileUnscheduled(ctx context.Context, scanConfig *models.ScanConfig) error {
	scanConfigPatch := &models.ScanConfig{
		Disabled: utils.PointerTo(true),
	}

	if err := w.backend.PatchScanConfig(ctx, *scanConfig.Id, scanConfigPatch); err != nil {
		return fmt.Errorf("failed to patch ScanConfig. ScanConfigID=%s: %w", *scanConfig.Id, err)
	}

	return nil
}

func (w *Watcher) reconcileDue(ctx context.Context, scanConfig *models.ScanConfig, schedule *ScanConfigSchedule) error {
	if err := w.createScan(ctx, scanConfig); err != nil {
		return fmt.Errorf("failed to reconcile new Scan for ScanConfig. ScanConfigID=%s: %w", *scanConfig.Id, err)
	}
	nextOperationTime := schedule.OperationTime.NextAfter(schedule.Window.Next().Start())
	// FIXME: disable ScanConfig if it was a oneshot
	scanConfigPatch := &models.ScanConfig{
		Scheduled: &models.RuntimeScheduleScanConfig{
			CronLine:      scanConfig.Scheduled.CronLine,
			OperationTime: utils.PointerTo(nextOperationTime.Time()),
		},
	}

	if err := w.backend.PatchScanConfig(ctx, *scanConfig.Id, scanConfigPatch); err != nil {
		return fmt.Errorf("failed to update operation time for ScanConfig. ScanConfigID=%s: %w", *scanConfig.Id, err)
	}

	return nil
}

func (w *Watcher) createScan(ctx context.Context, scanConfig *models.ScanConfig) error {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	filter := fmt.Sprintf("scanConfig/id eq '%s' and state ne '%s' and state ne '%s'", *scanConfig.Id,
		models.ScanStateDone, models.ScanStateFailed)
	scans, err := w.backend.GetScans(ctx, models.GetScansParams{
		Filter: utils.PointerTo(filter),
		Select: utils.PointerTo("id"),
		Count:  utils.PointerTo(true),
	})
	if err != nil || scans == nil {
		return fmt.Errorf("failed to fetch scans for ScanConfig. ScanConfigID=%s: %w", *scanConfig.Id, err)
	}

	var scansInProgress int
	if scans.Count != nil {
		scansInProgress = *scans.Count
	} else {
		scansInProgress = len(*scans.Items)
	}

	if scansInProgress > 0 {
		logger.Warnf("Skipping ScanConfig as it has Scan(s) already in-progress")
		return nil
	}

	scan := newScanFromScanConfig(scanConfig)
	scan.StartTime = utils.PointerTo(time.Now())

	_, err = w.backend.PostScan(ctx, *scan)
	if err != nil {
		var conflictErr backendclient.ScanConflictError
		if errors.As(err, &conflictErr) {
			logger.Debugf("Scan already exist. ScanID=%s", *conflictErr.ConflictingScan.Id)
			return nil
		}

		return fmt.Errorf("failed to create new Scan for ScanConfig. ScanConfigID=%s: %w", *scanConfig.Id, err)
	}

	return nil
}

func (w *Watcher) reconcileOverdue(ctx context.Context, scanConfig *models.ScanConfig, schedule *ScanConfigSchedule) error {
	nextOperationTime := schedule.OperationTime.NextAfter(schedule.Window.Next().Start())

	scanConfigPatch := &models.ScanConfig{
		Scheduled: &models.RuntimeScheduleScanConfig{
			CronLine:      scanConfig.Scheduled.CronLine,
			OperationTime: utils.PointerTo(nextOperationTime.Time()),
		},
	}

	if err := w.backend.PatchScanConfig(ctx, *scanConfig.Id, scanConfigPatch); err != nil {
		return fmt.Errorf("failed to update operation time for ScanConfig with %s id: %w", *scanConfig.Id, err)
	}

	return nil
}
