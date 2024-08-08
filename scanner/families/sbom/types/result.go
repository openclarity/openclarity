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
	"fmt"
	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/openclarity/vmclarity/scanner/utils/converter"
)

type Result struct {
	SBOM *cdx.BOM `json:"sbom"`
}

func NewResult(sbom *cdx.BOM) *Result {
	return &Result{
		SBOM: sbom,
	}
}

func (r *Result) EncodeToBytes(outputFormat string) ([]byte, error) {
	f, err := converter.StringToSbomFormat(outputFormat)
	if err != nil {
		return nil, fmt.Errorf("unable to parse output format: %w", err)
	}

	bomBytes, err := converter.CycloneDxToBytes(r.SBOM, f)
	if err != nil {
		return nil, fmt.Errorf("unable to encode results to bytes: %w", err)
	}

	return bomBytes, nil
}
