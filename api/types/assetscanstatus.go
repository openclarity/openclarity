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

package types

import (
	"fmt"
	"time"
)

var assetScanStatusStateTransitions = map[AssetScanStatusState][]AssetScanStatusState{
	AssetScanStatusStatePending: {
		AssetScanStatusStateScheduled,
		AssetScanStatusStateAborted,
	},
	AssetScanStatusStateScheduled: {
		AssetScanStatusStateReadyToScan,
		AssetScanStatusStateFailed,
		AssetScanStatusStateAborted,
	},
	AssetScanStatusStateReadyToScan: {
		AssetScanStatusStateInProgress,
		AssetScanStatusStateFailed,
		AssetScanStatusStateAborted,
	},
	AssetScanStatusStateInProgress: {
		AssetScanStatusStateDone,
		AssetScanStatusStateFailed,
		AssetScanStatusStateAborted,
	},
	AssetScanStatusStateAborted: {
		AssetScanStatusStateFailed,
	},
}

var assetScanStatusReasonMapping = map[AssetScanStatusState][]AssetScanStatusReason{
	AssetScanStatusStatePending: {
		AssetScanStatusReasonCreated,
	},
	AssetScanStatusStateScheduled: {
		AssetScanStatusReasonProvisioning,
	},
	AssetScanStatusStateReadyToScan: {
		AssetScanStatusReasonUnSupervised,
		AssetScanStatusReasonResourcesReady,
	},
	AssetScanStatusStateInProgress: {
		AssetScanStatusReasonScannerIsRunning,
	},
	AssetScanStatusStateAborted: {
		AssetScanStatusReasonCancellation,
	},
	AssetScanStatusStateFailed: {
		AssetScanStatusReasonError,
		AssetScanStatusReasonAbortTimeout,
	},
	AssetScanStatusStateDone: {
		AssetScanStatusReasonSuccess,
	},
}

func NewAssetScanStatus(s AssetScanStatusState, r AssetScanStatusReason, m *string) *AssetScanStatus {
	return &AssetScanStatus{
		State:              s,
		Reason:             r,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}

func (a *AssetScanStatus) Equals(b AssetScanStatus) bool {
	if a.Message == nil && b.Message != nil {
		return false
	}
	if b.Message == nil && a.Message != nil {
		return false
	}
	if a.Message == nil && b.Message == nil {
		return a.State == b.State && a.Reason == b.Reason
	}

	return a.State == b.State && a.Reason == b.Reason && *a.Message == *b.Message
}

func (a *AssetScanStatus) isValidStatusTransition(b AssetScanStatus) error {
	transitions := assetScanStatusStateTransitions[a.State]
	for _, transition := range transitions {
		if transition == b.State {
			return nil
		}
	}

	return fmt.Errorf("invalid transition: from=%s to=%s", a.State, b.State)
}

func (a *AssetScanStatus) isValidReason() error {
	reasons := assetScanStatusReasonMapping[a.State]
	for _, reason := range reasons {
		if reason == a.Reason {
			return nil
		}
	}

	return fmt.Errorf("invalid reason for state: state=%s reason=%s", a.State, a.Reason)
}

func (a *AssetScanStatus) IsValidTransition(b AssetScanStatus) error {
	if a.Equals(b) {
		return nil
	}

	if err := b.isValidReason(); err != nil {
		return err
	}

	if err := a.isValidStatusTransition(b); err != nil {
		return err
	}

	return nil
}
