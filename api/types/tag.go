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

import "github.com/oapi-codegen/nullable"

func MapToTags(tags map[string]string) nullable.Nullable[[]Tag] {
	ret := make([]Tag, 0, len(tags))
	for key, val := range tags {
		ret = append(ret, Tag{
			Key:   key,
			Value: val,
		})
	}
	return nullable.NewNullableWithValue(ret)
}

func MergeTags(left, right nullable.Nullable[[]Tag]) nullable.Nullable[[]Tag] {
	if left.IsNull() && right.IsNull() {
		return nullable.NewNullNullable[[]Tag]()
	}

	leftValues, err := left.Get()
	if err != nil {
		return right
	}

	rightValues, err := right.Get()
	if err != nil {
		return left
	}

	m := map[string]string{}
	for _, tag := range leftValues {
		m[tag.Key] = tag.Value
	}

	for _, tag := range rightValues {
		m[tag.Key] = tag.Value
	}

	return MapToTags(m)
}
