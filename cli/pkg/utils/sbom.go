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

	"github.com/openclarity/kubeclarity/shared/pkg/converter"
	"github.com/openclarity/kubeclarity/shared/pkg/formatter"
)

func ConvertInputSBOMIfNeeded(inputSBOMFile, outputFormat string) ([]byte, error) {
	inputSBOM, err := os.ReadFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file %s: %v", inputSBOMFile, err)
	}
	inputSBOMFormat := converter.DetermineCycloneDXFormat(inputSBOM)
	if inputSBOMFormat == outputFormat {
		return inputSBOM, nil
	}

	// Create cycloneDX formatter to convert input SBOM to the defined output format.
	cdxFormatter := formatter.New(inputSBOMFormat, inputSBOM)
	if err = cdxFormatter.Decode(inputSBOMFormat); err != nil {
		return nil, fmt.Errorf("failed to decode input SBOM %s: %v", inputSBOMFile, err)
	}
	if err := cdxFormatter.Encode(outputFormat); err != nil {
		return nil, fmt.Errorf("failed to encode input SBOM: %v", err)
	}

	return cdxFormatter.GetSBOMBytes(), nil
}
