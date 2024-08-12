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

func (rs *ResourceCleanupStatus) Equals(r *ResourceCleanupStatus) bool {
	if rs.Message == nil && r.Message != nil {
		return false
	}
	if r.Message == nil && rs.Message != nil {
		return false
	}
	if rs.Message == nil && r.Message == nil {
		return rs.State == r.State && rs.Reason == r.Reason
	}

	return rs.State == r.State && rs.Reason == r.Reason && *rs.Message == *r.Message
}

func (rs *ResourceCleanupStatus) isValidStatusTransition(r *ResourceCleanupStatus) error {
	transitions, ok := resourceCleanupStatusStateTransitions[rs.State]
	if ok {
		for _, transition := range transitions {
			if transition == r.State {
				return nil
			}
		}
	}

	return fmt.Errorf("invalid transition: from=%s to=%s", rs.State, r.State)
}

func (rs *ResourceCleanupStatus) isValidReason(r *ResourceCleanupStatus) error {
	reasons, ok := resourceCleanupStatusReasonMapping[r.State]
	if ok {
		for _, reason := range reasons {
			if reason == r.Reason {
				return nil
			}
		}
	}

	return fmt.Errorf("invalid reason for state: state=%s reason=%s", r.State, r.Reason)
}

func (rs *ResourceCleanupStatus) IsValidTransition(r *ResourceCleanupStatus) error {
	if rs.Equals(r) {
		return nil
	}

	if err := rs.isValidStatusTransition(r); err != nil {
		return err
	}
	if err := rs.isValidReason(r); err != nil {
		return err
	}

	return nil
}
