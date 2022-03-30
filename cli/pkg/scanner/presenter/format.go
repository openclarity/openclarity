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

package presenter

import (
	"strings"
)

const (
	unknownFormat format = "unknown"
	jsonFormat    format = "json"
	tableFormat   format = "table"
)

// format is a dedicated type to represent a specific kind of presenter output format.
type format string

func (f format) String() string {
	return string(f)
}

// getFormat returns the presenter.format specified by the given user input.
func getFormat(userInput string) format {
	switch strings.ToLower(userInput) {
	case "", strings.ToLower(tableFormat.String()):
		return tableFormat
	case strings.ToLower(jsonFormat.String()):
		return jsonFormat
	default:
		return unknownFormat
	}
}

// AvailableFormats is a list of presenter format options available to users.
var AvailableFormats = []format{
	jsonFormat,
	tableFormat,
}
