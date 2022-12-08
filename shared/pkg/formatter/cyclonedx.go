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

package formatter

import (
	"bytes"
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

const (
	CycloneDXFormat     = "cyclonedx"
	CycloneDXJSONFormat = "cyclonedx-json"
)

type CycloneDX struct {
	name       string
	sbomBytes  []byte   // encoded BOM bytes
	sbomStruct *cdx.BOM // decoded cycloneDX BOM struct
}

func newCycloneDXFormatter(sbomBytes []byte) Formatter {
	return &CycloneDX{
		name:       CycloneDXFormat,
		sbomBytes:  sbomBytes,
		sbomStruct: &cdx.BOM{},
	}
}

func (f *CycloneDX) SetSBOM(bom interface{}) error {
	sbom, ok := bom.(*cdx.BOM)
	if !ok {
		return fmt.Errorf("failed to set BOM in cycloneDX formatter")
	}
	f.sbomStruct = sbom

	return nil
}

func (f *CycloneDX) GetSBOM() interface{} {
	return f.sbomStruct
}

func (f *CycloneDX) GetSBOMBytes() []byte {
	return f.sbomBytes
}

func (f *CycloneDX) Encode(format string) error {
	cdxFormat := getCycloneDXFormat(format)

	var buff bytes.Buffer
	encoder := cdx.NewBOMEncoder(&buff, cdxFormat)
	encoder.SetPretty(true)

	if err := encoder.Encode(f.sbomStruct); err != nil {
		return fmt.Errorf("failed to encode sbom: %v", err)
	}
	f.sbomBytes = buff.Bytes()

	return nil
}

func (f *CycloneDX) Decode(format string) error {
	cdxFormat := getCycloneDXFormat(format)
	reader := bytes.NewReader(f.sbomBytes)
	// Decode the BOM
	bom := new(cdx.BOM)
	decoder := cdx.NewBOMDecoder(reader, cdxFormat)
	if err := decoder.Decode(bom); err != nil {
		return fmt.Errorf("unable to decode BOM data: %v", err)
	}
	if err := f.SetSBOM(bom); err != nil {
		return fmt.Errorf("unable to set BOM data: %v", err)
	}

	return nil
}

func getCycloneDXFormat(format string) cdx.BOMFileFormat {
	formatOption := cdx.BOMFileFormatXML
	switch format {
	case CycloneDXFormat:
		formatOption = cdx.BOMFileFormatXML
	case CycloneDXJSONFormat:
		formatOption = cdx.BOMFileFormatJSON
	}
	return formatOption
}
