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

package lsblk

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Bytes int

func (b Bytes) Int() int {
	return int(b)
}

func (b Bytes) String() string {
	return fmt.Sprintf("%d", b)
}

func (b *Bytes) UnmarshalJSON(data []byte) error {
	var value interface{}

	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("failed to unmarshal value: %s", data)
	}

	switch v := value.(type) {
	case int:
		*b = Bytes(v)
	case json.Number:
		i, err := v.Int64()
		if err != nil {
			return fmt.Errorf("failed to unmarshal json.Number value as int64: %w", err)
		}
		*b = Bytes(i)
	case float64:
		*b = Bytes(int(v))
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("failed to unmarshal string value as int: %w", err)
		}
		*b = Bytes(i)
	}

	return nil
}
