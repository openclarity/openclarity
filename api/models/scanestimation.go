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

package models

import "time"

func (s *ScanEstimation) GetState() (ScanEstimationStateState, bool) {
	var state ScanEstimationStateState
	var ok bool

	if s.State != nil && s.State.State != nil {
		state, ok = *s.State.State, true
	}

	return state, ok
}

func (s *ScanEstimation) GetID() (string, bool) {
	var id string
	var ok bool

	if s.Id != nil {
		id, ok = *s.Id, true
	}

	return id, ok
}

func (s *ScanEstimation) IsTimedOut(defaultTimeout time.Duration) bool {
	if s == nil || s.StartTime == nil {
		return false
	}
	// Use the provided timeout to calculate the timeoutTime by default.
	timeoutTime := s.StartTime.Add(defaultTimeout)

	return time.Now().After(timeoutTime)
}

func (s *ScanEstimation) GetScope() (string, bool) {
	var scope string
	var ok bool

	if s.ScanTemplate != nil && s.ScanTemplate.Scope != nil {
		scope, ok = *s.ScanTemplate.Scope, true
	}
	return scope, ok
}
