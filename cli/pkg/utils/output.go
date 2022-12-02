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

package utils

import (
	"fmt"
	"os"
)

func WriteSBOM(sbom []byte, output string) error {
	if output == "" {
		os.Stdout.Write(sbom)
		return nil
	}

	file, err := os.OpenFile(output, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666) // nolint:gomnd,gofumpt
	if err != nil {
		return fmt.Errorf("failed open file %s: %v", output, err)
	}
	defer file.Close()

	_, err = file.Write(sbom)
	if err != nil {
		return fmt.Errorf("failed to write sbom to file %s: %v", output, err)
	}
	return nil
}
