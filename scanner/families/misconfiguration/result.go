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

package misconfiguration

import (
	"time"

	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/types"
	familiestypes "github.com/openclarity/vmclarity/scanner/families/types"
)

type FlattenedMisconfiguration struct {
	ScannerName string `json:"ScannerName"`
	types.Misconfiguration
}

type Results struct {
	Metadata          familiestypes.Metadata      `json:"Metadata"`
	Misconfigurations []FlattenedMisconfiguration `json:"Misconfigurations"`
}

func NewResults() *Results {
	return &Results{
		Metadata: familiestypes.Metadata{
			Timestamp: time.Now(),
			Scanners:  []string{},
		},
		Misconfigurations: []FlattenedMisconfiguration{},
	}
}

func (*Results) IsResults() {}

func (r *Results) addScannerNameToMetadata(name string) {
	for _, scannerName := range r.Metadata.Scanners {
		if scannerName == name {
			return
		}
	}
	r.Metadata.Scanners = append(r.Metadata.Scanners, name)
}

func (r *Results) AddScannerResult(scannerResult types.ScannerResult) {
	r.addScannerNameToMetadata(scannerResult.ScannerName)

	for _, misconfiguration := range scannerResult.Misconfigurations {
		r.Misconfigurations = append(r.Misconfigurations, FlattenedMisconfiguration{
			ScannerName:      scannerResult.ScannerName,
			Misconfiguration: misconfiguration,
		})
	}

	// bump the timestamp as there are new results
	r.Metadata.Timestamp = time.Now()
}
