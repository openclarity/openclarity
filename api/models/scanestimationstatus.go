package models

import (
	"fmt"
	"time"
)

var scanEstimationStatusStateTransitions = map[ScanEstimationStatusState][]ScanEstimationStatusState{
	ScanEstimationStatusStatePending: {
		ScanEstimationStatusStateDiscovered,
		ScanEstimationStatusStateAborted,
		ScanEstimationStatusStateFailed,
		ScanEstimationStatusStateDone,
	},
	ScanEstimationStatusStateDiscovered: {
		ScanEstimationStatusStateInProgress,
		ScanEstimationStatusStateAborted,
		ScanEstimationStatusStateFailed,
	},
	ScanEstimationStatusStateInProgress: {
		ScanEstimationStatusStateAborted,
		ScanEstimationStatusStateFailed,
		ScanEstimationStatusStateDone,
	},
	ScanEstimationStatusStateAborted: {
		ScanEstimationStatusStateFailed,
	},
}

var scanEstimationStatusReasonMapping = map[ScanEstimationStatusState][]ScanEstimationStatusReason{
	ScanEstimationStatusStatePending: {
		ScanEstimationStatusReasonCreated,
	},
	ScanEstimationStatusStateDiscovered: {
		ScanEstimationStatusReasonSuccessfulDiscovery,
	},
	ScanEstimationStatusStateInProgress: {
		ScanEstimationStatusReasonRunning,
	},
	ScanEstimationStatusStateAborted: {
		ScanEstimationStatusReasonCancellation,
	},
	ScanEstimationStatusStateFailed: {
		ScanEstimationStatusReasonAborted,
		ScanEstimationStatusReasonError,
		ScanEstimationStatusReasonTimeout,
	},
	ScanEstimationStatusStateDone: {
		ScanEstimationStatusReasonNothingToEstimate,
		ScanEstimationStatusReasonSuccess,
	},
}

func NewScanEstimationStatus(s ScanEstimationStatusState, r ScanEstimationStatusReason, m *string) *ScanEstimationStatus {
	return &ScanEstimationStatus{
		State:              s,
		Reason:             r,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}

func (a *ScanEstimationStatus) Equals(b ScanEstimationStatus) bool {
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

func (a *ScanEstimationStatus) isValidStatusTransition(b ScanEstimationStatus) error {
	transitions := scanEstimationStatusStateTransitions[a.State]
	for _, transition := range transitions {
		if transition == b.State {
			return nil
		}
	}

	return fmt.Errorf("invalid transition: from=%s to=%s", a.State, b.State)
}

func (a *ScanEstimationStatus) isValidReason() error {
	reasons := scanEstimationStatusReasonMapping[a.State]
	for _, reason := range reasons {
		if reason == a.Reason {
			return nil
		}
	}

	return fmt.Errorf("invalid reason for state: state=%s reason=%s", a.State, a.Reason)
}

func (a *ScanEstimationStatus) IsValidTransition(b ScanEstimationStatus) error {
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
