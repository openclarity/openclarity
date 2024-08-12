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

package results

import (
	"fmt"
)

// Results store slice of results from all families.
type Results struct {
	results []any
}

func New() *Results {
	return &Results{}
}

func (r *Results) SetResults(result any) {
	r.results = append(r.results, result)
}

// GetResult returns results for a specific family from the given results slice.
func GetResult[familyType any](r *Results) (familyType, error) {
	for _, result := range r.results {
		res, ok := result.(familyType)
		if ok {
			return res, nil
		}
	}
	var res familyType
	return res, fmt.Errorf("missing result")
}
