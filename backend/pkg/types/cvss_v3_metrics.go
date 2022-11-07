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

package types

import (
	"github.com/openclarity/kubeclarity/api/server/models"
	runtime_scan_models "github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
)

type CVSSV3Metrics struct {
	// base score
	BaseScore float64 `json:"baseScore,omitempty"`

	// exploitability score
	ExploitabilityScore float64 `json:"exploitabilityScore,omitempty"`

	// impact score
	ImpactScore float64 `json:"impactScore,omitempty"`
}

func (m *CVSSV3Metrics) getBaseScore() float64 {
	if m == nil {
		return 0
	}

	return m.BaseScore
}

// nolint:gomnd
func (m *CVSSV3Metrics) getCVSSSeverity() models.VulnerabilitySeverity {
	if m == nil {
		return ""
	}
	/*
		https://nvd.nist.gov/vuln-metrics/cvss
		CVSS v3.0 Ratings
			Severity	Base Score Range
			None		0.0
			Low			0.1-3.9
			Medium		4.0-6.9
			High		7.0-8.9
			Critical	9.0-10.0
	*/
	if m.BaseScore < 0.1 {
		return ""
	}
	if m.BaseScore <= 3.9 {
		return models.VulnerabilitySeverityLOW
	}
	if m.BaseScore <= 6.9 {
		return models.VulnerabilitySeverityMEDIUM
	}
	if m.BaseScore <= 8.9 {
		return models.VulnerabilitySeverityHIGH
	}
	return models.VulnerabilitySeverityCRITICAL
}

func (m *CVSSV3Metrics) toCVSSBackendAPI() *models.CVSSV3Metrics {
	if m == nil {
		return nil
	}

	return &models.CVSSV3Metrics{
		BaseScore:           m.BaseScore,
		ExploitabilityScore: m.ExploitabilityScore,
		ImpactScore:         m.ImpactScore,
		Severity:            m.getCVSSSeverity(),
	}
}

func cvssV3MetricsFromRuntimeScan(metrics *runtime_scan_models.CVSSV3Metrics) *CVSSV3Metrics {
	return &CVSSV3Metrics{
		BaseScore:           metrics.BaseScore,
		ExploitabilityScore: metrics.ExploitabilityScore,
		ImpactScore:         metrics.ImpactScore,
	}
}

func cvssV3MetricsFromBackendAPI(metrics *models.CVSSV3Metrics) *CVSSV3Metrics {
	return &CVSSV3Metrics{
		BaseScore:           metrics.BaseScore,
		ExploitabilityScore: metrics.ExploitabilityScore,
		ImpactScore:         metrics.ImpactScore,
	}
}
