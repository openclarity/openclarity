// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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

package scanresultwatcher

import log "github.com/sirupsen/logrus"

type DeleteJobPolicyType string

const (
	DeleteJobPolicyAlways    DeleteJobPolicyType = "Always"
	DeleteJobPolicyNever     DeleteJobPolicyType = "Never"
	DeleteJobPolicyOnSuccess DeleteJobPolicyType = "OnSuccess"
)

func (p DeleteJobPolicyType) IsValid() bool {
	switch p {
	case DeleteJobPolicyAlways, DeleteJobPolicyNever, DeleteJobPolicyOnSuccess:
		return true
	default:
		return false
	}
}

func GetDeleteJobPolicyType(policyType string) DeleteJobPolicyType {
	deleteJobPolicy := DeleteJobPolicyType(policyType)
	if !deleteJobPolicy.IsValid() {
		log.Warnf("Invalid DeleteJobPolicy type %s. Falling back to default: %s", policyType, DeleteJobPolicyAlways)
		deleteJobPolicy = DeleteJobPolicyAlways
	}

	return deleteJobPolicy
}
