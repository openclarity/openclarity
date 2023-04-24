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

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/runtime_scan/pkg/orchestrator/common"
	runtimeScanUtils "github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
	"github.com/openclarity/vmclarity/shared/pkg/backendclient"
)

const (
	DefaultPollInterval     = time.Minute
	DefaultReconcileTimeout = time.Minute
)

type ScanReconcileEvent struct {
	ScanID models.ScanID
}

type (
	ScanQueue      = common.Queue[ScanReconcileEvent]
	ScanPoller     = common.Poller[ScanReconcileEvent]
	ScanReconciler = common.Reconciler[ScanReconcileEvent]
)

type Config struct {
	Backend          *backendclient.BackendClient
	PollPeriod       time.Duration
	ReconcileTimeout time.Duration
}

func New(c Config) *Watcher {
	logger := log.WithFields(log.Fields{"controller": "ScanWatcher"})
	return &Watcher{
		logger,
		c.Backend,
		c.PollPeriod,
		c.ReconcileTimeout,
	}
}

type Watcher struct {
	logger           *log.Entry
	client           *backendclient.BackendClient
	pollPeriod       time.Duration
	reconcileTimeout time.Duration
}

func (w *Watcher) Start(ctx context.Context) {
	queue := common.NewQueue[ScanReconcileEvent]()

	poller := &ScanPoller{
		Logger:     w.logger,
		PollPeriod: w.pollPeriod,
		Queue:      queue,
		GetItems:   w.GetAbortedScans,
	}
	poller.Start(ctx)

	reconciler := &ScanReconciler{
		ReconcileTimeout:  w.reconcileTimeout,
		Queue:             queue,
		ReconcileFunction: w.Reconcile,
	}
	reconciler.Start(ctx)
}

func (w *Watcher) GetAbortedScans(ctx context.Context) ([]ScanReconcileEvent, error) {
	scans, err := w.getScansByState(ctx, models.ScanStateAborted)
	if err != nil || scans.Items == nil || len(*scans.Items) <= 0 {
		return nil, err
	}

	count := len(*scans.Items)
	r := make([]ScanReconcileEvent, count)
	for i, scan := range *scans.Items {
		r[i] = ScanReconcileEvent{
			ScanID: *scan.Id,
		}
	}

	return r, nil
}

func (w *Watcher) Reconcile(ctx context.Context, event ScanReconcileEvent) error {
	w.logger.Infof("Reconciling scan event: %v", event)

	selector := "id,state,stateReason"
	params := models.GetScansScanIDParams{
		Select: &selector,
	}

	scan, err := w.client.GetScan(ctx, event.ScanID, params)
	if err != nil || scan == nil {
		return fmt.Errorf("getting scan with id %s failed: %v", event.ScanID, err)
	}

	state, ok := scan.GetState()
	if !ok {
		return fmt.Errorf("cannot determine state of Scan with %s id", event.ScanID)
	}

	switch state {
	case models.ScanStateDone, models.ScanStateFailed:
		w.logger.Debugf("Reconciling scan event is skipped as Scan is already finished: %v", event)
	case models.ScanStateAborted:
		return w.reconcileAborted(ctx, event)
	case models.ScanStatePending, models.ScanStateDiscovered, models.ScanStateInProgress:
		fallthrough
	default:
	}

	return nil
}

func (w *Watcher) getScansByState(ctx context.Context, s models.ScanState) (models.Scans, error) {
	filter := fmt.Sprintf("state eq '%s'", s)
	selector := "id"
	params := models.GetScansParams{
		Filter: &filter,
		Select: &selector,
	}
	scans, err := w.client.GetScans(ctx, params)
	if err != nil {
		err = fmt.Errorf("getting Scan(s) by their state failed: %v", err)
	}
	if scans == nil {
		scans = &models.Scans{}
	}

	return *scans, err
}

func (w *Watcher) reconcileAborted(ctx context.Context, event ScanReconcileEvent) error {
	filter := fmt.Sprintf("scan/id eq '%s' and status/general/state ne '%s' and status/general/state ne '%s'",
		event.ScanID, models.ABORTED, models.DONE)
	selector := "id,status"
	params := models.GetScanResultsParams{
		Filter: &filter,
		Select: &selector,
	}

	scanResults, err := w.client.GetScanResults(ctx, params)
	if err != nil {
		return fmt.Errorf("getting ScanResult(s) for Scan with %s id failed: %v", event.ScanID, err)
	}

	if scanResults.Items != nil && len(*scanResults.Items) > 0 {
		var reconciliationFailed bool
		var wg sync.WaitGroup

		for _, scanResult := range *scanResults.Items {
			if scanResult.Id == nil {
				continue
			}
			id := *scanResult.Id

			wg.Add(1)
			go func() {
				defer wg.Done()
				sr := models.TargetScanResult{
					Status: &models.TargetScanStatus{
						General: &models.TargetScanState{
							State: runtimeScanUtils.PointerTo(models.ABORTED),
						},
					},
				}

				err := w.client.PatchScanResult(ctx, sr, id)
				if err != nil {
					w.logger.Errorf("Failed to patch ScanResult with id: %s", id)
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

	// FIXME(chrisgacsal): updating the Scan state here collides the Scan logic in Scanner.job_management
	//                     therefore it is disabled until Scan lifecycle management is moved to ScanWatcher.
	// scan := &models.Scan{
	// 	EndTime:     runtimeScanUtils.PointerTo(time.Now().UTC()),
	// 	State:       runtimeScanUtils.PointerTo(models.ScanStateFailed),
	// 	StateReason: runtimeScanUtils.PointerTo(models.ScanStateReasonAborted),
	// }
	//
	// err = w.client.PatchScan(ctx, event.ScanID, scan)
	// if err != nil {
	// 	return fmt.Errorf("failed to patch Scan with id: %s: %v", event.ScanID, err)
	// }

	return nil
}
