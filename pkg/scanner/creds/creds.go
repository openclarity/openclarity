package creds

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
)

type CredentialAdder interface {
	// Returns true if credentials should be added to a scanner job
	ShouldAdd() bool
	// Adds credentials to a scanner job
	Add(job *batchv1.Job)
}

type credsCommon struct {
	clientset       kubernetes.Interface
	isSecretExists  *bool
	secretNamespace string
}
