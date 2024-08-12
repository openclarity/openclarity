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

	"github.com/openclarity/vmclarity/scanner/utils"

	trivyDBTypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	"github.com/aquasecurity/trivy/pkg/fanal/types"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	trivyLog "github.com/aquasecurity/trivy/pkg/log"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"
	trivyFsutils "github.com/aquasecurity/trivy/pkg/utils/fsutils"
	sloglogrus "github.com/samber/slog-logrus/v2"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/config"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/scanner"
	"github.com/openclarity/vmclarity/scanner/utils/image_helper"
	"github.com/openclarity/vmclarity/scanner/utils/sbom"
	utilsTrivy "github.com/openclarity/vmclarity/scanner/utils/trivy"
	utilsVul "github.com/openclarity/vmclarity/scanner/utils/vulnerability"
)

const ScannerName = "trivy"

func New(_ string, c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(*config.Config) // nolint:forcetypeassert

	logger = logger.Dup().WithField("scanner", ScannerName)

	// Set up the logger for trivy
	tlogger := trivyLog.New(sloglogrus.Option{Logger: logger.Logger}.NewLogrusHandler())
	trivyLog.SetDefault(tlogger)

	return &Scanner{
		logger:     logger,
		config:     config.CreateScannerTrivyConfigEx(conf.Scanner, conf.Registry),
		resultChan: resultChan,
		localImage: conf.LocalImageScan,
	}
}

type Scanner struct {
	logger     *log.Entry
	config     config.ScannerTrivyConfigEx
	resultChan chan job_manager.Result
	localImage bool
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

func (a *Scanner) createTrivyOptions(output string, userInput string) (trivyFlag.Options, error) {
	severities, err := getAllTrivySeverities()
	if err != nil {
		return trivyFlag.Options{}, fmt.Errorf("unable to get all trivy severities: %w", err)
	}

	cacheDir := trivyFsutils.CacheDir()
	if a.config.CacheDir != "" {
		cacheDir = a.config.CacheDir
	}

	dbOptions, err := utilsTrivy.GetTrivyDBOptions()
	if err != nil {
		return trivyFlag.Options{}, fmt.Errorf("unable to get db options: %w", err)
	}

	trivyOptions := trivyFlag.Options{
		GlobalOptions: trivyFlag.GlobalOptions{
			Timeout:  a.config.Timeout,
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
		VulnerabilityOptions: trivyFlag.VulnerabilityOptions{
			VulnType: trivyTypes.VulnTypes, // Scan all vuln types trivy supports
		},
		ImageOptions: trivyFlag.ImageOptions{
			ImageSources: types.AllImageSources,
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

// nolint:cyclop
func (a *Scanner) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	a.logger.Infof("Called %s scanner on source %v %v", ScannerName, sourceType, userInput)

	tempFile, err := os.CreateTemp(a.config.CacheDir, "trivy.scan.*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	go func(ctx context.Context) {
		defer os.Remove(tempFile.Name())

		var hash string
		var metadata map[string]string
		switch sourceType {
		case utils.IMAGE, utils.ROOTFS, utils.DIR, utils.FILE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		case utils.SBOM:
			var err error
			bom, err := sbom.NewCycloneDX(userInput)
			if err != nil {
				a.setError(fmt.Errorf("failed to create CycloneDX SBOM: %w", err))
				return
			}
			metadata = bom.GetMetadataFromSBOM()
			hash, err = bom.GetHashFromSBOM()
			if err != nil {
				a.setError(fmt.Errorf("failed to get original hash from SBOM: %w", err))
				return
			}
		default:
			a.logger.Infof("Skipping scan for unsupported source type: %s", sourceType)
			a.resultChan <- a.CreateResult(nil, hash, metadata)
			return
		}

		trivyOptions, err := a.createTrivyOptions(tempFile.Name(), userInput)
		if err != nil {
			a.setError(fmt.Errorf("unable to create trivy options: %w", err))
			return
		}

		// Configure Trivy image options according to the source type and user input.
		trivyOptions, cleanup, err := utilsTrivy.SetTrivyImageOptions(sourceType, userInput, trivyOptions)
		defer cleanup(a.logger)
		if err != nil {
			a.setError(fmt.Errorf("failed to configure trivy image options: %w", err))
			return
		}

		// Convert the source to the trivy source type
		trivySourceType, err := utilsTrivy.SourceToTrivySource(sourceType)
		if err != nil {
			a.setError(fmt.Errorf("failed to configure trivy: %w", err))
			return
		}

		// Ensure we're configured for private registry if required
		trivyOptions = utilsTrivy.SetTrivyRegistryConfigs(a.config.Registry, trivyOptions)

		err = artifact.Run(ctx, trivyOptions, trivySourceType)
		if err != nil {
			a.setError(fmt.Errorf("failed to scan for vulnerabilities: %w", err))
			return
		}

		file, err := os.ReadFile(tempFile.Name())
		if err != nil {
			a.setError(fmt.Errorf("failed to read scan results: %w", err))
			return
		}

		a.logger.Infof("Sending successful results")
		a.resultChan <- a.CreateResult(file, hash, metadata)
	}(ctx)

	return nil
}

func getCVSSesFromVul(vCvss trivyDBTypes.VendorCVSS) []scanner.CVSS {
	// TODO(sambetts) Work out how to include all the
	// CVSS's by vendor ID in our report, for now only
	// collect one of each version.
	cvsses := []scanner.CVSS{}
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

			cvsses = append(cvsses, scanner.CVSS{
				Version: utilsVul.GetCVSSV3VersionFromVector(cvss.V3Vector),
				Vector:  cvss.V3Vector,
				Metrics: scanner.CvssMetrics{
					BaseScore:           cvss.V3Score,
					ExploitabilityScore: &exploit,
					ImpactScore:         &impact,
				},
			})
			v3Collected = true
		}
		if cvss.V2Vector != "" && !v2Collected {
			exploit, impact := utilsVul.ExploitScoreAndImpactScoreFromV2Vector(cvss.V2Vector)

			cvsses = append(cvsses, scanner.CVSS{
				Version: "2.0",
				Vector:  cvss.V2Vector,
				Metrics: scanner.CvssMetrics{
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

// nolint:cyclop
func (a *Scanner) CreateResult(trivyJSON []byte, hash string, metadata map[string]string) *scanner.Results {
	result := &scanner.Results{
		Matches: nil, // empty results,
		ScannerInfo: scanner.Info{
			Name: ScannerName,
		},
	}

	if len(trivyJSON) == 0 {
		return result
	}

	var report trivyTypes.Report
	err := json.Unmarshal(trivyJSON, &report)
	if err != nil {
		a.logger.Error(err)
		result.Error = err
		return result
	}

	matches := []scanner.Match{}
	for _, result := range report.Results {
		for _, vul := range result.Vulnerabilities {
			typ, err := getTypeFromPurl(vul.PkgIdentifier.BOMRef)
			if err != nil {
				a.logger.Error(err)
				typ = ""
			}

			cvsses := getCVSSesFromVul(vul.CVSS)

			fix := scanner.Fix{}
			if vul.FixedVersion != "" {
				fix.Versions = []string{
					vul.FixedVersion,
				}
			}

			distro := scanner.Distro{}
			if report.Metadata.OS != nil {
				// Trivy calls the distro name (ubuntu, debian, alpine) the family
				distro.Name = string(report.Metadata.OS.Family)
				// Trivy calls the version (11, hardy heron, 22.04) the name
				distro.Version = report.Metadata.OS.Name
			}

			links := make([]string, 0, len(vul.Vulnerability.References))
			links = append(links, vul.Vulnerability.References...)
			kbVul := scanner.Vulnerability{
				ID:          vul.VulnerabilityID,
				Description: vul.Description,
				Links:       links,
				Distro:      distro,
				CVSS:        cvsses,
				Fix:         fix,
				Severity:    strings.ToUpper(vul.Severity),
				Package: scanner.Package{
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

			matches = append(matches, scanner.Match{
				Vulnerability: kbVul,
			})
		}
	}

	a.logger.Infof("Found %d vulnerabilities", len(matches))

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
			log.Warningf("Failed to get image hash from repo digests or image id: %v", err)
		}
	}

	source := scanner.Source{
		Name:     report.ArtifactName,
		Type:     string(report.ArtifactType),
		Hash:     hash,
		Metadata: metadata,
	}

	result.Matches = matches
	result.Source = source
	return result
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

func (a *Scanner) setError(err error) {
	a.logger.Error(err)
	a.resultChan <- &scanner.Results{
		Matches: nil, // empty results,
		ScannerInfo: scanner.Info{
			Name: ScannerName,
		},
		Error: err,
	}
}
