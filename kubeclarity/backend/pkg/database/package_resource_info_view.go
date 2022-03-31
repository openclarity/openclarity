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

package database

const (
	packageResourcesInfoView = "package_resources_info_view"

	// NOTE: when changing one of the column names change also the gorm label in PackageResourcesInfoView.
	columnPackageResourcesInfoViewPackageID    = columnJoinTablePackageID
	columnPackageResourcesInfoViewAnalyzers    = "analyzers"
	columnPackageResourcesInfoViewResourceName = "resource_name"
	columnPackageResourcesInfoViewResourceHash = "resource_hash"
)

type PackageResourcesInfoView struct {
	ResourcePackages
	ResourceName string `json:"resource_name,omitempty" gorm:"column:resource_name"`
	ResourceHash string `json:"resource_hash,omitempty" gorm:"column:resource_hash"`
}

func (PackageResourcesInfoView) TableName() string {
	return packageResourcesInfoView
}
