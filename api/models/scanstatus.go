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

func (s *AssetScanStatus) GetGeneralState() (AssetScanStateState, bool) {
	var state AssetScanStateState
	var ok bool

	if s.General != nil {
		state, ok = s.General.GetState()
	}

	return state, ok
}

func (s *AssetScanStatus) GetGeneralErrors() []string {
	var errs []string

	if s.General != nil {
		errs = s.General.GetErrors()
	}

	return errs
}
