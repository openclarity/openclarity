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
