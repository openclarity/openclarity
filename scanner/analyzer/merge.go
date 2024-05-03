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
	"sort"
	"strings"
	"time"

	"github.com/openclarity/vmclarity/scanner/utils"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/converter"
)

type componentKey string // Unique identification of a package (name and version)

type MergedResults struct {
	MergedComponentByKey map[componentKey]*MergedComponent
	Source               utils.SourceType
	SourceHash           string
	SrcMetaData          *cdx.Metadata
	SrcMetaDataBomRefs   []string
	Dependencies         *[]cdx.Dependency
}

type MergedComponent struct {
	Component    cdx.Component
	AnalyzerInfo []string
	BomRefs      []string
}

func NewMergedResults(sourceType utils.SourceType, hash string) *MergedResults {
	return &MergedResults{
		MergedComponentByKey: map[componentKey]*MergedComponent{},
		Source:               sourceType,
		SourceHash:           hash,
	}
}

func (m *MergedResults) Merge(other *Results) *MergedResults {
	if other.Sbom == nil {
		return m
	}
	bom := other.Sbom

	// merge bom.Metadata.Component if it exists
	m.mergeMainComponent(bom)

	if other.AppInfo.SourceHash != "" {
		m.addSourceHash(other.AppInfo.SourceHash)
	}

	m.addSourceMetadata(other.AppInfo.SourceMetadata)

	if bom.Components != nil {
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
	}

	if bom.Dependencies != nil {
		newDependencies := m.normalizeDependencies(bom.Dependencies)
		m.Dependencies = mergeDependencies(m.Dependencies, newDependencies)
	}

	return m
}

func (m *MergedResults) mergeMainComponent(bom *cdx.BOM) {
	// Keep track of all SBOM refs given to the main component so that we
	// can normalize them later.
	if bom.Metadata != nil && bom.Metadata.Component != nil && bom.Metadata.Component.BOMRef != "" {
		m.SrcMetaDataBomRefs = append(m.SrcMetaDataBomRefs, bom.Metadata.Component.BOMRef)
	}

	if m.SrcMetaData == nil {
		m.SrcMetaData = bom.Metadata
	} else {
		component := mergeCDXComponent(*m.SrcMetaData.Component, *bom.Metadata.Component, true)
		m.SrcMetaData.Component = &component
	}
}

func (m *MergedResults) CreateMergedSBOMBytes(format, version string) ([]byte, error) {
	cdxSBOM := m.createMergedSBOM(version)
	f, err := converter.StringToSbomFormat(format)
	if err != nil {
		return nil, fmt.Errorf("unable to parse output format: %w", err)
	}
	bomBytes, err := converter.CycloneDxToBytes(cdxSBOM, f)
	if err != nil {
		return nil, fmt.Errorf("failed to get bom bytes: %w", err)
	}
	return bomBytes, nil
}

func (m *MergedResults) createComponentListFromMap() *[]cdx.Component {
	components := make([]cdx.Component, 0, len(m.MergedComponentByKey))
	for _, component := range m.MergedComponentByKey {
		components = append(components, component.Component)
	}

	return &components
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
	if component.BOMRef != "" {
		mergedComponent.BomRefs = append(mergedComponent.BomRefs, component.BOMRef)
	}

	return mergedComponent
}

func handleComponentWithExistingKey(mergedComponent *MergedComponent, otherComponent cdx.Component, analyzerInfo string) *MergedComponent {
	mergedComponent.Component = mergeCDXComponent(mergedComponent.Component, otherComponent, false)
	mergedComponent.appendAnalyzerInfo(analyzerInfo)
	if otherComponent.BOMRef != "" {
		mergedComponent.BomRefs = append(mergedComponent.BomRefs, otherComponent.BOMRef)
	}

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

	if mergedComponent.BOMRef == "" {
		mergedComponent.BOMRef = otherComponent.BOMRef
	}

	if mergedComponent.CPE == "" {
		mergedComponent.CPE = otherComponent.CPE
	}

	if mergedComponent.Author == "" {
		mergedComponent.Author = otherComponent.Author
	}

	if mergedComponent.Copyright == "" {
		mergedComponent.Copyright = otherComponent.Copyright
	}

	if mergedComponent.Description == "" {
		mergedComponent.Description = otherComponent.Description
	}

	if mergedComponent.Evidence == nil {
		mergedComponent.Evidence = otherComponent.Evidence
	}

	if mergedComponent.ExternalReferences == nil {
		mergedComponent.ExternalReferences = otherComponent.ExternalReferences
	}

	if mergedComponent.SWID == nil {
		mergedComponent.SWID = otherComponent.SWID
	}

	if mergedComponent.Supplier == nil {
		mergedComponent.Supplier = otherComponent.Supplier
	}

	if mergedComponent.Publisher == "" {
		mergedComponent.Publisher = otherComponent.Publisher
	}

	if mergedComponent.Group == "" {
		mergedComponent.Group = otherComponent.Group
	}

	if mergedComponent.Scope == "" {
		mergedComponent.Scope = otherComponent.Scope
	}

	if mergedComponent.Hashes == nil {
		mergedComponent.Hashes = otherComponent.Hashes
	}

	mergedComponent.Licenses = mergeLicenses(mergedComponent.Licenses, otherComponent.Licenses)

	if otherComponent.Properties != nil {
		mergedComponent.Properties = mergeProperties(mergedComponent.Properties, otherComponent.Properties)
	}

	mergedComponent.PackageURL = mergePurlStrings(mergedComponent.PackageURL, otherComponent.PackageURL)
	// Other unprovided fields: MIMEType, Supplier, Modified, Pedigree

	return mergedComponent
}

func mergeLicenses(licenseA, licenseB *cdx.Licenses) *cdx.Licenses {
	// nothing to merge into A so return the A untouched
	if licenseB == nil {
		return licenseA
	}

	// If A has no licenses initialise it
	if licenseA == nil {
		licenseA = &cdx.Licenses{}
	}

	// Merge B into A, assign to new variable to avoid modifying licenseA
	newthing := append(*licenseA, *licenseB...)
	return &newthing
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
	cdxBOM.Metadata = toBomDescriptor("vmclarity", versionInfo, m.Source, m.SrcMetaData, m.SourceHash)
	cdxBOM.Components = m.createComponentListFromMap()
	cdxBOM.Dependencies = m.Dependencies

	return cdxBOM
}

func (m *MergedResults) normalizeDependencies(dependencies *[]cdx.Dependency) *[]cdx.Dependency {
	if dependencies == nil {
		return nil
	}

	output := []cdx.Dependency{}
	for _, dependency := range *dependencies {
		newDep := cdx.Dependency{
			Ref: m.getRealBomRefFromPreviousBomRef(dependency.Ref),
		}

		if dependency.Dependencies != nil {
			var newDependsOn []string
			for _, dependsOnRef := range *dependency.Dependencies {
				newDependsOn = append(newDependsOn, m.getRealBomRefFromPreviousBomRef(dependsOnRef))
			}
			newDep.Dependencies = &newDependsOn
		}

		output = append(output, newDep)
	}

	return &output
}

func mergeDependencies(depsA, depsB *[]cdx.Dependency) *[]cdx.Dependency {
	refToDepends := map[string]map[string]struct{}{}
	addDepsToRefToDepends := func(ref string, deps *[]string) {
		if deps == nil {
			return
		}

		// initialize refToDepends entry if it doesn't exist
		existing, ok := refToDepends[ref]
		if !ok {
			refToDepends[ref] = map[string]struct{}{}
			existing = refToDepends[ref]
		}

		// add entries to the refToDepends set
		for _, dependsOnRef := range *deps {
			existing[dependsOnRef] = struct{}{}
		}
	}

	if depsA != nil {
		for _, dependency := range *depsA {
			addDepsToRefToDepends(dependency.Ref, dependency.Dependencies)
		}
	}

	if depsB != nil {
		for _, dependency := range *depsB {
			addDepsToRefToDepends(dependency.Ref, dependency.Dependencies)
		}
	}

	output := []cdx.Dependency{}
	for ref, depends := range refToDepends {
		var dependsOn []string
		for dRef := range depends {
			dependsOn = append(dependsOn, dRef)
		}
		sort.Strings(dependsOn)
		output = append(output, cdx.Dependency{
			Ref:          ref,
			Dependencies: &dependsOn,
		})
	}

	return &output
}

func (m *MergedResults) getRealBomRefFromPreviousBomRef(bomRef string) string {
	for _, mergedComponent := range m.MergedComponentByKey {
		for _, ref := range mergedComponent.BomRefs {
			if ref == bomRef {
				return mergedComponent.Component.BOMRef
			}
		}
	}

	for _, ref := range m.SrcMetaDataBomRefs {
		if ref == bomRef {
			return m.SrcMetaData.Component.BOMRef
		}
	}

	return bomRef
}

// toBomDescriptor returns metadata tailored for the current time and tool details.
func toBomDescriptor(name, version string, source utils.SourceType, srcMetadata *cdx.Metadata, hash string) *cdx.Metadata {
	return &cdx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Tools: &cdx.ToolsChoice{
			Components: &[]cdx.Component{
				{
					Author:  "vmclarity",
					Name:    name,
					Version: version,
				},
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

	switch sourceType {
	case utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		metaDataComponent.Type = cdx.ComponentTypeContainer
	case utils.DIR, utils.FILE, utils.ROOTFS:
		metaDataComponent.Type = cdx.ComponentTypeFile
		metaDataComponent.Hashes = &[]cdx.Hash{
			{
				Algorithm: cdx.HashAlgoSHA256,
				Value:     hash,
			},
		}
	case utils.SBOM:
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

	// We should have a single hash, need to make sure the current one (if exists) is the same and alert otherwise.
	if m.SrcMetaData.Component.Hashes != nil && len(*m.SrcMetaData.Component.Hashes) != 0 {
		currentSourceHash := (*m.SrcMetaData.Component.Hashes)[0].Value
		if currentSourceHash != sourceHash {
			log.Errorf("Conflicting hashes: new hash %q is different from existing hash %q", sourceHash, currentSourceHash)
		}
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

func (m *MergedResults) addSourceMetadata(metadata map[string]string) *MergedResults {
	if m == nil || m.SrcMetaData == nil || m.SrcMetaData.Component == nil {
		return m
	}

	if m.SrcMetaData.Component.Properties == nil {
		m.SrcMetaData.Component.Properties = &[]cdx.Property{}
	}

	properties := m.SrcMetaData.Component.Properties

	for key, value := range metadata {
		if key == "" || value == "" {
			continue
		}

		*properties = append(*properties, cdx.Property{
			Name:  key,
			Value: value,
		})
	}

	return m
}
