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

package types

func MapToTags(tags map[string]string) *[]Tag {
	ret := make([]Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, Tag{
			Key:   key,
			Value: val,
		})
	}
	return &ret
}

func MergeTags(left, right *[]Tag) *[]Tag {
	if left == nil && right == nil {
		return nil
	}

	merged := &[]Tag{}
	if left == nil || len(*left) == 0 {
		*merged = *right

		return merged
	}

	if right == nil || len(*right) == 0 {
		*merged = *left

		return merged
	}

	m := map[string]string{}
	for _, tag := range *left {
		m[tag.Key] = tag.Value
	}

	for _, tag := range *right {
		m[tag.Key] = tag.Value
	}

	return MapToTags(m)
}
