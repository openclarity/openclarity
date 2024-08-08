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

	"github.com/openclarity/vmclarity/core/log"
	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families"
	"github.com/openclarity/vmclarity/scanner/families/vulnerabilities/types"

	grype_models "github.com/anchore/grype/grype/presenter/models"
	transport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	grype_client "github.com/openclarity/grype-server/api/client/client"
	grype_client_operations "github.com/openclarity/grype-server/api/client/client/operations"
	grype_client_models "github.com/openclarity/grype-server/api/client/models"

	sbom "github.com/openclarity/vmclarity/scanner/utils/sbom"
)

type RemoteScanner struct {
	client  *grype_client.GrypeServer
	timeout time.Duration
}

func newRemoteScanner(config types.Config) families.Scanner[*types.ScannerResult] {
	grypeConfig := config.ScannersConfig.Grype.Remote
	cfg := grype_client.DefaultTransportConfig().
		WithSchemes(grypeConfig.GrypeServerSchemes).
		WithHost(grypeConfig.GrypeServerAddress)

	return &RemoteScanner{
		client:  grype_client.New(transport.New(cfg.Host, cfg.BasePath, cfg.Schemes), strfmt.Default),
		timeout: grypeConfig.GrypeServerTimeout,
	}
}

func (s *RemoteScanner) Scan(ctx context.Context, sourceType common.InputType, userInput string) (*types.ScannerResult, error) {
	// remote-grype supports only SBOM as a source input since it sends the SBOM to a centralized grype server for scanning.
	if !sourceType.IsOneOf(common.SBOM) {
		return nil, fmt.Errorf("unsupported input type=%v", sourceType)
	}

	logger := log.GetLoggerFromContextOrDefault(ctx).WithField("grype-type", "remote")

	sbomBytes, err := os.ReadFile(userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	doc, err := s.scanSbomWithGrypeServer(sbomBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to scan sbom with grype server: %w", err)
	}

	bom, err := sbom.NewCycloneDX(userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to create CycloneDX SBOM: %w", err)
	}

	targetName := bom.GetTargetNameFromSBOM()
	metadata := bom.GetMetadataFromSBOM()
	hash, err := bom.GetHashFromSBOM()
	if err != nil {
		return nil, fmt.Errorf("failed to get original hash from SBOM: %w", err)
	}

	logger.Infof("Sending successful results")
	result := createResults(*doc, targetName, ScannerName, hash, metadata)

	return result, nil
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
