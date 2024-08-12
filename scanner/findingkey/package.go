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

package findingkey

import (
	"fmt"

	apitypes "github.com/openclarity/vmclarity/api/types"
)

type PackageKey struct {
	PackageName    string
	PackageVersion string
}

// String returns an unique string representation of the package finding.
func (k PackageKey) String() string {
	return fmt.Sprintf("%s.%s", k.PackageName, k.PackageVersion)
}

// PackageString returns an unique string representation of the package independent of
// where the package finding was found by the scanner.
func (k PackageKey) PackageString() string {
	return k.String()
}

// Filter returns a string that can be used to filter the package finding in the database.
func (k PackageKey) Filter() string {
	return fmt.Sprintf(
		"findingInfo/name eq '%s' and findingInfo/version eq '%s'",
		k.PackageName, k.PackageVersion,
	)
}

func GeneratePackageKey(info apitypes.PackageFindingInfo) PackageKey {
	return PackageKey{
		PackageName:    *info.Name,
		PackageVersion: *info.Version,
	}
}
