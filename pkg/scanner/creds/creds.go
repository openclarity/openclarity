package creds

import batchv1 "k8s.io/api/batch/v1"

type CredentialAdder interface {
	// Returns true if credentials should be added to a scanner job
	ShouldAdd() bool
	// Adds credentials to a scanner job
	Add(job *batchv1.Job)
}
