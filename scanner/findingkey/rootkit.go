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
	"github.com/openclarity/vmclarity/core/to"
)

type RootkitKey struct {
	Name        string
	RootkitType string
	Message     string
}

// String returns an unique string representation of the rootkit finding.
func (k RootkitKey) String() string {
	return fmt.Sprintf("%s.%s.%s", k.Name, k.RootkitType, k.Message)
}

// RootkitString returns an unique string representation of the rootkit independent of
// where the rootkit finding was found by the scanner.
func (k RootkitKey) RootkitString() string {
	return k.String()
}

// Filter returns a string that can be used to filter the rootkit finding in the database.
func (k RootkitKey) Filter() string {
	return fmt.Sprintf(
		"findingInfo/rootkitName eq '%s' and findingInfo/rootkitType eq '%s' and findingInfo/message eq '%s'",
		k.Name, k.RootkitType, k.Message,
	)
}

func GenerateRootkitKey(info apitypes.RootkitFindingInfo) RootkitKey {
	return RootkitKey{
		Name:        *info.RootkitName,
		RootkitType: string(*info.RootkitType),
		Message:     to.ValueOrZero(info.Message),
	}
}
