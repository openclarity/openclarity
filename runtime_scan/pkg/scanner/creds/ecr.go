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

package creds

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// nolint: gosec
const (
	AwsAccessKeyID     = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	AwsDefaultRegion   = "AWS_DEFAULT_REGION"
	EcrSaSecretName    = "ecr-sa"
)

var ecrEnvs = []corev1.EnvVar{
	{
		Name: AwsAccessKeyID,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: EcrSaSecretName,
				},
				Key: AwsAccessKeyID,
			},
		},
	},
	{
		Name: AwsSecretAccessKey,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: EcrSaSecretName,
				},
				Key: AwsSecretAccessKey,
			},
		},
	},
	{
		Name: AwsDefaultRegion,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: EcrSaSecretName,
				},
				Key: AwsDefaultRegion,
			},
		},
	},
}

type ECR struct {
	credsCommon
}

// ensure type implement the requisite interface.
var _ CredentialAdder = &ECR{}

func CreateECR(clientset kubernetes.Interface, secretNamespace string) *ECR {
	return &ECR{
		credsCommon: credsCommon{
			clientset:       clientset,
			secretNamespace: secretNamespace,
		},
	}
}

func (e *ECR) ShouldAdd() bool {
	if e.isSecretExists == nil {
		found := isSecretExists(e.clientset, EcrSaSecretName, e.secretNamespace)
		e.isSecretExists = &found
	}

	return *e.isSecretExists
}

// Add The scanner is using AWS SDK to pull the username and the password required to pull the image.
// We need to set the following env variables from the `EcrSaSecretName` secret:
// 1. AWS_ACCESS_KEY_ID
// 2. AWS_SECRET_ACCESS_KEY
// 3. AWS_DEFAULT_REGION.
func (e *ECR) Add(job *batchv1.Job) {
	job.Namespace = e.secretNamespace
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, ecrEnvs...)
	}
}

func (e *ECR) GetNamespace() string {
	return e.secretNamespace
}
