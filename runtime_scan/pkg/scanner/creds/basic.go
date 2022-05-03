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

	shared "github.com/openclarity/kubeclarity/shared/pkg/config"
)

const (
	BasicRegCredSecretName = "basic-regcred" // nolint: gosec
)

type BasicRegCred struct {
	credsCommon
}

// ensure type implement the requisite interface.
var _ CredentialAdder = &BasicRegCred{}

func CreateBasicRegCred(clientset kubernetes.Interface, secretNamespace string) *BasicRegCred {
	return &BasicRegCred{
		credsCommon: credsCommon{
			clientset:       clientset,
			secretNamespace: secretNamespace,
		},
	}
}

func (u *BasicRegCred) ShouldAdd() bool {
	if u.isSecretExists == nil {
		found := isSecretExists(u.clientset, BasicRegCredSecretName, u.secretNamespace)
		u.isSecretExists = &found
	}

	return *u.isSecretExists
}

func (u *BasicRegCred) Add(job *batchv1.Job) {
	job.Namespace = u.secretNamespace
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{
			Name: shared.ImagePullSecret, ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: BasicRegCredSecretName,
					},
					Key: corev1.DockerConfigJsonKey,
				},
			},
		})
	}
}
