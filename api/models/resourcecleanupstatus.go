package models

import (
	"fmt"
	"time"
)

var resourceCleanupStatusStateTransitions = map[ResourceCleanupStatusState][]ResourceCleanupStatusState{
	ResourceCleanupStatusStatePending: {
		ResourceCleanupStatusStateSkipped,
		ResourceCleanupStatusStateFailed,
		ResourceCleanupStatusStateDone,
	},
}

var resourceCleanupStatusReasonMapping = map[ResourceCleanupStatusState][]ResourceCleanupStatusReason{
	ResourceCleanupStatusStatePending: {
		ResourceCleanupStatusReasonAssetScanCreated,
	},
	ResourceCleanupStatusStateSkipped: {
		ResourceCleanupStatusReasonDeletePolicy,
		ResourceCleanupStatusReasonNotApplicable,
	},
	ResourceCleanupStatusStateFailed: {
		ResourceCleanupStatusReasonProviderError,
		ResourceCleanupStatusReasonInternalError,
	},
	ResourceCleanupStatusStateDone: {
		ResourceCleanupStatusReasonSuccess,
	},
}

func NewResourceCleanupStatus(s ResourceCleanupStatusState, r ResourceCleanupStatusReason, m *string) *ResourceCleanupStatus {
	return &ResourceCleanupStatus{
		State:              s,
		Reason:             r,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}

func (a *ResourceCleanupStatus) Equals(b ResourceCleanupStatus) bool {
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

func (a *ResourceCleanupStatus) isValidStatusTransition(b ResourceCleanupStatus) error {
	transitions := resourceCleanupStatusStateTransitions[a.State]
	for _, transition := range transitions {
		if transition == b.State {
			return nil
		}
	}

	return fmt.Errorf("invalid transition: from=%s to=%s", a.State, b.State)
}

func (a *ResourceCleanupStatus) isValidReason() error {
	reasons := resourceCleanupStatusReasonMapping[a.State]
	for _, reason := range reasons {
		if reason == a.Reason {
			return nil
		}
	}

	return fmt.Errorf("invalid reason for state: state=%s reason=%s", a.State, a.Reason)
}

func (a *ResourceCleanupStatus) IsValidTransition(b ResourceCleanupStatus) error {
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
