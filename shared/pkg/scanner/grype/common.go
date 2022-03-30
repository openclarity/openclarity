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
	"errors"
	"fmt"
	"os"
	"strings"

	grype_models "github.com/anchore/grype/grype/presenter/models"
	syft_source "github.com/anchore/syft/syft/source"
	log "github.com/sirupsen/logrus"

	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/config"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/converter"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/job_manager"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/scanner"
	"wwwin-github.cisco.com/eti/scan-gazr/shared/pkg/utils/image_helper"
)

const (
	ScannerName = "grype"
)

func New(conf *config.Config, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	switch conf.Scanner.GrypeConfig.Mode {
	case config.ModeLocal:
		return newLocalScanner(conf, logger, resultChan)
	case config.ModeRemote:
		return newRemoteScanner(conf, logger, resultChan)
	}

	// We shouldn't get here since grype mode was already validated.
	log.Fatalf("Unsupported grype mode %q.", conf.Scanner.GrypeConfig.Mode)
	return nil
}

func ReportError(resultChan chan job_manager.Result, err error, logger *log.Entry) {
	res := &scanner.Results{
		Error: err,
	}

	logger.Error(res.Error)
	resultChan <- res
}

func ConvertCycloneDXFileToSyftJSONFile(inputFilePath string, logger *log.Entry) (outputFilePath string, cleanup func(), err error) {
	outputFilePath = inputFilePath + ".syft.json"
	logger.Infof("Converting %q to syft format.", inputFilePath)

	if err = converter.ConvertCycloneDXToSyftJSONFromFile(inputFilePath, outputFilePath); err != nil {
		if errors.Is(err, converter.ErrFailedToGetCycloneDXSBOM) {
			logger.Infof("Not a CycloneDX input - returning current input.")
			return inputFilePath, func() {}, nil
		}

		return "", nil, fmt.Errorf("failed to convert sbom file: %w", err)
	}

	logger.Infof("Conversion succeeded. outputFilePath=%v", outputFilePath)

	return outputFilePath, func() { _ = os.Remove(outputFilePath) }, nil
}

func CreateResults(doc grype_models.Document, userInput, scannerName string) *scanner.Results {
	distro := getDistro(doc)

	matches := make(scanner.Matches, len(doc.Matches))
	for i := range doc.Matches {
		match := doc.Matches[i]

		layerID, path := getLayerIDAndPath(match.Artifact.Locations)

		matches[i] = scanner.Match{
			Vulnerability: scanner.Vulnerability{
				ID:          match.Vulnerability.ID,
				Description: getDescription(match),
				Links:       match.Vulnerability.URLs,
				Distro:      distro,
				CVSS:        getCVSS(match),
				Fix: scanner.Fix{
					Versions: match.Vulnerability.Fix.Versions,
					State:    match.Vulnerability.Fix.State,
				},
				Severity: strings.ToUpper(match.Vulnerability.Severity),
				Package: scanner.Package{
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

	return &scanner.Results{
		Matches: matches,
		ScannerInfo: scanner.Info{
			Name: scannerName,
		},
		Source: getSource(doc, userInput),
	}
}

func getSource(doc grype_models.Document, userInput string) scanner.Source {
	var source scanner.Source
	if doc.Source == nil {
		return source
	}

	var srcName, hash string
	switch doc.Source.Target.(type) {
	case syft_source.ImageMetadata:
		srcName = doc.Source.Target.(syft_source.ImageMetadata).UserInput
		hash = image_helper.GetHashFromRepoDigest(doc.Source.Target.(syft_source.ImageMetadata).RepoDigests, userInput)
		if hash == "" {
			// set hash using ManifestDigest if RepoDigest is missing
			hash = doc.Source.Target.(syft_source.ImageMetadata).ManifestDigest
		}
	case string:
		srcName = doc.Source.Target.(string) // nolint:forcetypeassert
	}

	return scanner.Source{
		Type: doc.Source.Type,
		Name: srcName,
		Hash: hash,
	}
}

func getDistro(doc grype_models.Document) scanner.Distro {
	return scanner.Distro{
		Name:    doc.Distro.Name,
		Version: doc.Distro.Version,
		IDLike:  doc.Distro.IDLike,
	}
}

func getCVSS(match grype_models.Match) []scanner.CVSS {
	cvssFromMatch := getCVSSFromMatch(match)
	if len(cvssFromMatch) == 0 {
		return nil
	}

	ret := make([]scanner.CVSS, len(cvssFromMatch))
	for i := range cvssFromMatch {
		cvss := cvssFromMatch[i]
		ret[i] = scanner.CVSS{
			Version: cvss.Version,
			Vector:  cvss.Vector,
			Metrics: scanner.CvssMetrics{
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
