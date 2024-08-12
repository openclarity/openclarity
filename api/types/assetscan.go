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

package types

func (r *AssetScan) GetStatus() (AssetScanStatus, bool) {
	if r.Status == nil {
		return AssetScanStatus{}, false
	}

	return *r.Status, true
}

func (r *AssetScan) GetID() (string, bool) {
	var id string
	var ok bool

	if r.Id != nil {
		id, ok = *r.Id, true
	}

	return id, ok
}

func (r *AssetScan) GetScanID() (string, bool) {
	var scanID string
	var ok bool

	if r.Scan != nil {
		scanID, ok = r.Scan.Id, true
	}

	return scanID, ok
}

func (r *AssetScan) GetAssetID() (string, bool) {
	var assetID string
	var ok bool

	if r.Asset != nil {
		assetID, ok = r.Asset.Id, true
	}

	return assetID, ok
}

func (r *AssetScan) GetResourceCleanupStatus() (ResourceCleanupStatus, bool) {
	if r.ResourceCleanupStatus == nil {
		return ResourceCleanupStatus{}, false
	}

	return *r.ResourceCleanupStatus, true
}
