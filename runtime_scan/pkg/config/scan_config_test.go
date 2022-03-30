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

package config

import (
	"testing"

	"gotest.tools/assert"
)

func Test_getDeleteJobPolicyType(t *testing.T) {
	assert.Assert(t, getDeleteJobPolicyType("invalid") == DeleteJobPolicySuccessful)
	assert.Assert(t, getDeleteJobPolicyType(string(DeleteJobPolicySuccessful)) == DeleteJobPolicySuccessful)
	assert.Assert(t, getDeleteJobPolicyType(string(DeleteJobPolicyNever)) == DeleteJobPolicyNever)
	assert.Assert(t, getDeleteJobPolicyType(string(DeleteJobPolicyAll)) == DeleteJobPolicyAll)
}
