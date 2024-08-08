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

package scanconfig

import (
	log "github.com/sirupsen/logrus"

	apitypes "github.com/openclarity/vmclarity/api/types"
)

type ScanConfigReconcileEvent struct {
	ScanConfigID apitypes.ScanConfigID
}

func (e ScanConfigReconcileEvent) ToFields() log.Fields {
	return log.Fields{
		"ScanConfigID": e.ScanConfigID,
	}
}

func (e ScanConfigReconcileEvent) String() string {
	return "ScanConfigID=" + e.ScanConfigID
}

func (e ScanConfigReconcileEvent) Hash() string {
	return e.ScanConfigID
}
