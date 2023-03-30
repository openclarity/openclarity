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
	"fmt"
	"strings"

	grype_models "github.com/anchore/grype/grype/presenter/models"
	syft_source "github.com/anchore/syft/syft/source"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/types"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/image_helper"
)

const (
	ScannerName = "grype"
)

func New(conf *config.Config, logger *log.Entry) (job_manager.Job[utils.SourceInput, types.Results], error) {
	switch conf.Scanner.GrypeConfig.Mode {
	case config.ModeLocal:
		return newLocalScanner(conf, logger)
	case config.ModeRemote:
		return newRemoteScanner(conf, logger)
	default:
		return nil, fmt.Errorf("unsupported grype mode %q", conf.Scanner.GrypeConfig.Mode)
	}
}

func CreateResults(doc grype_models.Document, userInput, scannerName, hash string) types.Results {
	distro := getDistro(doc)

	matches := make(types.Matches, len(doc.Matches))
	for i := range doc.Matches {
		match := doc.Matches[i]

		layerID, path := getLayerIDAndPath(match.Artifact.Locations)

		matches[i] = types.Match{
			Vulnerability: types.Vulnerability{
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
			},
		}
	}

	return types.Results{
		Matches: matches,
		ScannerInfo: types.Info{
			Name: scannerName,
		},
		Source: getSource(doc, userInput, hash),
	}
}

func getSource(doc grype_models.Document, userInput, hash string) types.Source {
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
		hash = image_helper.GetHashFromRepoDigest(imageMetadata.RepoDigests, userInput)
		if hash == "" {
			// set hash using ManifestDigest if RepoDigest is missing
			manifestHash := imageMetadata.ManifestDigest
			if idx := strings.Index(manifestHash, ":"); idx != -1 {
				hash = manifestHash[idx+1:]
			}
		}
	case string:
		srcName = doc.Source.Target.(string) // nolint:forcetypeassert
	}

	return types.Source{
		Type: doc.Source.Type,
		Name: srcName,
		Hash: hash,
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
	if len(match.RelatedVulnerabilities) != 0 {
		return match.RelatedVulnerabilities[0].Cvss
	}

	return match.Vulnerability.VulnerabilityMetadata.Cvss
}

func getDescription(match grype_models.Match) string {
	if len(match.RelatedVulnerabilities) != 0 {
		return match.RelatedVulnerabilities[0].Description
	}

	return match.Vulnerability.Description
}

// nolint:nonamedreturns
func getLayerIDAndPath(coordinates []syft_source.Coordinates) (layerID, path string) {
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
