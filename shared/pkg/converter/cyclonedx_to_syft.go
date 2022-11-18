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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/formats/common/cyclonedxhelpers"
	syft_sbom "github.com/anchore/syft/syft/sbom"

	"github.com/openclarity/kubeclarity/shared/pkg/formatter"
)

var ErrFailedToGetCycloneDXSBOM = errors.New("failed to get CycloneDX SBOM from file")

func ConvertCycloneDXToSyftJSONFromFile(inputSBOMFile string, outputSBOMFile string) error {
	inputSBOM, err := getCycloneDXSBOMBytesFromFile(inputSBOMFile)
	if err != nil {
		return fmt.Errorf("failed to get CycloneDX SBOM  bytes from file: %v", err)
	}
	syftBOM, err := convertCycloneDXtoSyft(inputSBOM)
	if err != nil {
		return fmt.Errorf("failed to convert CycloneDX to syft format: %v", err)
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
	inputSBOMBytes, err := getCycloneDXSBOMBytesFromFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM bytes from file: %v", err)
	}

	// Ensure input is converted to cyclonedx regardless of the
	// input SBOM type.
	r := bytes.NewReader(inputSBOMBytes)
	sbom, format, err := syft.Decode(r)
	if err != nil {
		// syft's Decode has an issue with identifying cyclonedx XML
		// with an empty component list, if syft errors, and the first
		// line is the XML header, then just assume it is cyclonedx XML
		// and pass it on to switch statement below so that
		// cyclonedx-go can decode it.
		bufReader := bufio.NewReader(bytes.NewReader(inputSBOMBytes))
		firstLine, _, rErr := bufReader.ReadLine()
		if rErr == nil && string(firstLine) == `<?xml version="1.0" encoding="UTF-8"?>` {
			format = syft.FormatByName("cyclonedx")
		} else {
			// If no luck manually identifying the file as XML,
			// then just return the syft error.
			return nil, fmt.Errorf("failed to identify or decode file %s to recognised SBOM format: %w", inputSBOMFile, err)
		}
	}

	// If we've been given cyclonedx as the input decode directly using
	// cyclonedx-go instead of going through syft's intermediary struct
	// otherwise we may lose some properties/metadata which syft doesn't
	// understand. Other format's metadata is best effort so we'll use
	// syfts intermediary struct and then use ToFormatModel to switch it to
	// cdx.BOM.
	var bom *cdx.BOM
	cdxFormat := cdx.BOMFileFormatXML
	switch format {
	case syft.FormatByName("cyclonedxjson"):
		cdxFormat = cdx.BOMFileFormatJSON
		fallthrough
	case syft.FormatByName("cyclonedx"):
		bom = new(cdx.BOM)
		reader := bytes.NewReader(inputSBOMBytes)
		decoder := cdx.NewBOMDecoder(reader, cdxFormat)
		if err = decoder.Decode(bom); err != nil {
			return nil, fmt.Errorf("unable to decode CycloneDX BOM data: %w", err)
		}
	default:
		bom = cyclonedxhelpers.ToFormatModel(*sbom)
	}

	return bom, nil
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
