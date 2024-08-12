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

func (r *AssetScanEstimation) GetState() (AssetScanEstimationStateState, bool) {
	var state AssetScanEstimationStateState
	var ok bool

	if r.State != nil {
		state, ok = r.State.GetState()
	}

	return state, ok
}

func (r *AssetScanEstimation) GetID() (string, bool) {
	var id string
	var ok bool

	if r.Id != nil {
		id, ok = *r.Id, true
	}

	return id, ok
}

func (r *AssetScanEstimation) GetScanEstimationID() (string, bool) {
	var scanEstimationID string
	var ok bool

	if r.ScanEstimation != nil && r.ScanEstimation.Id != nil {
		scanEstimationID, ok = *r.ScanEstimation.Id, true
	}

	return scanEstimationID, ok
}

func (r *AssetScanEstimation) GetAssetID() (string, bool) {
	var assetID string
	var ok bool

	if r.Asset != nil {
		assetID, ok = r.Asset.Id, true
	}

	return assetID, ok
}

func (s *AssetScanEstimationState) GetState() (AssetScanEstimationStateState, bool) {
	var state AssetScanEstimationStateState
	var ok bool

	if s.State != nil {
		state, ok = *s.State, true
	}

	return state, ok
}
