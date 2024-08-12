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

package aws

import (
	"github.com/openclarity/vmclarity/api/models"
)

const (
	maxResults = 500
)

const (
	EC2TagKeyOwner         = "Owner"
	EC2TagKeyName          = "Name"
	EC2TagValueNamePattern = "vmclarity-scanner-%s"
	EC2TagValueOwner       = "VMClarity"
	EC2TagKeyScanID        = "VMClarity.ScanID"
	EC2TagKeyAssetScanID   = "VMClarity.AssetScanID"
	EC2TagKeyAssetID       = "VMClarity.AssetID"
	EC2TagKeyAssetVolumeID = "VMClarity.AssetVolumeID"

	EC2SnapshotDescription = "Volume snapshot created by VMClarity for scanning"
)

const (
	VpcIDFilterName           = "vpc-id"
	SecurityGroupIDFilterName = "instance.group-id"
	InstanceStateFilterName   = "instance-state-name"
	SnapshotIDFilterName      = "snapshot-id"
)

type ScanScope struct {
	AllRegions  bool
	Regions     []Region
	ScanStopped bool
	// Only assets that have these tags will be selected for scanning within the selected scan scope.
	// Multiple tags will be treated as an AND operator.
	TagSelector []models.Tag
	// Assets that have these tags will be excluded from the scan, even if they match the tag selector.
	// Multiple tags will be treated as an AND operator.
	ExcludeTags []models.Tag
}

type VPC struct {
	ID             string
	SecurityGroups []models.SecurityGroup
}

type Region struct {
	Name string
	VPCs []VPC
}
