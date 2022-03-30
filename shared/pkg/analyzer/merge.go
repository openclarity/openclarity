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

package analyzer

import (
	"fmt"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/formatter"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils"
)

type componentKey string // Unique identification of a package (name and version)

type MergedResults struct {
	MergedComponentByKey map[componentKey]*MergedComponent
	Source               utils.SourceType
	SourceHash           string
	SrcMetaData          *cdx.Metadata
}

type MergedComponent struct {
	Component    cdx.Component
	AnalyzerInfo []string
}

func NewMergedResults(sourceType utils.SourceType, hash string) *MergedResults {
	return &MergedResults{
		MergedComponentByKey: map[componentKey]*MergedComponent{},
		Source:               sourceType,
		SourceHash:           hash,
	}
}

func (m *MergedResults) Merge(other *Results, format string) *MergedResults {
	if other.Sbom == nil {
		return m
	}
	bom, err := decodeResults(other, format)
	if err != nil {
		log.Errorf("Failed to decode results: %v", err)
		return m
	}
	if m.SrcMetaData == nil {
		m.SrcMetaData = bom.Metadata
	} else {
		component := mergeCDXComponent(*m.SrcMetaData.Component, *bom.Metadata.Component, true)
		m.SrcMetaData.Component = &component
	}
	if other.AppInfo.SourceHash != "" {
		m.addSourceHash(other.AppInfo.SourceHash)
	}
	if bom.Components == nil {
		log.Errorf("Decoded bom doesn't contain any components")
		return m
	}
	otherComponentsByKey := toComponentByKey(bom.Components)
	for key, otherComponent := range otherComponentsByKey {
		if mergedComponent, ok := m.MergedComponentByKey[key]; !ok {
			log.Debugf("Adding new component results from %v. key=%v", other.AnalyzerInfo, key)
			m.MergedComponentByKey[key] = newMergedComponent(otherComponent, other.AnalyzerInfo)
		} else {
			log.Debugf("Adding existing component results from %v. key=%v", other.AnalyzerInfo, key)
			m.MergedComponentByKey[key] = handleComponentWithExistingKey(mergedComponent, otherComponent, other.AnalyzerInfo)
		}
	}

	return m
}

func (m *MergedResults) CreateMergedSBOMBytes(format, version string) ([]byte, error) {
	output := formatter.New(format, []byte{})
	if err := output.SetSBOM(m.createMergedSBOM(version)); err != nil {
		return nil, fmt.Errorf("failed to set SBOM: %v", err)
	}
	if err := output.Encode(format); err != nil {
		return nil, fmt.Errorf("failed to encode SBOM: %v", err)
	}
	return output.GetSBOMBytes(), nil
}

func (m *MergedResults) createComponentListFromMap() *[]cdx.Component {
	components := make([]cdx.Component, 0, len(m.MergedComponentByKey))
	for _, component := range m.MergedComponentByKey {
		components = append(components, component.Component)
	}

	return &components
}

func decodeResults(other *Results, format string) (*cdx.BOM, error) {
	bom := formatter.New(format, other.Sbom)
	if err := bom.Decode(format); err != nil {
		return nil, fmt.Errorf("failed to decode %s BOM", format)
	}
	cdxBOM, ok := bom.GetSBOM().(*cdx.BOM)
	if !ok {
		return nil, fmt.Errorf("failed to cast %s BOM", format)
	}

	return cdxBOM, nil
}

func createComponentKey(component cdx.Component) componentKey {
	return componentKey(fmt.Sprintf("%s.%s", component.Name, component.Version))
}

func toComponentByKey(components *[]cdx.Component) map[componentKey]cdx.Component {
	ret := make(map[componentKey]cdx.Component, len(*components))
	for _, component := range *components {
		ret[createComponentKey(component)] = component
	}
	return ret
}

func newMergedComponent(component cdx.Component, analyzerInfo string) *MergedComponent {
	mergedComponent := &MergedComponent{
		Component: component,
	}
	mergedComponent.appendAnalyzerInfo(analyzerInfo)

	return mergedComponent
}

func handleComponentWithExistingKey(mergedComponent *MergedComponent, otherComponent cdx.Component, analyzerInfo string) *MergedComponent {
	mergedComponent.Component = mergeCDXComponent(mergedComponent.Component, otherComponent, false)
	mergedComponent.appendAnalyzerInfo(analyzerInfo)

	return mergedComponent
}

// simple merge functionality because we currently use only syft and cyclonedx-gomod
// later we need to extend it. Syft cycloneDX output is very poor, only contains name, version, purl and type.
// nolint:cyclop
func mergeCDXComponent(mergedComponent, otherComponent cdx.Component, main bool) cdx.Component {
	// check main component contained by the metadata
	if main {
		// The name of the component is overwritten if it's a package name.
		// Only used for the CycloneDX BOM metadata.
		if mergedComponent.Name != otherComponent.Name {
			mergedComponent.Name = checkMainComponentName(mergedComponent.Name, otherComponent.Name)
		}
		// The version of the component is overwritten if it's a main component.
		// Only used for the CycloneDX BOM metadata.
		if mergedComponent.Version == "" {
			mergedComponent.Version = otherComponent.Version
		}
	}
	// BOMRef is only provided by the cycloneDX-gomod, need to check it before override it.
	if mergedComponent.BOMRef == "" {
		mergedComponent.BOMRef = otherComponent.BOMRef
	}
	// CPE is not porvided by the cycloneDX-gomod and sift at the moment.
	if mergedComponent.CPE == "" {
		mergedComponent.CPE = otherComponent.CPE
	}
	// Author isn't provided by cycloneDX-gomod and syft at the moment.
	if mergedComponent.Author == "" {
		mergedComponent.Author = otherComponent.Author
	}
	// Copyright isn't provided by cycloneDX-gomod and syft at the moment.
	if mergedComponent.Copyright == "" {
		mergedComponent.Copyright = otherComponent.Copyright
	}
	// Description can be provided by the cycloneDX-gomod.
	if mergedComponent.Description == "" {
		mergedComponent.Description = otherComponent.Description
	}
	// Evidence isn't provided by cycloneDX-gomod and syft at the moment.
	if mergedComponent.Evidence == nil {
		mergedComponent.Evidence = otherComponent.Evidence
	}
	// ExternalReferences are provided by the cycloneDX-gomod.
	if mergedComponent.ExternalReferences == nil {
		mergedComponent.ExternalReferences = otherComponent.ExternalReferences
	}
	// SWID isn't provided by cycloneDX-gomod and syft at the moment.
	if mergedComponent.SWID == nil {
		mergedComponent.SWID = otherComponent.SWID
	}
	// Supplier isn't provided by cycloneDX-gomod and syft at the moment.
	if mergedComponent.Supplier == nil {
		mergedComponent.Supplier = otherComponent.Supplier
	}
	// Publisher isn't provided by cycloneDX-gomod and syft at the moment.
	if mergedComponent.Publisher == "" {
		mergedComponent.Publisher = otherComponent.Publisher
	}
	// Group can be provided by only the cycloneDX-gomod.
	if mergedComponent.Group == "" {
		mergedComponent.Group = otherComponent.Group
	}
	// Scope can be provided by only the cycloneDX-gomod.
	if mergedComponent.Scope == "" {
		mergedComponent.Scope = otherComponent.Scope
	}
	// Hashes can be provided by only the cycloneDX-gomod.
	if mergedComponent.Hashes == nil {
		mergedComponent.Hashes = otherComponent.Hashes
	}
	// Licenses can be provided by only the cycloneDX-gomod.
	if mergedComponent.Licenses == nil {
		mergedComponent.Licenses = otherComponent.Licenses
	}
	// Properties can be provided by the syft and te cycloneDX-gomod, needs to be merged.
	if otherComponent.Properties != nil {
		mergeProperties(mergedComponent.Properties, otherComponent.Properties)
	}
	// PackageURL in the case of cycloneDX-gomod contains type of the package as well.
	// We use the longer PURL.
	if len(mergedComponent.PackageURL) < len(otherComponent.PackageURL) {
		mergedComponent.PackageURL = otherComponent.PackageURL
	}
	// Other unprovided fields: MIMEType, Supplier, Modified, Pedigree

	return mergedComponent
}

func mergeProperties(properties, otherProperties *[]cdx.Property) *[]cdx.Property {
	if properties == nil {
		properties = &[]cdx.Property{}
	}
	propSlice := *properties
	propSlice = append(propSlice, *otherProperties...)

	return &propSlice
}

func (mc *MergedComponent) appendAnalyzerInfo(info string) *MergedComponent {
	analyzerInfo := cdx.Property{
		Name:  "analyzers",
		Value: info,
	}
	if mc.Component.Properties == nil {
		mc.Component.Properties = &[]cdx.Property{analyzerInfo}
	} else {
		analyzers := *mc.Component.Properties
		analyzers = append(analyzers, analyzerInfo)
		mc.Component.Properties = &analyzers
	}

	mc.AnalyzerInfo = append(mc.AnalyzerInfo, info)

	return mc
}

func (m *MergedResults) createMergedSBOM(version string) *cdx.BOM {
	cdxBOM := cdx.NewBOM()
	versionInfo := version
	cdxBOM.SerialNumber = uuid.New().URN()
	if m.SrcMetaData == nil {
		log.Errorf("Failed to get source metadata")
		return cdxBOM
	}
	cdxBOM.Metadata = toBomDescriptor("kubeclarity", versionInfo, m.Source, m.SrcMetaData, m.SourceHash)
	cdxBOM.Components = m.createComponentListFromMap()

	return cdxBOM
}

// toBomDescriptor returns metadata tailored for the current time and tool details.
func toBomDescriptor(name, version string, source utils.SourceType, srcMetadata *cdx.Metadata, hash string) *cdx.Metadata {
	return &cdx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Tools: &[]cdx.Tool{
			{
				Vendor:  "kubeclarity",
				Name:    name,
				Version: version,
			},
		},
		Component: toBomDescriptorComponent(source, srcMetadata, hash),
	}
}

func toBomDescriptorComponent(sourceType utils.SourceType, srcMetadata *cdx.Metadata, hash string) *cdx.Component {
	if srcMetadata.Component == nil {
		return nil
	}
	metaDataComponent := srcMetadata.Component

	// nolint:exhaustive
	switch sourceType {
	case utils.IMAGE:
		metaDataComponent.Type = cdx.ComponentTypeContainer
	case utils.DIR, utils.FILE:
		metaDataComponent.Type = cdx.ComponentTypeFile
		metaDataComponent.Hashes = &[]cdx.Hash{
			{
				Algorithm: cdx.HashAlgoSHA256,
				Value:     hash,
			},
		}
	}

	return metaDataComponent
}

// check the name of main module if it got successfully by the cyclonedx-gomod.
func checkMainComponentName(mergedName, otherName string) string {
	var path bool
	if strings.HasPrefix(mergedName, "/") || strings.HasPrefix(mergedName, ".") {
		path = true
	}
	if !(strings.HasPrefix(otherName, "/") || strings.HasPrefix(otherName, ".")) && path && otherName != "" {
		return otherName
	}

	return mergedName
}

func (m *MergedResults) addSourceHash(sourceHash string) *MergedResults {
	if sourceHash == "" {
		return m
	}
	repoDigestHash := cdx.Hash{
		Algorithm: cdx.HashAlgoSHA256,
		Value:     sourceHash,
	}
	m.SrcMetaData.Component.Hashes = &[]cdx.Hash{
		repoDigestHash,
	}

	return m
}
