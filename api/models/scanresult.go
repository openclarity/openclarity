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

func (r *TargetScanResult) GetGeneralState() (TargetScanStateState, bool) {
	var state TargetScanStateState
	var ok bool

	if r.Status != nil {
		state, ok = r.Status.GetGeneralState()
	}

	return state, ok
}

func (r *TargetScanResult) GetGeneralErrors() []string {
	var errs []string

	if r.Status != nil {
		errs = r.Status.GetGeneralErrors()
	}

	return errs
}

func (r *TargetScanResult) GetID() (string, bool) {
	var id string
	var ok bool

	if r.Id != nil {
		id, ok = *r.Id, true
	}

	return id, ok
}

func (r *TargetScanResult) GetScanID() (string, bool) {
	var scanID string
	var ok bool

	if r.Scan != nil {
		scanID, ok = r.Scan.Id, true
	}

	return scanID, ok
}

func (r *TargetScanResult) GetTargetID() (string, bool) {
	var targetID string
	var ok bool

	if r.Target != nil {
		targetID, ok = r.Target.Id, true
	}

	return targetID, ok
}

func (r *TargetScanResult) IsDone() (bool, bool) {
	var done bool
	var ok bool
	var state TargetScanStateState

	if state, ok = r.GetGeneralState(); ok && state == TargetScanStateStateDONE {
		done = true
	}

	return done, ok
}

func (r *TargetScanResult) HasErrors() bool {
	var has bool

	if errs := r.GetGeneralErrors(); len(errs) > 0 {
		has = true
	}

	return has
}
