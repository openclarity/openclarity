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

package grype

import (
	"context"
	"fmt"
	"strings"

	grype_models "github.com/anchore/grype/grype/presenter/models"
	"github.com/anchore/syft/syft/file"
	syft_source "github.com/anchore/syft/syft/source"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/families"
	grypeconfig "github.com/openclarity/vmclarity/scanner/families/vulnerabilities/grype/config"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"
	"github.com/openclarity/vmclarity/scanner/utils/image_helper"
)

const (
	ScannerName = "grype"
)

func New(_ context.Context, _ string, config types.Config) (families.Scanner[*types.ScannerResult], error) {
	switch mode := config.ScannersConfig.Grype.Mode; mode {
	case grypeconfig.ModeLocal:
		return newLocalScanner(config), nil
	case grypeconfig.ModeRemote:
		return newRemoteScanner(config), nil
	default:
		return nil, fmt.Errorf("unsupported grype mode %q", mode)
	}
}

func createResults(doc grype_models.Document, userInput, scannerName, hash string, metadata map[string]string) *types.ScannerResult {
	distro := getDistro(doc)

	vulnerabilities := make([]types.Vulnerability, len(doc.Matches))
	for i := range doc.Matches {
		match := doc.Matches[i]

		layerID, path := getLayerIDAndPath(match.Artifact.Locations)

		vulnerabilities[i] = types.Vulnerability{
			ID:          match.Vulnerability.ID,
			Description: getDescription(match),
			Links:       match.Vulnerability.URLs,
			Distro:      distro,
			CVSS:        getCVSS(match),
			Fix: types.Fix{
				Versions: match.Vulnerability.Fix.Versions,
				State:    match.Vulnerability.Fix.State,
			},
			Severity: strings.ToUpper(match.Vulnerability.Severity),
			Package: types.Package{
				Name:     match.Artifact.Name,
				Version:  match.Artifact.Version,
				Type:     string(match.Artifact.Type),
				Language: string(match.Artifact.Language),
				Licenses: match.Artifact.Licenses,
				CPEs:     match.Artifact.CPEs,
				PURL:     match.Artifact.PURL,
			},
			LayerID: layerID,
			Path:    path,
		}
	}

	source := getSource(doc, userInput, hash, metadata)

	return types.NewScannerResult(scannerName, source, vulnerabilities)
}

func getSource(doc grype_models.Document, userInput, hash string, metadata map[string]string) types.Source {
	var source types.Source
	if doc.Source == nil {
		return source
	}

	var srcName string
	switch doc.Source.Target.(type) {
	case syft_source.ImageMetadata:
		imageMetadata := doc.Source.Target.(syft_source.ImageMetadata) // nolint:forcetypeassert
		srcName = imageMetadata.UserInput
		// If the userInput is a SBOM, the srcName and hash will be got from the SBOM.
		if srcName == "" {
			srcName = userInput
		}
		if hash != "" {
			break
		}

		imageInfo := image_helper.ImageInfo{
			Name:    userInput,
			ID:      imageMetadata.ID,
			Tags:    imageMetadata.Tags,
			Digests: imageMetadata.RepoDigests,
		}
		if h, err := imageInfo.GetHashFromRepoDigestsOrImageID(); err != nil {
			log.Warningf("Failed to get image hash from repo digests or image id: %v", err)
		} else {
			hash = h
			metadata = imageInfo.ToMetadata()
		}
	case string:
		srcName = doc.Source.Target.(string) // nolint:forcetypeassert
	}

	return types.Source{
		Metadata: metadata,
		Type:     doc.Source.Type,
		Name:     srcName,
		Hash:     hash,
	}
}

func getDistro(doc grype_models.Document) types.Distro {
	return types.Distro{
		Name:    doc.Distro.Name,
		Version: doc.Distro.Version,
		IDLike:  doc.Distro.IDLike,
	}
}

func getCVSS(match grype_models.Match) []types.CVSS {
	cvssFromMatch := getCVSSFromMatch(match)
	if len(cvssFromMatch) == 0 {
		return nil
	}

	ret := make([]types.CVSS, len(cvssFromMatch))
	for i := range cvssFromMatch {
		cvss := cvssFromMatch[i]
		ret[i] = types.CVSS{
			Source:  cvss.Source,
			Type:    cvss.Type,
			Version: cvss.Version,
			Vector:  cvss.Vector,
			Metrics: types.CvssMetrics{
				BaseScore:           cvss.Metrics.BaseScore,
				ExploitabilityScore: cvss.Metrics.ExploitabilityScore,
				ImpactScore:         cvss.Metrics.ImpactScore,
			},
			VendorMetadata: cvss.VendorMetadata,
		}
	}

	return ret
}

func getCVSSFromMatch(match grype_models.Match) []grype_models.Cvss {
	// Due to NormalizeByCVE mode we prefer to get the info from the root if exists.
	if len(match.Vulnerability.VulnerabilityMetadata.Cvss) != 0 {
		return match.Vulnerability.VulnerabilityMetadata.Cvss
	}

	if len(match.RelatedVulnerabilities) != 0 && len(match.RelatedVulnerabilities[0].Cvss) > 0 {
		return match.RelatedVulnerabilities[0].Cvss
	}

	return match.Vulnerability.VulnerabilityMetadata.Cvss
}

func getDescription(match grype_models.Match) string {
	// Due to NormalizeByCVE mode we prefer to get the info from the root if exists.
	if match.Vulnerability.Description != "" {
		return match.Vulnerability.Description
	}

	if len(match.RelatedVulnerabilities) != 0 {
		return match.RelatedVulnerabilities[0].Description
	}

	return match.Vulnerability.Description
}

// nolint:nonamedreturns
func getLayerIDAndPath(coordinates []file.Coordinates) (layerID, path string) {
	if len(coordinates) == 0 {
		return "", ""
	}

	// The vulnerability can consist of several files (locations) related to the package from several layers.
	// We'll take the last, according to the last layer that is related to the vulnerable package.
	coordinate := coordinates[len(coordinates)-1]
	return parseLayerHex(coordinate.FileSystemID), coordinate.RealPath
}

func parseLayerHex(layerID string) string {
	index := strings.LastIndexByte(layerID, ':')
	if index == -1 {
		return layerID
	}

	return layerID[index+1:]
}
