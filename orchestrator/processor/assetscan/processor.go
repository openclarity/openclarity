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
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	apiclient "github.com/openclarity/vmclarity/api/client"
	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/core/to"
	"github.com/openclarity/vmclarity/orchestrator/common"
)

type AssetScanProcessor struct {
	client           *apiclient.Client
	pollPeriod       time.Duration
	reconcileTimeout time.Duration
}

func New(config Config) *AssetScanProcessor {
	return &AssetScanProcessor{
		client:           config.Client,
		pollPeriod:       config.PollPeriod,
		reconcileTimeout: config.ReconcileTimeout,
	}
}

// Returns true if AssetScanStatus.State is DONE and there are no Errors.
func statusCompletedWithNoErrors(status *apitypes.ScannerStatus) bool {
	return status != nil && status.State == apitypes.ScannerStatusStateDone
}

// nolint:cyclop,gocognit
func (asp *AssetScanProcessor) Reconcile(ctx context.Context, event AssetScanReconcileEvent) error {
	// Get latest information, in case we've been sat in the reconcile
	// queue for a while
	assetScan, err := asp.client.GetAssetScan(ctx, event.AssetScanID, apitypes.GetAssetScansAssetScanIDParams{})
	if err != nil {
		return fmt.Errorf("failed to get asset scan from API: %w", err)
	}

	// Re-check the findingsProcessed boolean, we might have been re-queued
	// while already being reconciled, if so we can short circuit here.
	if assetScan.FindingsProcessed != nil && *assetScan.FindingsProcessed {
		return nil
	}

	newFailedToReconcileTypeError := func(err error, t string) error {
		return fmt.Errorf("failed to reconcile asset scan %s %s to findings: %w", *assetScan.Id, t, err)
	}

	// Process each of the successfully scanned (state DONE and no errors) families into findings.
	if statusCompletedWithNoErrors(assetScan.Vulnerabilities.Status) {
		if err := asp.reconcileResultVulnerabilitiesToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "vulnerabilities")
		}

		// NOTE: vulnerabilities can reference packages which also need to be reconciled
		// to add referenced packages as independent findings
		if err := asp.reconcileResultPackagesToFindings(ctx, assetScan, withVulnerabilityPackageExtractor); err != nil {
			return newFailedToReconcileTypeError(err, "vulnerabilities")
		}
	}

	// NOTE: always reconcile vulnerability findings before package findings to
	// ensure that package finding summaries can be properly updated.
	if statusCompletedWithNoErrors(assetScan.Sbom.Status) {
		if err := asp.reconcileResultPackagesToFindings(ctx, assetScan, withSbomPackageExtractor); err != nil {
			return newFailedToReconcileTypeError(err, "sbom")
		}
	}

	if statusCompletedWithNoErrors(assetScan.Exploits.Status) {
		if err := asp.reconcileResultExploitsToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "exploits")
		}
	}

	if statusCompletedWithNoErrors(assetScan.Secrets.Status) {
		if err := asp.reconcileResultSecretsToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "secrets")
		}
	}

	if statusCompletedWithNoErrors(assetScan.Malware.Status) {
		if err := asp.reconcileResultMalwareToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "malware")
		}
	}

	if statusCompletedWithNoErrors(assetScan.Rootkits.Status) {
		if err := asp.reconcileResultRootkitsToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "rootkits")
		}
	}

	if statusCompletedWithNoErrors(assetScan.Misconfigurations.Status) {
		if err := asp.reconcileResultMisconfigurationsToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "misconfigurations")
		}
	}

	if statusCompletedWithNoErrors(assetScan.InfoFinder.Status) {
		if err := asp.reconcileResultInfoFindersToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "infoFinder")
		}
	}

	if statusCompletedWithNoErrors(assetScan.Plugins.Status) {
		if err := asp.reconcileResultPluginsToFindings(ctx, assetScan); err != nil {
			return newFailedToReconcileTypeError(err, "plugins")
		}
	}

	// Mark post-processing completed for this asset scan
	assetScan.FindingsProcessed = to.Ptr(true)
	err = asp.client.PatchAssetScan(ctx, assetScan, *assetScan.Id)
	if err != nil {
		return fmt.Errorf("failed to update asset scan %s: %w", *assetScan.Id, err)
	}

	return nil
}

type AssetScanReconcileEvent struct {
	AssetScanID apitypes.AssetScanID
}

func (e AssetScanReconcileEvent) ToFields() logrus.Fields {
	return logrus.Fields{
		"AssetScanID": e.AssetScanID,
	}
}

func (e AssetScanReconcileEvent) String() string {
	return "AssetScanID=" + e.AssetScanID
}

func (e AssetScanReconcileEvent) Hash() string {
	return e.AssetScanID
}

func (asp *AssetScanProcessor) GetItems(ctx context.Context) ([]AssetScanReconcileEvent, error) {
	filter := fmt.Sprintf("(status/state eq '%s' or status/state eq '%s') and (findingsProcessed eq false or findingsProcessed eq null)",
		apitypes.AssetScanStatusStateDone, apitypes.AssetScanStatusStateFailed)
	assetScans, err := asp.client.GetAssetScans(ctx, apitypes.GetAssetScansParams{
		Filter: to.Ptr(filter),
		Select: to.Ptr("id"),
	})
	if err != nil {
		return []AssetScanReconcileEvent{}, fmt.Errorf("failed to get asset scans from API: %w", err)
	}

	items := make([]AssetScanReconcileEvent, len(*assetScans.Items))
	for i, res := range *assetScans.Items {
		items[i] = AssetScanReconcileEvent{*res.Id}
	}

	return items, nil
}

func (asp *AssetScanProcessor) Start(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx).WithField("controller", "AssetScanProcessor")
	ctx = log.SetLoggerForContext(ctx, logger)

	queue := common.NewQueue[AssetScanReconcileEvent]()

	poller := common.Poller[AssetScanReconcileEvent]{
		PollPeriod: asp.pollPeriod,
		GetItems:   asp.GetItems,
		Queue:      queue,
	}
	poller.Start(ctx)

	reconciler := common.Reconciler[AssetScanReconcileEvent]{
		ReconcileFunction: asp.Reconcile,
		ReconcileTimeout:  asp.reconcileTimeout,
		Queue:             queue,
	}
	reconciler.Start(ctx)
}
