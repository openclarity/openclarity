package config

import (
	"gotest.tools/assert"
	"testing"
)

func Test_getDeleteJobPolicyType(t *testing.T) {
	assert.Assert(t, getDeleteJobPolicyType("invalid") == DeleteJobPolicySuccessful)
	assert.Assert(t, getDeleteJobPolicyType(string(DeleteJobPolicySuccessful)) == DeleteJobPolicySuccessful)
	assert.Assert(t, getDeleteJobPolicyType(string(DeleteJobPolicyNever)) == DeleteJobPolicyNever)
	assert.Assert(t, getDeleteJobPolicyType(string(DeleteJobPolicyAll)) == DeleteJobPolicyAll)
}