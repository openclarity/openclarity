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
	"encoding/json"
	"fmt"
	"os"
	"time"

	grype_client "github.com/Portshift/grype-server/api/client/client"
	grype_client_operations "github.com/Portshift/grype-server/api/client/client/operations"
	grype_client_models "github.com/Portshift/grype-server/api/client/models"
	grype_models "github.com/anchore/grype/grype/presenter/models"
	transport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/scanner/types"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	utilsSBOM "github.com/openclarity/kubeclarity/shared/pkg/utils/sbom"
)

type RemoteScanner struct {
	logger  *log.Entry
	client  *grype_client.GrypeServer
	timeout time.Duration
}

func newRemoteScanner(conf *config.Config, logger *log.Entry) (job_manager.Job[utils.SourceInput, types.Results], error) {
	cfg := grype_client.DefaultTransportConfig().WithHost(conf.Scanner.GrypeConfig.GrypeServerAddress)

	return &RemoteScanner{
		logger:  logger.Dup().WithField("scanner", ScannerName).WithField("scanner-mode", "remote"),
		client:  grype_client.New(transport.New(cfg.Host, cfg.BasePath, cfg.Schemes), strfmt.Default),
		timeout: conf.Scanner.GrypeConfig.GrypeServerTimeout,
	}, nil
}

func (s *RemoteScanner) Run(sourceInput utils.SourceInput) (types.Results, error) {
	res := types.Results{
		Matches: nil, // empty results
		ScannerInfo: types.Info{
			Name: ScannerName,
		},
	}

	// remote-grype supports only SBOM as a source input since it sends the SBOM to a centralized grype server for scanning.
	if sourceInput.Type != utils.SBOM {
		s.logger.Infof("Ignoring non SBOM input. type=%v", sourceInput.Type)
		return res, nil
	}

	sbom, err := os.ReadFile(sourceInput.Source)
	if err != nil {
		return res, fmt.Errorf("failed to read input file: %w", err)
	}

	doc, err := s.scanSbomWithGrypeServer(sbom)
	if err != nil {
		return res, fmt.Errorf("failed to scan sbom with grype server: %w", err)
	}

	userInput, hash, err := utilsSBOM.GetTargetNameAndHashFromSBOM(sourceInput.Source)
	if err != nil {
		return res, fmt.Errorf("failed to get original source and hash from SBOM: %w", err)
	}

	s.logger.Infof("Sending successful results")
	return CreateResults(*doc, userInput, ScannerName, hash), nil
}

func (s *RemoteScanner) scanSbomWithGrypeServer(sbom []byte) (*grype_models.Document, error) {
	params := grype_client_operations.NewPostScanSBOMParams().
		WithBody(&grype_client_models.SBOM{
			Sbom: sbom,
		}).WithTimeout(s.timeout)
	ok, err := s.client.Operations.PostScanSBOM(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send sbom for scan: %v", err)
	}
	doc := grype_models.Document{}

	err = json.Unmarshal(ok.Payload.Vulnerabilities, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vulnerabilities document: %v", err)
	}

	return &doc, nil
}
