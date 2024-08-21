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

import (
	"fmt"
	"strings"
)

func (a *VMInfoArchitecture) UnmarshalText(text []byte) error {
	var arch VMInfoArchitecture

	switch string(text) {
	case "x86_64":
		arch = Amd64
	case "arm64":
		arch = Arm64
	default:
		return fmt.Errorf("failed to unmarshal text into VMInfoArchitecture: %s", text)
	}

	*a = arch

	return nil
}

type FromArchitectureMapping map[VMInfoArchitecture]string

func (m *FromArchitectureMapping) UnmarshalText(text []byte) error {
	mapping := make(FromArchitectureMapping)
	items := strings.Split(string(text), ",")

	numOfParts := 2
	for _, item := range items {
		pair := strings.Split(item, ":")
		if len(pair) != numOfParts {
			continue
		}

		var arch VMInfoArchitecture
		err := arch.UnmarshalText([]byte(pair[0]))
		if err != nil {
			return fmt.Errorf("failed to unmarshal architecture: %w", err)
		}

		mapping[arch] = pair[1]
	}
	*m = mapping

	return nil
}
