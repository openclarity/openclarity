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

package aws

const (
	maxResults = 50
)

type ScanScope struct {
	AllRegions  bool
	Regions     []Region
	ScanStopped bool
	// Only targets that have these tags will be selected for scanning within the selected scan scope.
	// Multiple tags will be treated as an AND operator.
	TagSelector []Tag
	// Targets that have these tags will be excluded from the scan, even if they match the tag selector.
	// Multiple tags will be treated as an AND operator.
	ExcludeTags []Tag
}

type Tag struct {
	Key string
	Val string
}

type SecurityGroup struct {
	id string
}

type VPC struct {
	id             string
	securityGroups []SecurityGroup
}

type Region struct {
	name string
	vpcs []VPC
}
