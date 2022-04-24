package types

import (
	"strings"

	"github.com/cisco-open/kubei/api/server/models"
	runtime_scan_models "github.com/cisco-open/kubei/runtime_scan/api/server/models"
)

type CISDockerBenchmarkResult struct {
	Code         string `json:"code,omitempty"`
	Level        int64  `json:"level,omitempty"`
	Descriptions string `json:"descriptions"`
}

func CISDockerBenchmarkResultsFromBackendAPI(in []*models.CISDockerBenchmarkCodeInfo) []*CISDockerBenchmarkResult {
	ret := make([]*CISDockerBenchmarkResult, len(in))
	for i, scan := range in {
		ret[i] = cisDockerBenchmarkResultFromBackendAPI(scan)
	}
	return ret
}

func cisDockerBenchmarkResultFromBackendAPI(in *models.CISDockerBenchmarkCodeInfo) *CISDockerBenchmarkResult {
	return &CISDockerBenchmarkResult{
		Code:         in.Code,
		Level:        in.Level,
		Descriptions: getDescriptionsFromBackendAPI(in.Assessments),
	}
}

func getDescriptionsFromBackendAPI(assessments []*models.CISDockerBenchmarkAssessment) string {
	description := make([]string, len(assessments))
	for i := range assessments {
		description[i] = assessments[i].Desc
	}

	return strings.Join(description, ", ")
}

func CISDockerBenchmarkResultsFromFromRuntimeScan(in []*runtime_scan_models.CISDockerBenchmarkCodeInfo) []*CISDockerBenchmarkResult {
	ret := make([]*CISDockerBenchmarkResult, len(in))
	for i, scan := range in {
		ret[i] = cisDockerBenchmarkResultFromFromRuntimeScan(scan)
	}
	return ret
}

func cisDockerBenchmarkResultFromFromRuntimeScan(in *runtime_scan_models.CISDockerBenchmarkCodeInfo) *CISDockerBenchmarkResult {
	return &CISDockerBenchmarkResult{
		Code:         in.Code,
		Level:        in.Level,
		Descriptions: getDescriptionsFromRuntimeScan(in.Assessments),
	}
}

func getDescriptionsFromRuntimeScan(assessments []*runtime_scan_models.CISDockerBenchmarkAssessment) string {
	description := make([]string, len(assessments))
	for i := range assessments {
		description[i] = assessments[i].Desc
	}

	return strings.Join(description, ", ")
}
