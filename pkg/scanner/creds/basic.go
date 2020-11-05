package creds

import (
	klar "github.com/Portshift/klar/docker/token/secret"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	BasicRegCredSecretName = "basic-regcred"
)

type BasicRegCred struct {
	credsCommon
}

// ensure type implement the requisite interface
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
			Name: klar.ImagePullSecretEnvVar, ValueFrom: &corev1.EnvVarSource{
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
