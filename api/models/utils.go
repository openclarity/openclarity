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

package models

import "fmt"

// CoalesceComparable will return original if target is not set, and target if
// original is not set. If both are set then they must be the same otherwise an
// error is raised.
func CoalesceComparable[T comparable](original, target T) (T, error) {
	var zero T
	if original != zero && target != zero && original != target {
		return zero, fmt.Errorf("%v does not match %v", original, target)
	}
	if original != zero {
		return original, nil
	}
	return target, nil
}

// UnionSlices returns the union of the input slices.
func UnionSlices[T comparable](inputs ...[]T) []T {
	seen := map[T]struct{}{}
	result := []T{}
	for _, i := range inputs {
		for _, j := range i {
			if _, ok := seen[j]; !ok {
				seen[j] = struct{}{}
				result = append(result, j)
			}
		}
	}
	return result
}
