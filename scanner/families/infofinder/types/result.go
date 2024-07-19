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
	Metadata families.ScanMetadata `json:"Metadata"`
	Infos    []FlattenedInfo       `json:"Infos"`
}

func NewResult() *Result {
	return &Result{
		Infos: []FlattenedInfo{},
	}
}

func (r *Result) Merge(meta families.ScanInputMetadata, infos []Info) {
	r.Metadata = r.Metadata.Merge(meta)

	for i := range infos {
		r.Infos = append(r.Infos, FlattenedInfo{
			Info:        infos[i],
			ScannerName: meta.ScannerName,
		})
	}
}
