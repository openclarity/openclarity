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

package scanresultwatcher

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/api/models"
)

type ScanResultReconcileEvent struct {
	ScanResultID models.ScanResultID
	ScanID       models.ScanID
	TargetID     models.TargetID
}

func (e *ScanResultReconcileEvent) ToFields() log.Fields {
	return log.Fields{
		"ScanResultID": e.ScanResultID,
		"ScanID":       e.ScanID,
		"TargetID":     e.TargetID,
	}
}

func (e ScanResultReconcileEvent) String() string {
	return fmt.Sprintf("ScanResultID=%s ScanID=%s TargetID=%s", e.ScanResultID, e.ScanID, e.TargetID)
}
