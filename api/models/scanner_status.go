package models

import "time"

func NewScannerStatus(s ScannerStatusState, r ScannerStatusReason, m *string) *ScannerStatus {
	return &ScannerStatus{
		State:              s,
		Reason:             r,
		Message:            m,
		LastTransitionTime: time.Now(),
	}
}
