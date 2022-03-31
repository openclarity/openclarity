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

import "fmt"

// Config is the presenter domain's configuration data structure.
type Config struct {
	format format
}

// CreateConfig returns a new, validated presenter.Config. If a valid Config cannot be created using the given input,
// an error is returned.
func CreateConfig(output string) (Config, error) {
	format := getFormat(output)

	if format == unknownFormat {
		return Config{}, fmt.Errorf("unsupported output format %q, supported formats are: %+v", output,
			AvailableFormats)
	}

	return Config{
		format: format,
	}, nil
}
