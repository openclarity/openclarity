package models

import (
	"fmt"
	"time"
)

var assetScanStatusStateTransitions = map[AssetScanStatusState][]AssetScanStatusState{
	AssetScanStatusStatePending: {
		AssetScanStatusStateScheduled,
	},
	AssetScanStatusStateScheduled: {
		AssetScanStatusStateReadyToScan,
		AssetScanStatusStateFailed,
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

func (as *AssetScanStatus) Equals(a *AssetScanStatus) bool {
	if as.Message == nil && a.Message != nil {
		return false
	}
	if a.Message == nil && as.Message != nil {
		return false
	}
	if as.Message == nil && a.Message == nil {
		return as.State == a.State && as.Reason == a.Reason
	}

	return as.State == a.State && as.Reason == a.Reason && *as.Message == *a.Message
}

func (as *AssetScanStatus) isValidStatusTransition(a *AssetScanStatus) error {
	transitions, ok := assetScanStatusStateTransitions[as.State]
	if ok {
		for _, transition := range transitions {
			if transition == a.State {
				return nil
			}
		}
	}

	return fmt.Errorf("invalid transition: from=%s to=%s", as.State, a.State)
}

func (as *AssetScanStatus) isValidReason(a *AssetScanStatus) error {
	reasons, ok := assetScanStatusReasonMapping[a.State]
	if ok {
		for _, reason := range reasons {
			if reason == a.Reason {
				return nil
			}
		}
	}

	return fmt.Errorf("invalid reason for state: state=%s reason=%s", a.State, a.Reason)
}

func (as *AssetScanStatus) IsValidTransition(r *AssetScanStatus) error {
	if as.Equals(r) {
		return nil
	}

	if err := as.isValidStatusTransition(r); err != nil {
		return err
	}
	if err := as.isValidReason(r); err != nil {
		return err
	}

	return nil
}
