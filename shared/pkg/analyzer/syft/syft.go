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
	"fmt"

	"github.com/anchore/syft/syft"
	syft_artifact "github.com/anchore/syft/syft/artifact"
	"github.com/anchore/syft/syft/formats/common/cyclonedxhelpers"
	"github.com/anchore/syft/syft/linux"
	syft_pkg "github.com/anchore/syft/syft/pkg"
	"github.com/anchore/syft/syft/pkg/cataloger"
	syft_sbom "github.com/anchore/syft/syft/sbom"
	"github.com/anchore/syft/syft/source"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/types"
	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/image_helper"
)

const AnalyzerName = "syft"

type Analyzer struct {
	name       string
	logger     *log.Entry
	config     config.SyftConfig
	localImage bool
}

func New(conf *config.Config, logger *log.Entry) (job_manager.Job[utils.SourceInput, types.Results], error) {
	return &Analyzer{
		name:       AnalyzerName,
		logger:     logger.Dup().WithField("analyzer", AnalyzerName),
		config:     config.CreateSyftConfig(conf.Analyzer, conf.Registry),
		localImage: conf.LocalImageScan,
	}, nil
}

func (a *Analyzer) Run(sourceInput utils.SourceInput) (types.Results, error) {
	res := types.Results{}
	src := utils.CreateSource(sourceInput.Type, sourceInput.Source, a.localImage)
	a.logger.Infof("Called %s analyzer on source %s", a.name, src)

	// TODO platform can be defined
	// https://github.com/anchore/syft/blob/b20310eaf847c259beb4fe5128c842bd8aa4d4fc/cmd/syft/cli/options/packages.go#L48
	input, err := source.ParseInput(src, "", false)
	if err != nil {
		return res, fmt.Errorf("failed to create input from source analyzer=%s: %v", a.name, err)
	}
	s, _, err := source.New(*input, a.config.RegistryOptions, []string{})
	if err != nil {
		return res, fmt.Errorf("failed to create source analyzer=%s: %v", a.name, err)
	}

	catalogerConfig := cataloger.Config{
		Search: cataloger.DefaultSearchConfig(),
	}
	catalogerConfig.Search.Scope = a.config.Scope

	p, r, d, err := syft.CatalogPackages(s, catalogerConfig)
	if err != nil {
		return res, fmt.Errorf("failed to write results: %v", err)
	}
	sbom := generateSBOM(p, r, d, s)

	cdxBom := cyclonedxhelpers.ToFormatModel(sbom)
	res = types.CreateResults(cdxBom, a.name, src, sourceInput.Type)

	// Syft uses ManifestDigest to fill version information in the case of an image.
	// We need RepoDigest as well which is not set by Syft if we using cycloneDX output.
	// Get the RepoDigest from image metadata and use it as SourceHash in the Result
	// that will be added to the component hash of metadata during the merge.
	if sourceInput.Type == utils.IMAGE {
		res.AppInfo.SourceHash = getImageHash(sbom, sourceInput.Source)
	}
	a.logger.Infof("Sending successful results")

	return res, nil
}

func generateSBOM(c *syft_pkg.Catalog, r []syft_artifact.Relationship, d *linux.Release, s *source.Source) syft_sbom.SBOM {
	return syft_sbom.SBOM{
		Artifacts: syft_sbom.Artifacts{
			PackageCatalog:    c,
			LinuxDistribution: d,
		},
		Source:        s.Metadata,
		Relationships: r,
	}
}

func getImageHash(sbom syft_sbom.SBOM, src string) string {
	return image_helper.GetHashFromRepoDigest(sbom.Source.ImageMetadata.RepoDigests, src)
}
