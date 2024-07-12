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

package syft

import (
	"context"
	"errors"
	"fmt"

	"github.com/anchore/syft/syft"
	"github.com/anchore/syft/syft/cataloging"
	"github.com/anchore/syft/syft/format/common/cyclonedxhelpers"
	syftsbom "github.com/anchore/syft/syft/sbom"
	syftsrc "github.com/anchore/syft/syft/source"

	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/sbom/syft/config"
	"github.com/openclarity/vmclarity/scanner/families/sbom/types"
	"github.com/openclarity/vmclarity/scanner/utils/image_helper"
)

const AnalyzerName = "syft"

type Analyzer struct {
	config config.Config
}

func New(_ context.Context, _ string, config types.Config) (families.Scanner[*types.ScannerResult], error) {
	syftConfig := config.AnalyzersConfig.Syft

	// Override from parent config if unset
	if syftConfig.Registry == nil {
		syftConfig.SetRegistry(config.Registry)
	}
	if !syftConfig.LocalImageScan {
		syftConfig.SetLocalImageScan(config.LocalImageScan)
	}

	return &Analyzer{
		config: syftConfig,
	}, nil
}

func (a *Analyzer) Scan(ctx context.Context, sourceType common.InputType, userInput string) (*types.ScannerResult, error) {
	// TODO platform can be defined
	// https://github.com/anchore/syft/blob/b20310eaf847c259beb4fe5128c842bd8aa4d4fc/cmd/syft/cli/options/packages.go#L48
	source, err := syft.GetSource(
		ctx,
		userInput,
		syft.DefaultGetSourceConfig().
			WithSources(sourceType.GetSource(a.config.LocalImageScan)).
			WithRegistryOptions(a.config.GetRegistryOptions()).
			WithExcludeConfig(a.config.GetExcludePaths()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create source analyzer=%s: %w", AnalyzerName, err)
	}

	sbomConfig := syft.DefaultCreateSBOMConfig().
		WithSearchConfig(cataloging.DefaultSearchConfig().WithScope(a.config.GetScope()))

	sbom, err := syft.CreateSBOM(ctx, source, sbomConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to write results: %w", err)
	}

	cdxBom := cyclonedxhelpers.ToFormatModel(*sbom)
	result := types.CreateScannerResult(cdxBom, AnalyzerName, userInput, sourceType)

	// Syft uses ManifestDigest to fill version information in the case of an image.
	// We need RepoDigest/ImageID as well which is not set by Syft if we're using cycloneDX output.
	// Get the RepoDigest/ImageID from image metadata and use it as SourceHash in the Result
	// that will be added to the component hash of metadata during the merge.
	switch sourceType {
	case common.IMAGE, common.DOCKERARCHIVE, common.OCIDIR, common.OCIARCHIVE:
		hash, imageInfo, err := getImageInfo(sbom, userInput)
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

func getImageInfo(s *syftsbom.SBOM, src string) (string, *image_helper.ImageInfo, error) {
	switch metadata := s.Source.Metadata.(type) {
	case syftsrc.ImageMetadata:
		imageInfo := &image_helper.ImageInfo{
			Name:    src,
			ID:      metadata.ID,
			Tags:    metadata.Tags,
			Digests: metadata.RepoDigests,
		}

		hash, err := imageInfo.GetHashFromRepoDigestsOrImageID()
		if err != nil {
			return "", nil, fmt.Errorf("failed to get image hash from repo digests or image id: %w", err)
		}

		return hash, imageInfo, nil
	default:
		return "", nil, errors.New("failed to get image hash from source metadata")
	}
}
