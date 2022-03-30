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
	"wwwin-github.cisco.com/eti/scan-gazr/api/server/models"
	runtime_scan_models "wwwin-github.cisco.com/eti/scan-gazr/runtime_scan/api/server/models"
)

type CVSSV3Vector struct {

	// attack complexity
	AttackComplexity AttackComplexity `json:"attackComplexity,omitempty"`

	// attack vector
	AttackVector AttackVector `json:"attackVector,omitempty"`

	// availability
	Availability Availability `json:"availability,omitempty"`

	// confidentiality
	Confidentiality Confidentiality `json:"confidentiality,omitempty"`

	// integrity
	Integrity Integrity `json:"integrity,omitempty"`

	// privileges required
	PrivilegesRequired PrivilegesRequired `json:"privilegesRequired,omitempty"`

	// scope
	Scope Scope `json:"scope,omitempty"`

	// user interaction
	UserInteraction UserInteraction `json:"userInteraction,omitempty"`

	// vector
	Vector string `json:"vector,omitempty"`
}

func (v *CVSSV3Vector) toCVSSBackendAPI() *models.CVSSV3Vector {
	if v == nil {
		return nil
	}

	return &models.CVSSV3Vector{
		AttackComplexity:   models.AttackComplexity(v.AttackComplexity),
		AttackVector:       models.AttackVector(v.AttackVector),
		Availability:       models.Availability(v.Availability),
		Confidentiality:    models.Confidentiality(v.Confidentiality),
		Integrity:          models.Integrity(v.Integrity),
		PrivilegesRequired: models.PrivilegesRequired(v.PrivilegesRequired),
		Scope:              models.Scope(v.Scope),
		UserInteraction:    models.UserInteraction(v.UserInteraction),
		Vector:             v.Vector,
	}
}

func cvssV3VectorFromRuntimeScan(vector *runtime_scan_models.CVSSV3Vector) *CVSSV3Vector {
	return &CVSSV3Vector{
		AttackComplexity:   AttackComplexity(vector.AttackComplexity),
		AttackVector:       AttackVector(vector.AttackVector),
		Availability:       Availability(vector.Availability),
		Confidentiality:    Confidentiality(vector.Confidentiality),
		Integrity:          Integrity(vector.Integrity),
		PrivilegesRequired: PrivilegesRequired(vector.PrivilegesRequired),
		Scope:              Scope(vector.Scope),
		UserInteraction:    UserInteraction(vector.UserInteraction),
		Vector:             vector.Vector,
	}
}

func cvssV3VectorFromBackendAPI(vector *models.CVSSV3Vector) *CVSSV3Vector {
	return &CVSSV3Vector{
		AttackComplexity:   AttackComplexity(vector.AttackComplexity),
		AttackVector:       AttackVector(vector.AttackVector),
		Availability:       Availability(vector.Availability),
		Confidentiality:    Confidentiality(vector.Confidentiality),
		Integrity:          Integrity(vector.Integrity),
		PrivilegesRequired: PrivilegesRequired(vector.PrivilegesRequired),
		Scope:              Scope(vector.Scope),
		UserInteraction:    UserInteraction(vector.UserInteraction),
		Vector:             vector.Vector,
	}
}
