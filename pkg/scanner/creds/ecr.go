package creds

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	AwsAccessKeyId     = "AWS_ACCESS_KEY_ID"
	AwsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	AwsDefaultRegion   = "AWS_DEFAULT_REGION"
	EcrSaSecretName    = "ecr-sa"
)

var ecrEnvs = []corev1.EnvVar{
	{
		Name: AwsAccessKeyId,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: EcrSaSecretName,
				},
				Key: AwsAccessKeyId,
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

// ensure type implement the requisite interface
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

// Klar is using google SDK to pull the user name ans password required to pull the image.
// We need to set the following env variables from the `EcrSaSecretName` secret:
// 1. AWS_ACCESS_KEY_ID
// 2. AWS_SECRET_ACCESS_KEY
// 3. AWS_DEFAULT_REGION
func (e *ECR) Add(job *batchv1.Job) {
	job.Namespace = e.secretNamespace
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, ecrEnvs...)
	}
}
