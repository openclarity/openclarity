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

type CVSS struct {

	// cvss v3 metrics
	CvssV3Metrics *CVSSV3Metrics `json:"cvssV3Metrics,omitempty"`

	// cvss v3 vector
	CvssV3Vector *CVSSV3Vector `json:"cvssV3Vector,omitempty"`
}

func (c *CVSS) GetBaseScore() float64 {
	if c == nil {
		return 0
	}

	return c.CvssV3Metrics.getBaseScore()
}

func (c *CVSS) GetCVSSSeverity() models.VulnerabilitySeverity {
	if c == nil {
		return ""
	}

	return c.CvssV3Metrics.getCVSSSeverity()
}

func (c *CVSS) ToCVSSBackendAPI() *models.CVSS {
	if c == nil {
		return nil
	}

	return &models.CVSS{
		CvssV3Metrics: c.CvssV3Metrics.toCVSSBackendAPI(),
		CvssV3Vector:  c.CvssV3Vector.toCVSSBackendAPI(),
	}
}

func cvssFromRuntimeScan(in *runtime_scan_models.CVSS) *CVSS {
	if in == nil {
		return nil
	}

	return &CVSS{
		CvssV3Metrics: cvssV3MetricsFromRuntimeScan(in.CvssV3Metrics),
		CvssV3Vector:  cvssV3VectorFromRuntimeScan(in.CvssV3Vector),
	}
}

func cvssFromBackendAPI(in *models.CVSS) *CVSS {
	if in == nil {
		return nil
	}

	return &CVSS{
		CvssV3Metrics: cvssV3MetricsFromBackendAPI(in.CvssV3Metrics),
		CvssV3Vector:  cvssV3VectorFromBackendAPI(in.CvssV3Vector),
	}
}
