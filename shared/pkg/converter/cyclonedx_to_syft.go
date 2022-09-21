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

package converter

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	syft_sbom "github.com/anchore/syft/syft/sbom"

	"github.com/openclarity/kubeclarity/shared/pkg/formatter"
)

var ErrFailedToGetCycloneDXSBOM = errors.New("failed to get CycloneDX SBOM from file")

func ConvertCycloneDXToSyftJSONFromFile(inputSBOMFile string, outputSBOMFile string) error {
	inputSBOM, err := getCycloneDXSBOMBytesFromFile(inputSBOMFile)
	if err != nil {
		return fmt.Errorf("failed to get cyclonDX SBOM  bytes from file: %v", err)
	}
	syftBOM, err := convertCycloneDXtoSyft(inputSBOM)
	if err != nil {
		return fmt.Errorf("failed to convert cycloneDX to syft format: %v", err)
	}

	if err = saveSyftSBOMToFile(syftBOM, outputSBOMFile); err != nil {
		return fmt.Errorf("failed to save syft SBOM: %v", err)
	}

	return nil
}

func getCycloneDXSBOMBytesFromFile(inputSBOMFile string) ([]byte, error) {
	inputSBOM, err := os.ReadFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file %s: %v", inputSBOMFile, err)
	}
	return inputSBOM, nil
}

func saveSyftSBOMToFile(syftBOM syft_sbom.SBOM, outputSBOMFile string) error {
	outputFormat := formatter.SyftFormat

	output := formatter.New(outputFormat, []byte{})
	if err := output.SetSBOM(syftBOM); err != nil {
		return fmt.Errorf("unable to set SBOM in formatter: %v", err)
	}

	if err := output.Encode(outputFormat); err != nil {
		return fmt.Errorf("failed to encode SBOM: %v", err)
	}

	if err := formatter.WriteSBOM(output.GetSBOMBytes(), outputSBOMFile); err != nil {
		return fmt.Errorf("failed to write syft SBOM to file %s: %v", outputSBOMFile, err)
	}

	return nil
}

func GetCycloneDXSBOMFromFile(inputSBOMFile string) (*cdx.BOM, error) {
	inputSBOM, err := getCycloneDXSBOMBytesFromFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get cyclonDX SBOM  bytes from file: %v", err)
	}
	inputFormat := DetermineCycloneDXFormat(inputSBOM)
	input := formatter.New(inputFormat, inputSBOM)
	// use the formatter
	if err = input.Decode(inputFormat); err != nil {
		return nil, fmt.Errorf("unable to decode input SBOM %s: %v", inputSBOMFile, err)
	}

	cdxBOM, ok := input.GetSBOM().(*cdx.BOM)
	if !ok {
		return nil, fmt.Errorf("failed to cast input SBOM: %v", err)
	}

	return cdxBOM, nil
}

func DetermineCycloneDXFormat(sbom []byte) string {
	var js json.RawMessage
	if json.Unmarshal(sbom, &js) == nil {
		return formatter.CycloneDXJSONFormat
	}

	return formatter.CycloneDXFormat
}

func convertCycloneDXtoSyft(sbomB []byte) (syft_sbom.SBOM, error) {
	output := formatter.New(formatter.SyftFormat, sbomB)

	if err := output.Decode(formatter.CycloneDXFormat); err != nil {
		return syft_sbom.SBOM{}, fmt.Errorf("failed to write results: %v", err)
	}
	sbom, ok := output.GetSBOM().(syft_sbom.SBOM)
	if !ok {
		return syft_sbom.SBOM{}, fmt.Errorf("type assertion of sbom failed")
	}
	return sbom, nil
}
