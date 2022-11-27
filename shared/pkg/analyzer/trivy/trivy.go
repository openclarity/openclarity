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

	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"

	"github.com/openclarity/kubeclarity/shared/pkg/analyzer"
	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/formatter"
	"github.com/openclarity/kubeclarity/shared/pkg/job_factory"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
)

const AnalyzerName = "trivy"

type Analyzer struct {
	name       string
	logger     *log.Entry
	config     config.AnalyzerTrivyConfigEx
	resultChan chan job_manager.Result
	localImage bool
}

func init() {
	job_factory.RegisterCreateJobFunc(AnalyzerName, New)
}

func New(c job_manager.IsConfig, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	conf := c.(config.Config) // nolint:forcetypeassert
	return &Analyzer{
		name:       AnalyzerName,
		logger:     logger.Dup().WithField("analyzer", AnalyzerName),
		config:     config.CreateAnalyzerTrivyConfigEx(&conf.Analyzer, &conf.Registry),
		resultChan: resultChan,
		localImage: conf.LocalImageScan,
	}
}

func (a *Analyzer) Run(sourceType utils.SourceType, userInput string) error {
	a.logger.Infof("Called %s analyzer on source %v %v", a.name, sourceType, userInput)
	go func() {
		res := &analyzer.Results{}

		var output bytes.Buffer
		trivyOptions := trivyFlag.Options{
			GlobalOptions: trivyFlag.GlobalOptions{
				Timeout: a.config.Timeout,
			},
			ScanOptions: trivyFlag.ScanOptions{
				Target:         userInput,
				SecurityChecks: nil, // Disable all security checks for SBOM only scan
			},
			ReportOptions: trivyFlag.ReportOptions{
				Format:       "cyclonedx", // Cyconedx format for SBOM so that we don't need to convert
				ReportFormat: "all",       // Full report not just summary
				Output:       &output,     // Save the output to our local buffer instead of Stdout
				ListAllPkgs:  true,        // By default Trivy only includes packages with vulnerabilities, for full SBOM set true.
			},
		}

		var trivySourceType artifact.TargetKind
		switch sourceType {
		case utils.IMAGE:
			trivySourceType = artifact.TargetContainerImage
		case utils.ROOTFS:
			trivySourceType = artifact.TargetRootfs
		case utils.DIR, utils.FILE:
			trivySourceType = artifact.TargetFilesystem
		case utils.SBOM:
			fallthrough
		default:
			a.logger.Infof("Skipping analyze unsupported source type: %s", sourceType)
			a.resultChan <- res
			return
		}

		err := artifact.Run(context.TODO(), trivyOptions, trivySourceType)
		if err != nil {
			a.setError(res, fmt.Errorf("failed to generate SBOM: %w", err))
			return
		}

		frm := formatter.New(formatter.CycloneDXJSONFormat, output.Bytes())
		if err := frm.Decode(formatter.CycloneDXJSONFormat); err != nil {
			a.setError(res, fmt.Errorf("failed to decode trivy results in formatter: %w", err))
			return
		}

		if err := frm.Encode(a.config.OutputFormat); err != nil {
			a.setError(res, fmt.Errorf("failed to encode trivy results: %w", err))
			return
		}

		res = analyzer.CreateResults(frm.GetSBOMBytes(), a.name, userInput, sourceType)
		a.logger.Infof("Sending successful results")
		a.resultChan <- res
	}()

	return nil
}

func (a *Analyzer) setError(res *analyzer.Results, err error) {
	res.Error = err
	a.logger.Error(res.Error)
	a.resultChan <- res
}
