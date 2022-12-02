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
	"bytes"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/formats/common/cyclonedxhelpers"
	syftSbom "github.com/anchore/syft/syft/sbom"
)

type SbomFormat uint

const (
	Unknown SbomFormat = iota
	CycloneDxJSON
	CycloneDxXML
	SpdxJSON
	SpdxTV
	SyftJSON
)

func (c SbomFormat) String() string {
	strings := []string{"cyclonedx-json", "cyclonedx-xml", "spdx-json", "spdx-keyvalue", "syft"}
	if c > Unknown && int(c) < len(strings) {
		return strings[c]
	}
	return "unknown"
}

func StringToSbomFormat(input string) (SbomFormat, error) {
	switch input {
	case "cyclonedx", "cyclonedx-xml":
		return CycloneDxXML, nil
	case "cyclonedx-json":
		return CycloneDxJSON, nil
	case "spdx", "spdx-json":
		return SpdxJSON, nil
	case "spdx-tv":
		return SpdxTV, nil
	case "syft", "syft-json":
		return SyftJSON, nil
	}
	return Unknown, fmt.Errorf("unknown sbom format %v", input)
}

func CycloneDxToBytes(sbom *cdx.BOM, format SbomFormat) ([]byte, error) {
	switch format {
	case CycloneDxXML, CycloneDxJSON:
		return cycloneDxToBytesUsingCycloneDxEncoder(sbom, format)
	case SpdxJSON, SpdxTV, SyftJSON:
		return cycloneDxToBytesUsingSyftConversion(sbom, format)
	case Unknown:
	default:
	}
	return nil, fmt.Errorf("can not convert cyclonedx SBOM to unknown format")
}

// cycloneDxToBytesUsingCycloneDxEncoder supports encoding a cdx.BOM to one of
// CycloneDX's native formats, cyclonedx-json or cyclonedx-xml, other formats
// will return an error.
func cycloneDxToBytesUsingCycloneDxEncoder(sbom *cdx.BOM, format SbomFormat) ([]byte, error) {
	var cdxFormat cdx.BOMFileFormat
	switch format {
	case CycloneDxXML:
		cdxFormat = cdx.BOMFileFormatXML
	case CycloneDxJSON:
		cdxFormat = cdx.BOMFileFormatJSON
	case SpdxJSON, SpdxTV, SyftJSON, Unknown:
		fallthrough
	default:
		return nil, fmt.Errorf("format %v is not a native cyclonedx format", format)
	}

	var buff bytes.Buffer
	encoder := cdx.NewBOMEncoder(&buff, cdxFormat)
	encoder.SetPretty(true)

	if err := encoder.Encode(sbom); err != nil {
		return nil, fmt.Errorf("failed to encode sbom: %w", err)
	}
	return buff.Bytes(), nil
}

// cycloneDxToBytesUsingSyftConversion supports encoding a cdx.BOM to a number
// of formats by converting it to a syft SBOM first and then re-encoding it to
// the destination format. The process can be lossy so if encoding to a
// cyclonedx format (json or xml) use cycloneDxToBytesUsingCycloneDxEncoder
// instead.
func cycloneDxToBytesUsingSyftConversion(sbom *cdx.BOM, format SbomFormat) ([]byte, error) {
	syftSBOM, err := cyclonedxhelpers.ToSyftModel(sbom)
	if err != nil {
		return nil, fmt.Errorf("unable to convert BOM to intermediary format: %w", err)
	}

	var syftFormat syftSbom.Format
	switch format {
	case SpdxJSON:
		syftFormat = syft.FormatByName("spdxjson")
	case SpdxTV:
		syftFormat = syft.FormatByName("spdxtagvalue")
	case SyftJSON:
		syftFormat = syft.FormatByName("syftjson")
	case CycloneDxXML, CycloneDxJSON, Unknown:
		fallthrough
	default:
		return nil, fmt.Errorf("format %v is a native cyclonedx format, use CycloneDxToNativeFormatBytes instead", format)
	}

	data, err := syft.Encode(*syftSBOM, syftFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sbom: %w", err)
	}
	return data, nil
}
