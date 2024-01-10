package models

import (
	"fmt"
	"time"
)

var scanStatusStateTransitions = map[ScanStatusState][]ScanStatusState{
	ScanStatusStatePending: {
		ScanStatusStateDiscovered,
		ScanStatusStateAborted,
		ScanStatusStateFailed,
		ScanStatusStateDone,
	},
	ScanStatusStateDiscovered: {
		ScanStatusStateInProgress,
		ScanStatusStateAborted,
	},
	ScanStatusStateInProgress: {
		ScanStatusStateAborted,
		ScanStatusStateFailed,
		ScanStatusStateDone,
	},
	ScanStatusStateAborted: {
		ScanStatusStateFailed,
	},
}

var scanStatusReasonMapping = map[ScanStatusState][]ScanStatusReason{
	ScanStatusStatePending: {
		ScanStatusReasonCreated,
	},
	ScanStatusStateDiscovered: {
		ScanStatusReasonAssetsDiscovered,
	},
	ScanStatusStateInProgress: {
		ScanStatusReasonAssetScansRunning,
	},
	ScanStatusStateAborted: {
		ScanStatusReasonCancellation,
	},
	ScanStatusStateFailed: {
		ScanStatusReasonCancellation,
		ScanStatusReasonError,
		ScanStatusReasonTimeout,
	},
	ScanStatusStateDone: {
		ScanStatusReasonNothingToScan,
		ScanStatusReasonSuccess,
	},
}

func NewScanStatus(s ScanStatusState, r ScanStatusReason, m *string) *ScanStatus {
	return &ScanStatus{
		State:              s,
		Reason:             r,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}

func (a *ScanStatus) Equals(b *ScanStatus) bool {
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

func (a *ScanStatus) isValidStatusTransition(b *ScanStatus) error {
	transitions := scanStatusStateTransitions[a.State]
	for _, transition := range transitions {
		if transition == b.State {
			return nil
		}
	}

	return fmt.Errorf("invalid transition: from=%s to=%s", a.State, b.State)
}

func (a *ScanStatus) isValidReason() error {
	reasons := scanStatusReasonMapping[a.State]
	for _, reason := range reasons {
		if reason == a.Reason {
			return nil
		}
	}

	return fmt.Errorf("invalid reason for state: state=%s reason=%s", a.State, a.Reason)
}

func (a *ScanStatus) IsValidTransition(b *ScanStatus) error {
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
