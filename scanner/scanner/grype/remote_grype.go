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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openclarity/vmclarity/scanner/utils"

	grype_models "github.com/anchore/grype/grype/presenter/models"
	transport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	grype_client "github.com/openclarity/grype-server/api/client/client"
	grype_client_operations "github.com/openclarity/grype-server/api/client/client/operations"
	grype_client_models "github.com/openclarity/grype-server/api/client/models"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/config"
	"github.com/openclarity/vmclarity/scanner/job_manager"
	"github.com/openclarity/vmclarity/scanner/scanner"
	sbom "github.com/openclarity/vmclarity/scanner/utils/sbom"
)

type RemoteScanner struct {
	logger     *log.Entry
	resultChan chan job_manager.Result
	client     *grype_client.GrypeServer
	timeout    time.Duration
}

func newRemoteScanner(conf *config.Config, logger *log.Entry, resultChan chan job_manager.Result) job_manager.Job {
	cfg := grype_client.DefaultTransportConfig().WithSchemes(conf.Scanner.GrypeConfig.GrypeServerSchemes).WithHost(conf.Scanner.GrypeConfig.GrypeServerAddress)

	return &RemoteScanner{
		logger:     logger.Dup().WithField("scanner", ScannerName).WithField("scanner-mode", "remote"),
		resultChan: resultChan,
		client:     grype_client.New(transport.New(cfg.Host, cfg.BasePath, cfg.Schemes), strfmt.Default),
		timeout:    conf.Scanner.GrypeConfig.GrypeServerTimeout,
	}
}

func (s *RemoteScanner) Run(ctx context.Context, sourceType utils.SourceType, userInput string) error {
	// remote-grype supports only SBOM as a source input since it sends the SBOM to a centralized grype server for scanning.
	if sourceType != utils.SBOM {
		s.logger.Infof("Ignoring non SBOM input. type=%v", sourceType)
		s.resultChan <- &scanner.Results{
			Matches: nil, // empty results
			ScannerInfo: scanner.Info{
				Name: ScannerName,
			},
		}
		return nil
	}

	go s.run(userInput)

	return nil
}

func (s *RemoteScanner) run(sbomInputFilePath string) {
	sbomBytes, err := os.ReadFile(sbomInputFilePath)
	if err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to read input file: %w", err), s.logger)
		return
	}

	doc, err := s.scanSbomWithGrypeServer(sbomBytes)
	if err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to scan sbom with grype server: %w", err), s.logger)
		return
	}

	bom, err := sbom.NewCycloneDX(sbomInputFilePath)
	if err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to create CycloneDX SBOM: %w", err), s.logger)
		return
	}

	userInput := bom.GetTargetNameFromSBOM()
	metadata := bom.GetMetadataFromSBOM()
	hash, err := bom.GetHashFromSBOM()
	if err != nil {
		ReportError(s.resultChan, fmt.Errorf("failed to get original hash from SBOM: %w", err), s.logger)
		return
	}

	s.logger.Infof("Sending successful results")
	s.resultChan <- CreateResults(*doc, userInput, ScannerName, hash, metadata)
}

func (s *RemoteScanner) scanSbomWithGrypeServer(sbom []byte) (*grype_models.Document, error) {
	params := grype_client_operations.NewPostScanSBOMParams().
		WithBody(&grype_client_models.SBOM{
			Sbom: sbom,
		}).WithTimeout(s.timeout)
	ok, err := s.client.Operations.PostScanSBOM(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send sbom for scan: %w", err)
	}
	doc := grype_models.Document{}

	err = json.Unmarshal(ok.Payload.Vulnerabilities, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vulnerabilities document: %w", err)
	}

	return &doc, nil
}
