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

package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/client-go/util/jsonpath"
)

func PrintJSONData(data interface{}, fields string) error {
	// If jsonpath is not set it will print the whole data as json format.
	if fields == "" {
		dataB, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		fmt.Println(string(dataB))
		return nil
	}
	j := jsonpath.New("parser")
	if err := j.Parse(fields); err != nil {
		return fmt.Errorf("failed to parse jsonpath: %w", err)
	}
	err := j.Execute(os.Stdout, data)
	if err != nil {
		return fmt.Errorf("failed to execute jsonpath: %w", err)
	}
	return nil
}
