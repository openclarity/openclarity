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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	trivyDBTypes "github.com/aquasecurity/trivy-db/pkg/types"
	trivyCache "github.com/aquasecurity/trivy/pkg/cache"
	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	trivyfTypes "github.com/aquasecurity/trivy/pkg/fanal/types"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	trivyLog "github.com/aquasecurity/trivy/pkg/log"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	sloglogrus "github.com/samber/slog-logrus/v2"
	"github.com/sirupsen/logrus"

	"github.com/openclarity/openclarity/core/log"
	"github.com/openclarity/openclarity/scanner/common"
	"github.com/openclarity/openclarity/scanner/families"
	utilsTrivy "github.com/openclarity/openclarity/scanner/families/utils/trivy"
	"github.com/openclarity/openclarity/scanner/families/vulnerabilities/trivy/config"
	"github.com/openclarity/openclarity/scanner/families/vulnerabilities/types"
	"github.com/openclarity/openclarity/scanner/utils/image_helper"
	"github.com/openclarity/openclarity/scanner/utils/sbom"
	utilsVul "github.com/openclarity/openclarity/scanner/utils/vulnerability"
)

const ScannerName = "trivy"

type Scanner struct {
	config config.Config
}

func New(ctx context.Context, _ string, config types.Config) (families.Scanner[*types.ScannerResult], error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	// Set up the logger for trivy
	tlogger := trivyLog.New(sloglogrus.Option{Logger: logger.Logger}.NewLogrusHandler())
	trivyLog.SetDefault(tlogger)

	// Override configs from parent
	trivyConfig := config.ScannersConfig.Trivy
	if trivyConfig.Registry == nil {
		trivyConfig.SetRegistry(config.Registry)
	}

	return &Scanner{
		config: trivyConfig,
	}, nil
}

// nolint:cyclop
func (a *Scanner) Scan(ctx context.Context, sourceType common.InputType, userInput string) (*types.ScannerResult, error) {
	if !sourceType.IsOneOf(common.SBOM) {
		return nil, fmt.Errorf("unsupported input type=%s", sourceType)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx)

	tempFile, err := os.CreateTemp(a.config.CacheDir, "trivy.scan.*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	bom, err := sbom.NewCycloneDX(userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create CycloneDX SBOM: %w", err)
	}

	metadata := bom.GetMetadataFromSBOM()
	hash, err := bom.GetHashFromSBOM()
	if err != nil {
		return nil, fmt.Errorf("failed to get original hash from SBOM: %w", err)
	}

	trivyOptions, err := a.createTrivyOptions(tempFile.Name(), userInput)
	if err != nil {
		return nil, fmt.Errorf("unable to create trivy options: %w", err)
	}

	// Configure Trivy image options according to the source type and user input.
	trivyOptions, cleanup, err := utilsTrivy.SetTrivyImageOptions(sourceType, userInput, trivyOptions)
	defer cleanup(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to configure trivy image options: %w", err)
	}

	// Convert the source to the trivy source type
	trivySourceType, err := utilsTrivy.SourceToTrivySource(sourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to configure trivy: %w", err)
	}

	// Ensure we're configured for private registry if required
	trivyOptions = utilsTrivy.SetTrivyRegistryConfigs(a.config.Registry, trivyOptions)

	err = artifact.Run(ctx, trivyOptions, trivySourceType)
	if err != nil {
		return nil, fmt.Errorf("failed to scan for vulnerabilities: %w", err)
	}

	file, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read scan results: %w", err)
	}

	result, err := a.createResult(logger, file, hash, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create result: %w", err)
	}

	return result, nil
}

// nolint:cyclop
func (a *Scanner) createResult(logger *logrus.Entry, trivyJSON []byte, hash string, metadata map[string]string) (*types.ScannerResult, error) {
	if len(trivyJSON) == 0 {
		return nil, errors.New("cannot produce result for empty trivy JSON")
	}

	var report trivyTypes.Report
	err := json.Unmarshal(trivyJSON, &report)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal trivy json: %w", err)
	}

	vulnerabilities := []types.Vulnerability{}
	for _, result := range report.Results {
		for _, vul := range result.Vulnerabilities {
			typ, err := getTypeFromPurl(vul.PkgIdentifier.BOMRef)
			if err != nil {
				logger.Error(err)
				typ = ""
			}

			cvsses := getCVSSesFromVul(vul.CVSS)

			fix := types.Fix{}
			if vul.FixedVersion != "" {
				fix.Versions = []string{
					vul.FixedVersion,
				}
			}

			distro := types.Distro{}
			if report.Metadata.OS != nil {
				// Trivy calls the distro name (ubuntu, debian, alpine) the family
				distro.Name = string(report.Metadata.OS.Family)
				// Trivy calls the version (11, hardy heron, 22.04) the name
				distro.Version = report.Metadata.OS.Name
			}

			links := make([]string, 0, len(vul.Vulnerability.References))
			links = append(links, vul.Vulnerability.References...)
			vulnerability := types.Vulnerability{
				ID:          vul.VulnerabilityID,
				Description: vul.Description,
				Links:       links,
				Distro:      distro,
				CVSS:        cvsses,
				Fix:         fix,
				Severity:    strings.ToUpper(vul.Severity),
				Package: types.Package{
					Name:    vul.PkgName,
					Version: vul.InstalledVersion,
					PURL:    vul.PkgIdentifier.BOMRef,
					Type:    typ,
					// TODO(sambetts) Trivy doesn't pass
					// through this info from the SBOM so
					// we might need to fill this out
					// ourselves.
					Language: "",
					Licenses: nil,
					CPEs:     nil,
				},
				LayerID: vul.Layer.Digest,
				Path:    vul.PkgPath,
			}

			vulnerabilities = append(vulnerabilities, vulnerability)
		}
	}

	logger.Infof("Found %d vulnerabilities", len(vulnerabilities))

	if hash == "" && string(report.ArtifactType) == "container_image" {
		// empty hash indicates empty image details, recreate hash and image info from artifact
		imageInfo := image_helper.ImageInfo{
			Name:    report.ArtifactName,
			ID:      report.Metadata.ImageID,
			Tags:    report.Metadata.RepoTags,
			Digests: report.Metadata.RepoDigests,
		}

		metadata = imageInfo.ToMetadata()
		hash, err = imageInfo.GetHashFromRepoDigestsOrImageID()
		if err != nil {
			logger.Warningf("Failed to get image hash from repo digests or image id: %v", err)
		}
	}

	return &types.ScannerResult{
		Vulnerabilities: vulnerabilities,
		Source: types.Source{
			Name:     report.ArtifactName,
			Type:     string(report.ArtifactType),
			Hash:     hash,
			Metadata: metadata,
		},
		Scanner: types.ScannerInfo{
			Name: ScannerName,
		},
	}, nil
}

func (a *Scanner) createTrivyOptions(output string, userInput string) (trivyFlag.Options, error) {
	severities, err := getAllTrivySeverities()
	if err != nil {
		return trivyFlag.Options{}, fmt.Errorf("unable to get all trivy severities: %w", err)
	}

	cacheDir := trivyCache.DefaultDir()
	if a.config.CacheDir != "" {
		cacheDir = a.config.CacheDir
	}

	dbOptions, err := utilsTrivy.GetTrivyDBOptions()
	if err != nil {
		return trivyFlag.Options{}, fmt.Errorf("unable to get db options: %w", err)
	}

	trivyOptions := trivyFlag.Options{
		GlobalOptions: trivyFlag.GlobalOptions{
			Timeout:  a.config.GetTimeout(),
			CacheDir: cacheDir,
		},
		ScanOptions: trivyFlag.ScanOptions{
			Target: userInput,
			Scanners: []trivyTypes.Scanner{
				trivyTypes.VulnerabilityScanner, // Enable just vuln scanning
			},
		},
		ReportOptions: trivyFlag.ReportOptions{
			Format:       trivyTypes.FormatJSON, // Trivy's own json format is the most complete for vuls
			ReportFormat: "all",                 // Full report not just summary
			Output:       output,                // Save the output to our temp file instead of Stdout
			ListAllPkgs:  false,                 // Only include packages with vulnerabilities
			Severities:   severities,            // All the severities from the above
		},
		DBOptions: dbOptions,
		ImageOptions: trivyFlag.ImageOptions{
			ImageSources: trivyfTypes.AllImageSources,
		},
		PackageOptions: trivyFlag.PackageOptions{
			PkgTypes: trivyTypes.PkgTypes, // Trivy disables analyzers for language packages if PkgTypeLibrary not in PkgType list
		},
	}

	// If provided use the trivy server mode
	if a.config.ServerAddr != "" {
		// trivy needs the token specified in both the Token
		// field and in the CustomHeaders field of the
		// RemoteOptions
		customHeaders := http.Header{}
		if a.config.ServerToken != "" {
			customHeaders.Set(trivyFlag.DefaultTokenHeader, a.config.ServerToken)
		}

		trivyOptions.RemoteOptions = trivyFlag.RemoteOptions{
			ServerAddr:    a.config.ServerAddr,
			Token:         a.config.ServerToken,
			TokenHeader:   trivyFlag.DefaultTokenHeader,
			CustomHeaders: customHeaders,
		}
	}

	return trivyOptions, nil
}

func getCVSSesFromVul(vCvss trivyDBTypes.VendorCVSS) []types.CVSS {
	// TODO(sambetts) Work out how to include all the
	// CVSS's by vendor ID in our report, for now only
	// collect one of each version.
	cvsses := []types.CVSS{}
	v2Collected := false
	v3Collected := false

	// Collect all the vendors from the trivy type and sort them so that we
	// have predictable results from this function as we're only taking one
	// of each.
	vendors := make([]string, 0, len(vCvss))
	for v := range vCvss {
		vendors = append(vendors, string(v))
	}
	sort.Strings(vendors)

	for _, vendor := range vendors {
		cvss := vCvss[trivyDBTypes.SourceID(vendor)]
		if cvss.V3Vector != "" && !v3Collected {
			exploit, impact := utilsVul.ExploitScoreAndImpactScoreFromV3Vector(cvss.V3Vector)

			cvsses = append(cvsses, types.CVSS{
				Version: utilsVul.GetCVSSV3VersionFromVector(cvss.V3Vector),
				Vector:  cvss.V3Vector,
				Metrics: types.CvssMetrics{
					BaseScore:           cvss.V3Score,
					ExploitabilityScore: &exploit,
					ImpactScore:         &impact,
				},
			})
			v3Collected = true
		}
		if cvss.V2Vector != "" && !v2Collected {
			exploit, impact := utilsVul.ExploitScoreAndImpactScoreFromV2Vector(cvss.V2Vector)

			cvsses = append(cvsses, types.CVSS{
				Version: "2.0",
				Vector:  cvss.V2Vector,
				Metrics: types.CvssMetrics{
					BaseScore:           cvss.V2Score,
					ExploitabilityScore: &exploit,
					ImpactScore:         &impact,
				},
			})
			v2Collected = true
		}
	}
	return cvsses
}

func getTypeFromPurl(purl string) (string, error) {
	u, err := url.Parse(purl)
	if err != nil {
		return "", fmt.Errorf("unable to parse purl: %w", err)
	}

	typ, _, found := strings.Cut(u.Opaque, "/")
	if !found {
		return "", errors.New("type not found in purl")
	}

	return typ, nil
}

func getAllTrivySeverities() ([]trivyDBTypes.Severity, error) {
	// Build a list of CVE severities for the trivy scanner to
	// report.  trivyDBTypes.SeverityNames contains all the
	// severities that trivy supports and we want them all in our
	// report at the moment.
	severities := []trivyDBTypes.Severity{}
	for _, name := range trivyDBTypes.SeverityNames {
		sev, err := trivyDBTypes.NewSeverity(strings.ToUpper(name))
		if err != nil {
			return nil, fmt.Errorf("unable to get trivy severity %s: %w", name, err)
		}
		severities = append(severities, sev)
	}
	return severities, nil
}
