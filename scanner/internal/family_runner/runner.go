// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package family_runner // nolint:revive,stylecheck

import (
	"context"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/families"
)

// Runner handles a specific family execution.
type Runner[T any] struct {
	family families.Family[T]
}

func New[T any](family families.Family[T]) *Runner[T] {
	return &Runner[T]{family: family}
}

func (r *Runner[T]) Run(ctx context.Context, notifier families.FamilyNotifier, results *families.Results) []error {
	var errs []error

	// Override context with family params
	familyType := r.family.GetType()
	ctx, logger := log.NewContextLoggerOrDefault(ctx, map[string]interface{}{
		"family": familyType,
	})
	logger.Infof("Running family %q in progress...", familyType)

	// Notify about start, return preemptively if it fails since we won't be able to
	// collect family results anyway.
	if err := notifier.FamilyStarted(ctx, familyType); err != nil {
		errs = append(errs, fmt.Errorf("family %q started notification failed: %w", familyType, err))
		return errs
	}

	// Run family
	startTime := time.Now()
	result, err := r.family.Run(ctx, results)
	familyResult := families.FamilyResult{
		Result:     result,
		FamilyType: familyType,
		Err:        err,
	}

	// Handle family result depending on returned data
	logger.Debugf("Received result from family %q: %v", familyType, familyResult)
	if err != nil {
		logger.WithError(err).Errorf("Family %q finished with error", familyType)

		// Submit run error so that we can check if the error are from the notifier or
		// from the actual family run
		errs = append(errs, &FamilyFailedError{
			Family: familyType,
			Err:    err,
		})
	} else {
		logger.Infof("Family %q finished with success", familyType)

		// Update family result metadata
		if metadata := getFamilyScanMetadata(result); metadata != nil {
			metadata.StartTime = startTime
			metadata.EndTime = time.Now()
		}

		// Set result in shared object for the family
		results.SetFamilyResult(result)
	}

	// Notify about finish
	if err := notifier.FamilyFinished(ctx, familyResult); err != nil {
		errs = append(errs, fmt.Errorf("family %q finished notification failed: %w", familyType, err))
	}

	return errs
}

// FamilyFailedError defines families.Family run fail error.
type FamilyFailedError struct {
	Family families.FamilyType
	Err    error
}

func (e *FamilyFailedError) Error() string {
	return fmt.Sprintf("family %q failed with %v", e.Family, e.Err)
}
