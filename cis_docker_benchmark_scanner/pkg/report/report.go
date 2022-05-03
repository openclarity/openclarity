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

package report

import (
	"context"
	"fmt"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	transport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/openclarity/kubeclarity/cis_docker_benchmark_scanner/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/client"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/client/operations"
	"github.com/openclarity/kubeclarity/runtime_scan/api/client/models"
)

type Reporter interface {
	ReportScanResults(assessmentMap dockle_types.AssessmentMap) error
	ReportScanError(scanError *models.ScanError) error
}

type ReporterImpl struct {
	client *client.KubeClarityRuntimeScanAPIs
	conf   *config.Config
	ctx    context.Context
}

func CreateReporter(ctx context.Context, conf *config.Config) Reporter {
	cfg := client.DefaultTransportConfig().WithHost(conf.ResultServiceAddress)

	return &ReporterImpl{
		client: client.New(transport.New(cfg.Host, cfg.BasePath, cfg.Schemes), strfmt.Default),
		conf:   conf,
		ctx:    ctx,
	}
}

func (r *ReporterImpl) ReportScanError(scanError *models.ScanError) error {
	_, err := r.client.Operations.PostScanScanUUIDCisDockerBenchmarkResults(operations.NewPostScanScanUUIDCisDockerBenchmarkResultsParams().
		WithScanUUID(strfmt.UUID(r.conf.ScanUUID)).
		WithBody(r.createFailedResults(scanError)).
		WithContext(r.ctx))
	if err != nil {
		return fmt.Errorf("failed to report scan error: %v", err)
	}

	return nil
}

func (r *ReporterImpl) ReportScanResults(assessmentMap dockle_types.AssessmentMap) error {
	_, err := r.client.Operations.PostScanScanUUIDCisDockerBenchmarkResults(operations.NewPostScanScanUUIDCisDockerBenchmarkResultsParams().
		WithScanUUID(strfmt.UUID(r.conf.ScanUUID)).
		WithBody(r.createSuccessfulResults(assessmentMap)).
		WithContext(r.ctx))
	if err != nil {
		return fmt.Errorf("failed to report scan results: %v", err)
	}

	return nil
}

func (r *ReporterImpl) createFailedResults(scanError *models.ScanError) *models.CISDockerBenchmarkScan {
	return &models.CISDockerBenchmarkScan{
		ImageID: r.conf.ImageIDToScan,
		CisDockerBenchmarkScanResult: &models.CISDockerBenchmarkScanResult{
			Error:  scanError,
			Status: models.ScanStatusFAILED,
		},
	}
}

func (r *ReporterImpl) createSuccessfulResults(assessmentMap dockle_types.AssessmentMap) *models.CISDockerBenchmarkScan {
	return &models.CISDockerBenchmarkScan{
		ImageID: r.conf.ImageIDToScan,
		CisDockerBenchmarkScanResult: &models.CISDockerBenchmarkScanResult{
			Result: convertAssessmentMapToResults(assessmentMap),
			Status: models.ScanStatusSUCCESS,
		},
	}
}

func convertAssessmentMapToResults(assessmentMap dockle_types.AssessmentMap) []*models.CISDockerBenchmarkCodeInfo {
	ret := make([]*models.CISDockerBenchmarkCodeInfo, 0, len(assessmentMap))
	for _, info := range assessmentMap {
		ret = append(ret, convertCodeInfo(info))
	}
	return ret
}

func convertCodeInfo(codeInfo dockle_types.CodeInfo) *models.CISDockerBenchmarkCodeInfo {
	return &models.CISDockerBenchmarkCodeInfo{
		Assessments: convertAssessments(codeInfo.Assessments),
		Code:        codeInfo.Code,
		Level:       int64(codeInfo.Level),
	}
}

func convertAssessments(assessments []*dockle_types.Assessment) []*models.CISDockerBenchmarkAssessment {
	ret := make([]*models.CISDockerBenchmarkAssessment, 0, len(assessments))
	for _, assessment := range assessments {
		ret = append(ret, convertAssessment(assessment))
	}

	return ret
}

func convertAssessment(assessment *dockle_types.Assessment) *models.CISDockerBenchmarkAssessment {
	return &models.CISDockerBenchmarkAssessment{
		Code:     assessment.Code,
		Desc:     assessment.Desc,
		Filename: assessment.Filename,
		Level:    int64(assessment.Level),
	}
}
