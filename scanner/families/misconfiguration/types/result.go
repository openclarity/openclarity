// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

import (
	"github.com/openclarity/vmclarity/scanner/families"
)

type Result struct {
	Metadata          families.ScanMetadata       `json:"Metadata"`
	Misconfigurations []FlattenedMisconfiguration `json:"Misconfigurations"`
}

func NewResult() *Result {
	return &Result{
		Metadata:          families.ScanMetadata{},
		Misconfigurations: []FlattenedMisconfiguration{},
	}
}

func (r *Result) Merge(meta families.ScanInputMetadata, misconfigurations []Misconfiguration) {
	for i := range misconfigurations {
		r.Misconfigurations = append(r.Misconfigurations, FlattenedMisconfiguration{
			Misconfiguration: misconfigurations[i],
			ScannerName:      meta.ScannerName,
		})
	}

	// Update metadata
	r.Metadata.Inputs = append(r.Metadata.Inputs, meta)
	r.Metadata.TotalFindings = len(r.Misconfigurations)
}
