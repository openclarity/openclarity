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

func (p *CloudProvider) UnmarshalText(text []byte) error {
	var provider CloudProvider

	switch strings.ToLower(string(text)) {
	case strings.ToLower(string(AWS)):
		provider = AWS
	case strings.ToLower(string(Azure)):
		provider = Azure
	case strings.ToLower(string(Docker)):
		provider = Docker
	case strings.ToLower(string(External)):
		provider = External
	case strings.ToLower(string(GCP)):
		provider = GCP
	case strings.ToLower(string(Kubernetes)):
		provider = Kubernetes
	default:
		return fmt.Errorf("failed to unmarshal text into Provider: %s", text)
	}

	*p = provider

	return nil
}
