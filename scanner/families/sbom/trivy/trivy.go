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

	cdx "github.com/CycloneDX/cyclonedx-go"
	trivyCache "github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	trivyfTypes "github.com/aquasecurity/trivy/pkg/fanal/types"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"

	"github.com/openclarity/openclarity/core/log"
	"github.com/openclarity/openclarity/scanner/common"
	"github.com/openclarity/openclarity/scanner/families"
	"github.com/openclarity/openclarity/scanner/families/sbom/trivy/config"
	"github.com/openclarity/openclarity/scanner/families/sbom/types"
	"github.com/openclarity/openclarity/scanner/families/utils/trivy"
	"github.com/openclarity/openclarity/scanner/utils/image_helper"
)

const AnalyzerName = "trivy"

type Analyzer struct {
	config config.Config
}

func New(_ context.Context, _ string, config types.Config) (families.Scanner[*types.ScannerResult], error) {
	trivyConfig := config.AnalyzersConfig.Trivy

	// Override from parent config if unset
	if trivyConfig.Registry == nil {
		trivyConfig.SetRegistry(config.Registry)
	}
	if !trivyConfig.LocalImageScan {
		trivyConfig.SetLocalImageScan(config.LocalImageScan)
	}

	return &Analyzer{
		config: trivyConfig,
	}, nil
}

// nolint:cyclop
func (a *Analyzer) Scan(ctx context.Context, sourceType common.InputType, userInput string) (*types.ScannerResult, error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	tempFile, err := os.CreateTemp(a.config.TempDir, "trivy.sbom.*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	dbOptions, err := trivy.GetTrivyDBOptions()
	if err != nil {
		return nil, fmt.Errorf("unable to get db options: %w", err)
	}

	// Skip this analyser for input types we don't support
	if !sourceType.IsOneOf(common.IMAGE, common.ROOTFS, common.DIR, common.FILE, common.DOCKERARCHIVE, common.OCIARCHIVE, common.OCIDIR) {
		return nil, fmt.Errorf("unsupported input type=%s", sourceType)
	}

	cacheDir := trivyCache.DefaultDir()
	if a.config.CacheDir != "" {
		cacheDir = a.config.CacheDir
	}

	trivyOptions := trivyFlag.Options{
		GlobalOptions: trivyFlag.GlobalOptions{
			Timeout:  a.config.GetTimeout(),
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
		ImageOptions: trivyFlag.ImageOptions{
			ImageSources: trivyfTypes.AllImageSources,
		},
		PackageOptions: trivyFlag.PackageOptions{
			PkgTypes: trivyTypes.PkgTypes, // Trivy disables analyzers for language packages if PkgTypeLibrary not in PkgType list
		},
	}

	// Convert the source to the trivy source type
	trivySourceType, err := trivy.SourceToTrivySource(sourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to configure trivy: %w", err)
	}

	// Configure Trivy image options according to the source type and user input.
	trivyOptions, cleanup, err := trivy.SetTrivyImageOptions(sourceType, userInput, trivyOptions)
	defer cleanup(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to configure trivy image options: %w", err)
	}

	// Ensure we're configured for private registry if required
	trivyOptions = trivy.SetTrivyRegistryConfigs(a.config.Registry, trivyOptions)

	err = artifact.Run(ctx, trivyOptions, trivySourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SBOM: %w", err)
	}

	// Decode the BOM
	bom := new(cdx.BOM)
	decoder := cdx.NewBOMDecoder(tempFile, cdx.BOMFileFormatJSON)
	if err = decoder.Decode(bom); err != nil {
		return nil, fmt.Errorf("unable to decode BOM data: %w", err)
	}

	result := types.CreateScannerResult(bom, AnalyzerName, userInput, sourceType)

	// Trivy doesn't include the version information in the
	// component of CycloneDX, but it does include the RepoDigest and the ImageID as
	// a property of the component.
	//
	// Get the RepoDigest/ImageID from image metadata and use it as
	// SourceHash in the Result that will be added to the component
	// hash of metadata during the merge.
	switch sourceType {
	case common.IMAGE, common.DOCKERARCHIVE, common.OCIDIR, common.OCIARCHIVE:
		hash, imageInfo, err := getImageInfo(bom.Metadata.Component.Properties, userInput)
		if err != nil {
			return nil, fmt.Errorf("failed to get image hash from sbom: %w", err)
		}

		// sync image details to result
		result.AppInfo.SourceHash = hash
		result.AppInfo.SourceMetadata = imageInfo.ToMetadata()

	case common.SBOM, common.DIR, common.ROOTFS, common.FILE, common.CSV:
		// ignore
	default:
		// ignore
	}

	return result, nil
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
