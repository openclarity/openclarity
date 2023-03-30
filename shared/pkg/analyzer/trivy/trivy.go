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
	"bytes"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	trivyUtils "github.com/aquasecurity/trivy/pkg/utils"

	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/types"
	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/image_helper"
	utilsTrivy "github.com/openclarity/kubeclarity/shared/pkg/utils/trivy"
)

const AnalyzerName = "trivy"

type Analyzer struct {
	name       string
	logger     *log.Entry
	config     config.AnalyzerTrivyConfigEx
	localImage bool
}

func New(conf *config.Config, logger *log.Entry) (job_manager.Job[utils.SourceInput, types.Results], error) {
	return &Analyzer{
		name:       AnalyzerName,
		logger:     logger.Dup().WithField("analyzer", AnalyzerName),
		config:     config.CreateAnalyzerTrivyConfigEx(conf.Analyzer, conf.Registry),
		localImage: conf.LocalImageScan,
	}, nil
}

// nolint:cyclop
func (a *Analyzer) Run(sourceInput utils.SourceInput) (types.Results, error) {
	a.logger.Infof("Called %s analyzer on source %v %v", a.name, sourceInput.Type, sourceInput.Source)
	res := types.Results{}

	// Skip this analyser for input types we don't support
	switch sourceInput.Type {
	case utils.IMAGE, utils.ROOTFS, utils.DIR, utils.FILE:
		// These are all supported for SBOM analysing so continue
	case utils.SBOM:
		fallthrough
	default:
		a.logger.Infof("Skipping analyze unsupported source type: %s", sourceInput.Type)
		return res, nil
	}

	cacheDir := trivyUtils.DefaultCacheDir()
	if a.config.CacheDir != "" {
		cacheDir = a.config.CacheDir
	}

	var output bytes.Buffer
	trivyOptions := trivyFlag.Options{
		GlobalOptions: trivyFlag.GlobalOptions{
			Timeout:  a.config.Timeout,
			CacheDir: cacheDir,
		},
		ScanOptions: trivyFlag.ScanOptions{
			Target:         sourceInput.Source,
			SecurityChecks: nil, // Disable all security checks for SBOM only scan
		},
		ReportOptions: trivyFlag.ReportOptions{
			Format:       "cyclonedx", // Cyconedx format for SBOM so that we don't need to convert
			ReportFormat: "all",       // Full report not just summary
			Output:       &output,     // Save the output to our local buffer instead of Stdout
			ListAllPkgs:  true,        // By default Trivy only includes packages with vulnerabilities, for full SBOM set true.
		},
		VulnerabilityOptions: trivyFlag.VulnerabilityOptions{
			VulnType: trivyTypes.VulnTypes, // Trivy disables analyzers for language packages if VulnTypeLibrary not in VulnType list
		},
	}

	// Convert the kubeclarity source to the trivy source type
	trivySourceType, err := utilsTrivy.KubeclaritySourceToTrivySource(sourceInput.Type)
	if err != nil {
		return res, fmt.Errorf("failed to configure trivy: %w", err)
	}

	// Ensure we're configured for private registry if required
	trivyOptions = utilsTrivy.SetTrivyRegistryConfigs(a.config.Registry, trivyOptions)

	err = artifact.Run(context.TODO(), trivyOptions, trivySourceType)
	if err != nil {
		return res, fmt.Errorf("failed to generate SBOM: %w", err)
	}

	// Decode the BOM
	bom := new(cdx.BOM)
	decoder := cdx.NewBOMDecoder(&output, cdx.BOMFileFormatJSON)
	if err = decoder.Decode(bom); err != nil {
		return res, fmt.Errorf("unable to decode BOM data: %v", err)
	}

	res = types.CreateResults(bom, a.name, sourceInput.Source, sourceInput.Type)

	// Trivy doesn't include the version information in the
	// component of CycloneDX but it does include the RepoDigest as
	// a property of the component.
	//
	// Get the RepoDigest from image metadata and use it as
	// SourceHash in the Result that will be added to the component
	// hash of metadata during the merge.
	if sourceInput.Type == utils.IMAGE {
		hash, err := getImageHash(bom.Metadata.Component.Properties, sourceInput.Source)
		if err != nil {
			return res, fmt.Errorf("failed to get image hash from sbom: %w", err)
		}
		res.AppInfo.SourceHash = hash
	}

	a.logger.Infof("Sending successful results")

	return res, nil
}

func getImageHash(properties *[]cdx.Property, src string) (string, error) {
	if properties == nil {
		return "", fmt.Errorf("properties was nil")
	}

	for _, property := range *properties {
		if property.Name == "aquasecurity:trivy:RepoDigest" {
			return image_helper.GetHashFromRepoDigest([]string{property.Value}, src), nil
		}
	}

	return "", fmt.Errorf("repo digest property missing from Metadata.Component")
}
