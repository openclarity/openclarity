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
	"io"

	"wwwin-github.cisco.com/eti/scan-gazr/cli/pkg/scanner/presenter/json"
	"wwwin-github.cisco.com/eti/scan-gazr/cli/pkg/scanner/presenter/table"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/scanner"
)

// Presenter is an interface for formatting and presenting scan results.
type Presenter interface {
	Present(io.Writer) error
}

// GetPresenter retrieves a Presenter that matches a CLI option.
func GetPresenter(conf Config, mergedResults *scanner.MergedResults) Presenter {
	switch conf.format {
	case jsonFormat:
		return json.NewPresenter(mergedResults)
	case tableFormat:
		return table.NewPresenter(mergedResults)
	case unknownFormat:
		return nil
	default:
		return nil
	}
}
