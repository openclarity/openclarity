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

package client

import (
	"fmt"

	"github.com/openclarity/vmclarity/api/types"
)

type AssetConflictError struct {
	ConflictingAsset *types.Asset
	Message          string
}

func (t AssetConflictError) Error() string {
	return fmt.Sprintf("Conflicting Asset Found with ID %s: %s", *t.ConflictingAsset.Id, t.Message)
}

type ScanConflictError struct {
	ConflictingScan *types.Scan
	Message         string
}

func (t ScanConflictError) Error() string {
	return fmt.Sprintf("Conflicting Scan Found with ID %s: %s", *t.ConflictingScan.Id, t.Message)
}

type AssetScanConflictError struct {
	ConflictingAssetScan *types.AssetScan
	Message              string
}

type ScanConfigConflictError struct {
	ConflictingScanConfig *types.ScanConfig
	Message               string
}

func (t ScanConfigConflictError) Error() string {
	return fmt.Sprintf("Conflicting Scan Config Found with ID %s: %s", *t.ConflictingScanConfig.Id, t.Message)
}

func (t AssetScanConflictError) Error() string {
	return fmt.Sprintf("Conflicting AssetScan Found with ID %s: %s", *t.ConflictingAssetScan.Id, t.Message)
}

type FindingConflictError struct {
	ConflictingFinding *types.Finding
	Message            string
}

func (t FindingConflictError) Error() string {
	return fmt.Sprintf("Conflicting Finding Found with ID %s: %s", *t.ConflictingFinding.Id, t.Message)
}

type AssetScanEstimationConflictError struct {
	ConflictingAssetScanEstimation *types.AssetScanEstimation
	Message                        string
}

func (t AssetScanEstimationConflictError) Error() string {
	return fmt.Sprintf("Conflicting AssetScanEstimation Found with ID %s: %s", *t.ConflictingAssetScanEstimation.Id, t.Message)
}

type ProviderConflictError struct {
	ConflictingProvider *types.Provider
	Message             string
}

func (t ProviderConflictError) Error() string {
	return fmt.Sprintf("Conflicting Provider Found with ID %s: %s", *t.ConflictingProvider.Id, t.Message)
}

type AssetFindingConflictError struct {
	ConflictingAssetFinding *types.AssetFinding
	Message                 string
}

func (t AssetFindingConflictError) Error() string {
	return fmt.Sprintf("Conflicting Asset Finding Found with ID %s: %s", *t.ConflictingAssetFinding.Id, t.Message)
}
