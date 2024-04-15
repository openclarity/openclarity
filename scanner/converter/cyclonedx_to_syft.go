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
	"errors"
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/anchore/syft/syft/format"
	"github.com/anchore/syft/syft/format/common/cyclonedxhelpers"
	"github.com/anchore/syft/syft/format/cyclonedxjson"
	"github.com/anchore/syft/syft/format/cyclonedxxml"
)

var ErrFailedToGetCycloneDXSBOM = errors.New("failed to get CycloneDX SBOM from file")

func GetCycloneDXSBOMFromFile(inputSBOMFile string) (*cdx.BOM, error) {
	inputSBOMBytes, err := os.ReadFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM bytes from file: %w", err)
	}

	return GetCycloneDXSBOMFromBytes(inputSBOMBytes)
}

func GetCycloneDXSBOMFromBytes(inputSBOMBytes []byte) (*cdx.BOM, error) {
	// Ensure input is converted to cyclonedx regardless of the
	// input SBOM type.
	r := bytes.NewReader(inputSBOMBytes)
	sbom, format, _, err := format.Decode(r)
	if err != nil {
		// syft's Decode has an issue with identifying cyclonedx XML
		// with an empty component list, if syft errors, and the first
		// line is the XML header, then just assume it is cyclonedx XML
		// and pass it on to switch statement below so that
		// cyclonedx-go can decode it.
		bufReader := bufio.NewReader(bytes.NewReader(inputSBOMBytes))
		firstLine, _, rErr := bufReader.ReadLine()
		if rErr == nil && string(firstLine) == `<?xml version="1.0" encoding="UTF-8"?>` {
			format = cyclonedxxml.ID
		} else {
			// If no luck manually identifying the file as XML,
			// then just return the syft error.
			return nil, fmt.Errorf("failed to identify or decode SBOM to recognised SBOM format: %w", err)
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
	case cyclonedxjson.ID:
		cdxFormat = cdx.BOMFileFormatJSON
		fallthrough
	case cyclonedxxml.ID:
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
