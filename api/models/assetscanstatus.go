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
