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

package trivy

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openclarity/vmclarity/scanner/utils"

	"github.com/aquasecurity/trivy/pkg/fanal/types"
	log "github.com/sirupsen/logrus"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	trivyFsutils "github.com/aquasecurity/trivy/pkg/utils/fsutils"

	"github.com/openclarity/vmclarity/scanner/analyzer"
	"github.com/openclarity/vmclarity/scanner/config"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/utils/image_helper"
	"github.com/openclarity/vmclarity/scanner/utils/trivy"
)

const AnalyzerName = "trivy"

type Analyzer struct {
	name       string
	logger     *log.Entry
	config     config.AnalyzerTrivyConfigEx
	resultChan chan job_manager.Result
	localImage bool
}

func New(_ string, c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(*config.Config) // nolint:forcetypeassert
	return &Analyzer{
		name:       AnalyzerName,
		logger:     logger.Dup().WithField("analyzer", AnalyzerName),
		config:     config.CreateAnalyzerTrivyConfigEx(conf.Analyzer, conf.Registry),
		resultChan: resultChan,
		localImage: conf.LocalImageScan,
	}
}

// nolint:cyclop
func (a *Analyzer) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	a.logger.Infof("Called %s analyzer on source %v %v", a.name, sourceType, userInput)

	tempFile, err := os.CreateTemp(a.config.TempDir, "trivy.sbom.*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	dbOptions, err := trivy.GetTrivyDBOptions()
	if err != nil {
		return fmt.Errorf("unable to get db options: %w", err)
	}

	go func(ctx context.Context) {
		defer os.Remove(tempFile.Name())

		res := &analyzer.Results{}

		// Skip this analyser for input types we don't support
		switch sourceType {
		case utils.IMAGE, utils.ROOTFS, utils.DIR, utils.FILE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
			// These are all supported for SBOM analysing so continue
		case utils.SBOM:
			fallthrough
		default:
			a.logger.Infof("Skipping analyze unsupported source type: %s", sourceType)
			a.resultChan <- res
			return
		}

		cacheDir := trivyFsutils.CacheDir()
		if a.config.CacheDir != "" {
			cacheDir = a.config.CacheDir
		}

		trivyOptions := trivyFlag.Options{
			GlobalOptions: trivyFlag.GlobalOptions{
				Timeout:  a.config.Timeout,
				CacheDir: cacheDir,
			},
			ScanOptions: trivyFlag.ScanOptions{
				Target:   userInput,
				Scanners: []trivyTypes.Scanner{}, // Disable all security checks for SBOM only scan
			},
			ReportOptions: trivyFlag.ReportOptions{
				Format:       trivyTypes.FormatCycloneDX, // Cyclonedx format for SBOM so that we don't need to convert
				ReportFormat: "all",                      // Full report not just summary
				Output:       tempFile.Name(),            // Save the output to our temp file instead of Stdout
				ListAllPkgs:  true,                       // By default, Trivy only includes packages with vulnerabilities, for full SBOM set true.
			},
			DBOptions: dbOptions,
			VulnerabilityOptions: trivyFlag.VulnerabilityOptions{
				VulnType: trivyTypes.VulnTypes, // Trivy disables analyzers for language packages if VulnTypeLibrary not in VulnType list
			},
			ImageOptions: trivyFlag.ImageOptions{
				ImageSources: types.AllImageSources,
			},
		}

		// Convert the source to the trivy source type
		trivySourceType, err := trivy.SourceToTrivySource(sourceType)
		if err != nil {
			a.setError(res, fmt.Errorf("failed to configure trivy: %w", err))
			return
		}

		// Configure Trivy image options according to the source type and user input.
		trivyOptions, cleanup, err := trivy.SetTrivyImageOptions(sourceType, userInput, trivyOptions)
		defer cleanup(a.logger)
		if err != nil {
			a.setError(res, fmt.Errorf("failed to configure trivy image options: %w", err))
			return
		}

		// Ensure we're configured for private registry if required
		trivyOptions = trivy.SetTrivyRegistryConfigs(a.config.Registry, trivyOptions)

		err = artifact.Run(ctx, trivyOptions, trivySourceType)
		if err != nil {
			a.setError(res, fmt.Errorf("failed to generate SBOM: %w", err))
			return
		}

		// Decode the BOM
		bom := new(cdx.BOM)
		decoder := cdx.NewBOMDecoder(tempFile, cdx.BOMFileFormatJSON)
		if err = decoder.Decode(bom); err != nil {
			a.setError(res, fmt.Errorf("unable to decode BOM data: %w", err))
			return
		}

		res = analyzer.CreateResults(bom, a.name, userInput, sourceType)

		// Trivy doesn't include the version information in the
		// component of CycloneDX, but it does include the RepoDigest and the ImageID as
		// a property of the component.
		//
		// Get the RepoDigest/ImageID from image metadata and use it as
		// SourceHash in the Result that will be added to the component
		// hash of metadata during the merge.
		switch sourceType {
		case utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIDIR, utils.OCIARCHIVE:
			hash, imageInfo, err := getImageInfo(bom.Metadata.Component.Properties, userInput)
			if err != nil {
				a.setError(res, fmt.Errorf("failed to get image hash from sbom: %w", err))
				return
			}

			// sync image details to result
			res.AppInfo.SourceHash = hash
			res.AppInfo.SourceMetadata = imageInfo.ToMetadata()

		case utils.SBOM, utils.DIR, utils.ROOTFS, utils.FILE:
			// ignore
		default:
			// ignore
		}

		a.logger.Infof("Sending successful results")
		a.resultChan <- res
	}(ctx)

	return nil
}

func (a *Analyzer) setError(res *analyzer.Results, err error) {
	res.Error = err
	a.logger.Error(res.Error)
	a.resultChan <- res
}

func getImageInfo(properties *[]cdx.Property, imageName string) (string, *image_helper.ImageInfo, error) {
	if properties == nil {
		return "", nil, errors.New("properties was nil")
	}

	var repoDigests []string
	var repoTags []string
	var imageID string

	for _, property := range *properties {
		switch property.Name {
		case "aquasecurity:trivy:ImageID":
			imageID = property.Value
		case "aquasecurity:trivy:RepoDigest":
			repoDigests = append(repoDigests, property.Value)
		case "aquasecurity:trivy:RepoTag":
			repoTags = append(repoTags, property.Value)
		default:
			// Ignore property
		}
	}

	imageInfo := &image_helper.ImageInfo{
		Name:    imageName,
		ID:      imageID,
		Digests: repoDigests,
		Tags:    repoTags,
	}

	hash, err := imageInfo.GetHashFromRepoDigestsOrImageID()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get image hash from repo digests or image id: %w", err)
	}

	return hash, imageInfo, nil
}
