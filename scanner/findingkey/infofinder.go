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

type InfoFinderKey struct {
	ScannerName string
	Type        string
	Data        string
	Path        string
}

func (k InfoFinderKey) String() string {
	return fmt.Sprintf("%s.%s.%s.%s", k.ScannerName, k.Type, k.Data, k.Path)
}

// Filter returns a string that can be used to filter the info finder finding in the database.
func (k InfoFinderKey) Filter() string {
	return fmt.Sprintf(
		"findingInfo/scannerName eq '%s' and findingInfo/type eq '%s' and findingInfo/data eq '%s' and findingInfo/path eq '%s'",
		k.ScannerName, k.Type, k.Data, k.Path,
	)
}

func GenerateInfoFinderKey(info apitypes.InfoFinderFindingInfo) InfoFinderKey {
	return InfoFinderKey{
		ScannerName: *info.ScannerName,
		Type:        string(*info.Type),
		Data:        *info.Data,
		Path:        *info.Path,
	}
}
