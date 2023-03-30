// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/types"
)

type Presenter struct {
	mergedResults *scanner.MergedResults
}

// NewPresenter is a *Presenter constructor.
func NewPresenter(mergedResults *scanner.MergedResults) *Presenter {
	return &Presenter{
		mergedResults: mergedResults,
	}
}

type results struct {
	Vulnerabilities [][]scanner.MergedVulnerability `json:"vulnerabilities"`
	Source          types.Source                    `json:"source"`
}

// Present creates a JSON-based reporting.
func (pres *Presenter) Present(output io.Writer) error {
	results := results{
		Vulnerabilities: pres.mergedResults.ToSlice(),
		Source:          pres.mergedResults.Source,
	}

	enc := json.NewEncoder(output)
	// prevent > and < from being escaped in the payload
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(&results); err != nil {
		return fmt.Errorf("failed to encode results: %v", err)
	}

	return nil
}
