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
	case "x86_64", "x64", "amd64":
		arch = Amd64
	case "arm64", "aarch64":
		arch = Arm64
	default:
		return fmt.Errorf("failed to unmarshal text into VMInfoArchitecture: %s", text)
	}

	*a = arch

	return nil
}

func (a *VMInfoArchitecture) MarshalText() (string, error) {
	switch *a {
	case Amd64:
		return "x86_64", nil
	case Arm64:
		return "arm64", nil
	case Unknown:
		return "", fmt.Errorf("unknown VMInfoArchitecture: %v", *a)
	default:
		return "", fmt.Errorf("failed to marshal VMInfoArchitecture into text: %v", *a)
	}
}

type FromArchitectureMapping map[string]string

func (m *FromArchitectureMapping) UnmarshalText(text []byte) error {
	mapping := make(FromArchitectureMapping)
	items := strings.Split(string(text), ",")

	numOfParts := 2
	for _, item := range items {
		pair := strings.Split(item, ":")
		if len(pair) != numOfParts {
			continue
		}

		mapping[pair[0]] = pair[1]
	}
	*m = mapping

	return nil
}
