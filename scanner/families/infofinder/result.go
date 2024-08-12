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

package infofinder

import (
	"time"

	"github.com/openclarity/vmclarity/scanner/families/infofinder/types"
	familiestypes "github.com/openclarity/vmclarity/scanner/families/types"
)

type FlattenedInfos struct {
	ScannerName string `json:"ScannerName"`
	types.Info
}

type Results struct {
	Metadata familiestypes.Metadata `json:"Metadata"`
	Infos    []FlattenedInfos       `json:"Infos"`
}

func NewResults() *Results {
	return &Results{
		Metadata: familiestypes.Metadata{
			Timestamp: time.Now(),
			Scanners:  []string{},
		},
		Infos: []FlattenedInfos{},
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

func (r *Results) AddScannerResult(scannerResult *types.ScannerResult) {
	r.addScannerNameToMetadata(scannerResult.ScannerName)

	for _, info := range scannerResult.Infos {
		r.Infos = append(r.Infos, FlattenedInfos{
			ScannerName: scannerResult.ScannerName,
			Info:        info,
		})
	}

	// bump the timestamp as there are new results
	r.Metadata.Timestamp = time.Now()
}
