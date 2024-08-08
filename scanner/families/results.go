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

package families

import (
	"errors"
	"sync"
)

// results stores data from all families. Safe for concurrent usage.
type results struct {
	mu      sync.RWMutex
	results map[FamilyType]any
}

func NewResultStore() ResultStore {
	return &results{
		results: make(map[FamilyType]any),
	}
}

func (r *results) GetAllFamilyResults() []any {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var ret []any
	for _, result := range r.results {
		ret = append(ret, result)
	}

	return ret
}

func (r *results) GetFamilyResult(family FamilyType) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if result, ok := r.results[family]; ok {
		return result, true
	}

	return nil, false
}

func (r *results) SetFamilyResult(family FamilyType, result any) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results[family] = result
}

// GetFamilyResultByType returns results for a specific family from the given
// results but using type instead of interface.
func GetFamilyResultByType[FamilyResultType any](store ResultStore) (FamilyResultType, error) {
	for _, result := range store.GetAllFamilyResults() {
		typedResult, ok := result.(FamilyResultType)
		if ok {
			return typedResult, nil
		}
	}

	var res FamilyResultType
	return res, errors.New("missing result")
}
