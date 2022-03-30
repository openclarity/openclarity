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
	"errors"
	"fmt"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/anchore/syft/syft/linux"
	syft_pkg "github.com/anchore/syft/syft/pkg"
	syft_cpe "github.com/anchore/syft/syft/pkg/cataloger/common/cpe"
	syft_sbom "github.com/anchore/syft/syft/sbom"
	syft_source "github.com/anchore/syft/syft/source"
	"github.com/google/uuid"
	purl "github.com/package-url/packageurl-go"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/formatter"
	cdx_helper "wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils/cyclonedx_helper"
)

type propertiesInfo struct {
	locations []syft_source.Location
	metaData  interface{}
}

type emptyMetadata struct{}

var ErrFailedToGetCycloneDXSBOM = errors.New("failed to get CycloneDX SBOM from file")

func ConvertCycloneDXToSyftJSONFromFile(inputSBOMFile string, outputSBOMFile string) error {
	cdxBOM, err := getCycloneDXSBOMFromFile(inputSBOMFile)
	if err != nil {
		return ErrFailedToGetCycloneDXSBOM
	}

	syftBOM, err := convertCycloneDXtoSyft(cdxBOM)
	if err != nil {
		return fmt.Errorf("failed to convert cycloneDX to syft format: %v", err)
	}

	if err = saveSyftSBOMToFile(syftBOM, outputSBOMFile); err != nil {
		return fmt.Errorf("failed to save syft SBOM: %v", err)
	}

	return nil
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

func getCycloneDXSBOMFromFile(inputSBOMFile string) (*cdx.BOM, error) {
	// TODO: Peter, is that the right input format?
	inputFormat := formatter.CycloneDXFormat

	inputSBOM, err := os.ReadFile(inputSBOMFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read SBOM file %s: %v", inputSBOMFile, err)
	}

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

// nolint:cyclop
func convertCycloneDXtoSyft(bom *cdx.BOM) (syft_sbom.SBOM, error) {
	var intype syft_source.Scheme
	if bom == nil {
		return syft_sbom.SBOM{}, fmt.Errorf("cycloneDX BOM is nil")
	}
	if bom.Metadata == nil {
		return syft_sbom.SBOM{}, fmt.Errorf("cycloneDX metadata is nil")
	}
	if bom.Metadata.Component == nil {
		return syft_sbom.SBOM{}, fmt.Errorf("cycloneDX metadata component is nil")
	}
	// nolint:exhaustive
	switch bom.Metadata.Component.Type {
	case cdx.ComponentTypeContainer:
		intype = syft_source.ImageScheme
	case cdx.ComponentTypeFile, cdx.ComponentTypeApplication, "":
		intype = syft_source.DirectoryScheme
	}

	syftSbom := syft_sbom.SBOM{
		Artifacts: syft_sbom.Artifacts{
			PackageCatalog: syft_pkg.NewCatalog(),
		},
		Source: syft_source.Metadata{
			Scheme: intype,
			Path:   bom.Metadata.Component.Name,
		},
	}
	if syft_source.Scheme(bom.Metadata.Component.Type) == "container" {
		syftSbom.Source.ImageMetadata = syft_source.ImageMetadata{}
	}

	for _, component := range *bom.Components {
		if component.Type == "operating-system" {
			if component.Properties != nil {
				syftSbom.Artifacts.LinuxDistribution = getImageDistro(component.Properties)
			}
		}
		if component.PackageURL == "" {
			continue
		}
		metaType, pkgType, lang := determineSourceMeta(component.PackageURL)
		propertyInfo := propertiesInfo{
			locations: []syft_source.Location{},
			metaData:  nil,
		}
		if component.Properties != nil {
			propertyInfo = parsePackageProperties(component.Properties)
		} else {
			// When syft generates a CycloneDX format, it can contain a package multiple times if a package has multiple locations.
			// If totally same package will be added to the syft PackageCatalog, the ID of it will be the same and cause problems.
			// If properties are not set which contain location information that is different in the same package,
			// we will have to generate unique locations to avoid adding a package to PackageCatalog with the same ID,
			// because the ID is generated based on package fields.
			propertyInfo.locations = []syft_source.Location{
				syft_source.NewLocation(fmt.Sprintf("%s/%s", component.Name, uuid.New().String())),
			}
		}
		metaData := propertyInfo.metaData
		if metaData == nil {
			metaData = emptyMetadata{}
		}
		pkg := syft_pkg.Package{
			Name:         component.Name,
			Version:      component.Version,
			Type:         pkgType,
			PURL:         component.PackageURL,
			MetadataType: metaType,
			Metadata:     metaData,
			Locations:    propertyInfo.locations,
			Language:     lang,
			Licenses:     cdx_helper.GetComponentLicenses(component),
		}
		pkg.CPEs = syft_cpe.Generate(pkg)
		syftSbom.Artifacts.PackageCatalog.Add(pkg)
	}

	return syftSbom, nil
}

// determineSourceMeta set syft package fields based on the PURL.
// nolint:cyclop
func determineSourceMeta(packageURL string) (syft_pkg.MetadataType, syft_pkg.Type, syft_pkg.Language) {
	var metaType syft_pkg.MetadataType
	var pkgType syft_pkg.Type
	var lang syft_pkg.Language
	purl, err := purl.FromString(packageURL)
	if err != nil {
		return metaType, pkgType, lang
	}
	switch purl.Type {
	case "golang":
		metaType = syft_pkg.GolangBinMetadataType
		pkgType = syft_pkg.GoModulePkg
		lang = syft_pkg.Go
	case "pypi":
		metaType = syft_pkg.PythonPackageMetadataType
		pkgType = syft_pkg.PythonPkg
		lang = syft_pkg.Python
	case "npm":
		metaType = syft_pkg.NpmPackageJSONMetadataType
		pkgType = syft_pkg.NpmPkg
		lang = syft_pkg.JavaScript
	case "gem":
		metaType = syft_pkg.GemMetadataType
		pkgType = syft_pkg.GemPkg
		lang = syft_pkg.Ruby
	case "cargo":
		metaType = syft_pkg.RustCargoPackageMetadataType
		pkgType = syft_pkg.RustPkg
		lang = syft_pkg.Rust
	case "maven":
		metaType = syft_pkg.JavaMetadataType
		pkgType = syft_pkg.JavaPkg
		lang = syft_pkg.Java
	case "alpine":
		metaType = syft_pkg.ApkMetadataType
		pkgType = syft_pkg.ApkPkg
	case "deb":
		metaType = syft_pkg.DpkgMetadataType
		pkgType = syft_pkg.DebPkg
	case "rpm":
		metaType = syft_pkg.RpmdbMetadataType
		pkgType = syft_pkg.RpmPkg
	default:
		metaType = syft_pkg.UnknownMetadataType
		pkgType = syft_pkg.UnknownPkg
		lang = syft_pkg.UnknownLanguage
	}
	return metaType, pkgType, lang
}

// parsePackageProperties get additional information from property fields and provide info for converting.
// nolint:cyclop
func parsePackageProperties(properties *[]cdx.Property) propertiesInfo {
	location := syft_source.Location{}
	apkMetadata := syft_pkg.ApkMetadata{}
	dpkgMetadata := syft_pkg.DpkgMetadata{}
	rpmMetadata := syft_pkg.RpmdbMetadata{}
	javaMetadata := syft_pkg.JavaMetadata{}
	var metaData interface{}
	for _, property := range *properties {
		switch property.Name {
		case "path":
			location.RealPath = property.Value
		case "layerID":
			location.FileSystemID = property.Value
		case "originPackage":
			apkMetadata.OriginPackage = property.Value
			metaData = apkMetadata
		case "source":
			dpkgMetadata.Source = property.Value
			metaData = dpkgMetadata
		case "sourceRpm":
			rpmMetadata.SourceRpm = property.Value
			metaData = rpmMetadata
		case "artifactID":
			if javaMetadata.PomProperties == nil {
				javaMetadata.PomProperties = &syft_pkg.PomProperties{
					ArtifactID: property.Value,
				}
			} else {
				javaMetadata.PomProperties.ArtifactID = property.Value
			}
			metaData = javaMetadata
		case "groupID":
			if javaMetadata.PomProperties == nil {
				javaMetadata.PomProperties = &syft_pkg.PomProperties{
					GroupID: property.Value,
				}
			} else {
				javaMetadata.PomProperties.GroupID = property.Value
			}
			metaData = javaMetadata
		}
	}
	return propertiesInfo{
		locations: []syft_source.Location{location},
		metaData:  metaData,
	}
}

// getImageDistro get distro from additional property filed.
func getImageDistro(properties *[]cdx.Property) *linux.Release {
	var d linux.Release
	for _, property := range *properties {
		switch property.Name {
		case "id":
			d.ID = property.Value
		case "versionID":
			d.VersionID = property.Value
		}
	}

	return &d
}
