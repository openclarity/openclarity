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

package assetscanestimation

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
)

type AssetScanEstimationReconcileEvent struct {
	AssetScanEstimationID apitypes.AssetScanEstimationID
	ScanEstimationID      apitypes.ScanEstimationID
	AssetID               apitypes.AssetScanID
}

func (e AssetScanEstimationReconcileEvent) ToFields() log.Fields {
	return log.Fields{
		"AssetScanEstimationID": e.AssetScanEstimationID,
		"ScanEstimationID":      e.ScanEstimationID,
		"AssetID":               e.AssetID,
	}
}

func (e AssetScanEstimationReconcileEvent) String() string {
	return fmt.Sprintf("AssetScanEstimationID=%s ScanEstimationID=%s AssetID=%s", e.AssetScanEstimationID, e.ScanEstimationID, e.AssetID)
}

func (e AssetScanEstimationReconcileEvent) Hash() string {
	return e.AssetScanEstimationID
}
