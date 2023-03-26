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
	"github.com/openclarity/vmclarity/runtime_scan/pkg/utils"
)

func (srp *ScanResultProcessor) newerExistingFindingTime(ctx context.Context, targetID string, findingType string, completedTime time.Time) (bool, time.Time, error) {
	var found bool
	var newerTime time.Time
	// ScanResults can be processed out of chronological order:
	//
	// If multiple scans of the same target complete, A first then B, we'll
	// pick up the event for A, and then B. If while reconciling A we hit a
	// failure (timeout or weird glitch), the reconciler will continue on
	// and try to reconcile B. It will then pick up A on the next poll and
	// re-reconcile it.
	//
	// So we need to check if any existing findings of this type exist with
	// a foundOn time newer than this results completed time. If there are
	// any newer results it means that this scan result's findings have
	// already been invalidated by a newer scan. We'll find the oldest
	// newer scan, and use its FoundOn time as the InvalidatedOn time for
	// this scan.
	newerFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: utils.PointerTo(fmt.Sprintf(
			"findingInfo/objectType eq '%s' and asset/id eq '%s' and foundOn gt %s",
			findingType, targetID, completedTime.Format(time.RFC3339))),
		OrderBy: utils.PointerTo("foundOn asc"),
		Top:     utils.PointerTo(1), // because of the ordering we only need to get one result here and it'll be the oldest finding which matches the filter
	})
	if err != nil {
		return found, newerTime, fmt.Errorf("failed to check for newer findings: %w", err)
	}

	found = len(*newerFindings.Items) > 0
	if found {
		newerTime = *(*newerFindings.Items)[0].FoundOn
	}

	return found, newerTime, nil
}

func (srp *ScanResultProcessor) invalidateOlderFindingsByType(ctx context.Context, findingType string, targetID string, completedTime time.Time) error {
	// Invalidate any findings of this type for this asset where foundOn is
	// older than this scan result, and has not already been invalidated by
	// a scan result older than this scan result.
	findingsToInvalidate, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
		Filter: utils.PointerTo(fmt.Sprintf(
			"findingInfo/objectType eq '%s' and asset/id eq '%s' and foundOn lt %s and (invalidatedOn gt %s or invalidatedOn eq null)",
			findingType, targetID, completedTime.Format(time.RFC3339), completedTime.Format(time.RFC3339))),
	})
	if err != nil {
		return fmt.Errorf("failed to query findings to invalidate: %w", err)
	}

	for _, finding := range *findingsToInvalidate.Items {
		finding.InvalidatedOn = &completedTime

		err := srp.client.PatchFinding(ctx, *finding.Id, finding)
		if err != nil {
			return fmt.Errorf("failed to update existing finding %s: %w", *finding.Id, err)
		}
	}

	return nil
}

func (srp *ScanResultProcessor) getActiveFindingsByType(ctx context.Context, findingType string, targetID string) (int, error) {
	filter := fmt.Sprintf("findingInfo/objectType eq '%s' and asset/id eq '%s' and invalidatedOn eq null",
		findingType, targetID)
	activeFindings, err := srp.client.GetFindings(ctx, models.GetFindingsParams{
		Count:  utils.PointerTo(true),
		Filter: &filter,

		// select the smallest amount of data to return in items, we
		// only care about the count.
		Top:    utils.PointerTo(1),
		Select: utils.PointerTo("id"),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list all active findings: %w", err)
	}
	return *activeFindings.Count, nil
}
