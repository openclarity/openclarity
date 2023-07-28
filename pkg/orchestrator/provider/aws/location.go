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
	"fmt"
	"strings"
)

const LocationSeparator = "/"

type Location struct {
	Region string
	Vpc    string
}

func (l Location) String() string {
	return fmt.Sprintf("%s%s%s", l.Region, LocationSeparator, l.Vpc)
}

// NOTE: pattern <region>/<vpc>.
func NewLocation(l string) (*Location, error) {
	numOfParts := 2
	s := strings.SplitN(l, LocationSeparator, numOfParts)
	if len(s) != numOfParts {
		return nil, fmt.Errorf("failed to parse Location string: %s", l)
	}

	return &Location{
		Region: s[0],
		Vpc:    s[1],
	}, nil
}
