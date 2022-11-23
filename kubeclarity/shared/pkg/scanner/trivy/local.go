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
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	trivyDBTypes "github.com/aquasecurity/trivy-db/pkg/types"
	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	trivyTypes "github.com/aquasecurity/trivy/pkg/types"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	utilsTrivy "github.com/openclarity/kubeclarity/shared/pkg/utils/trivy"
)

type LocalScanner struct {
	logger     *log.Entry
	config     config.LocalScannerTrivyConfigEx
	resultChan chan job_manager.Result
	localImage bool
}

func (a *LocalScanner) Run(sourceType utils.SourceType, userInput string) error {
	a.logger.Infof("Called %s scanner on source %v %v", ScannerName, sourceType, userInput)
	go func() {
		switch sourceType {
		case utils.IMAGE, utils.ROOTFS, utils.DIR, utils.FILE, utils.SBOM:
			// These are all supported for vuln scanning so continue
		default:
			a.logger.Infof("Skipping scan for unsupported source type: %s", sourceType)
			a.resultChan <- a.CreateResult(nil)
			return
		}

		var output bytes.Buffer

		// Get the Trivy CVE DB URL default value from the trivy
		// configuration, we may want to make this configurable in the
		// future.
		dbRepoDefaultValue, ok := trivyFlag.DBRepositoryFlag.Value.(string)
		if !ok {
			a.setError(fmt.Errorf("unable to get trivy DB repo config"))
			return
		}

		// Build a list of CVE severities for the trivy scanner to
		// report.  trivyDBTypes.SeverityNames contains all the
		// severities that trivy supports and we want them all in our
		// report at the moment.
		severities := []trivyDBTypes.Severity{}
		for _, name := range trivyDBTypes.SeverityNames {
			sev, err := trivyDBTypes.NewSeverity(strings.ToUpper(name))
			if err != nil {
				a.setError(fmt.Errorf("unable to get trivy severities: %w", err))
				return
			}
			severities = append(severities, sev)
		}

		trivyOptions := trivyFlag.Options{
			GlobalOptions: trivyFlag.GlobalOptions{
				Timeout: a.config.Timeout,
			},
			ScanOptions: trivyFlag.ScanOptions{
				Target: userInput,
				SecurityChecks: []string{
					trivyTypes.SecurityCheckVulnerability, // Enable just vuln scanning
				},
			},
			ReportOptions: trivyFlag.ReportOptions{
				Format:       "json",     // Trivy's own json format is the most complete for vuls
				ReportFormat: "all",      // Full report not just summary
				Output:       &output,    // Save the output to our local buffer instead of Stdout
				ListAllPkgs:  false,      // Only include packages with vulnerabilities
				Severities:   severities, // All the severities from the above
			},
			DBOptions: trivyFlag.DBOptions{
				DBRepository: dbRepoDefaultValue, // Use the default trivy source for the vuln DB
				NoProgress:   true,               // Disable the interactive progress bar
			},
			VulnerabilityOptions: trivyFlag.VulnerabilityOptions{
				VulnType: trivyTypes.VulnTypes, // Scan all vuln types trivy supports
			},
		}

		// Convert the kubeclarity source to the trivy source type
		trivySourceType, err := utilsTrivy.KubeclaritySourceToTrivySource(sourceType)
		if err != nil {
			a.setError(fmt.Errorf("failed to configure trivy: %w", err))
			return
		}

		// Ensure we're configured for private registry if required
		trivyOptions = utilsTrivy.SetTrivyRegistryConfigs(a.config.Registry, trivyOptions)

		err = artifact.Run(context.TODO(), trivyOptions, trivySourceType)
		if err != nil {
			a.setError(fmt.Errorf("failed to generate SBOM: %w", err))
			return
		}

		a.logger.Infof("Sending successful results")
		a.resultChan <- a.CreateResult(output.Bytes())
	}()

	return nil
}

func getCVSSesFromVul(vCvss trivyDBTypes.VendorCVSS) []scanner.CVSS {
	// TODO(sambetts) Work out how to include all the
	// CVSS's by vendor ID in our report, for now only
	// collect one of each version.
	cvsses := []scanner.CVSS{}
	v2Collected := false
	v3Collected := false
	for _, cvss := range vCvss {
		if cvss.V3Vector != "" && !v3Collected {
			cvsses = append(cvsses, scanner.CVSS{
				Version: "3.1",
				Vector:  cvss.V3Vector,
				Metrics: scanner.CvssMetrics{
					BaseScore: cvss.V3Score,
				},
			})
			v3Collected = true
		}
		if cvss.V2Vector != "" && !v2Collected {
			cvsses = append(cvsses, scanner.CVSS{
				Version: "2.0",
				Vector:  cvss.V2Vector,
				Metrics: scanner.CvssMetrics{
					BaseScore: cvss.V2Score,
				},
			})
			v2Collected = true
		}
	}
	return cvsses
}

func (a *LocalScanner) CreateResult(trivyJSON []byte) *scanner.Results {
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
			typ, err := getTypeFromPurl(vul.Ref)
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

			kbVul := scanner.Vulnerability{
				ID:          vul.VulnerabilityID,
				Description: vul.Description,
				Links: []string{
					vul.PrimaryURL,
				},
				Distro: scanner.Distro{
					Name:    report.Metadata.OS.Family,
					Version: report.Metadata.OS.Name,
				},
				CVSS:     cvsses,
				Fix:      fix,
				Severity: vul.Severity,
				Package: scanner.Package{
					Name:    vul.PkgName,
					Version: vul.InstalledVersion,
					PURL:    vul.Ref,
					Type:    typ,
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

	source := scanner.Source{
		Name: report.ArtifactName,
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
		return "", fmt.Errorf("type not found in purl")
	}

	return typ, nil
}

func (a *LocalScanner) setError(err error) {
	a.logger.Error(err)
	a.resultChan <- &scanner.Results{
		Matches: nil, // empty results,
		ScannerInfo: scanner.Info{
			Name: ScannerName,
		},
		Error: err,
	}
}
