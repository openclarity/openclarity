package creds

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	awsAccessKeyId     = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	awsDefaultRegion   = "AWS_DEFAULT_REGION"
	ecrSaSecretName    = "ecr-sa"
)

var ecrEnvs = []corev1.EnvVar{
	{
		Name: awsAccessKeyId,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ecrSaSecretName,
				},
				Key: awsAccessKeyId,
			},
		},
	},
	{
		Name: awsSecretAccessKey,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ecrSaSecretName,
				},
				Key: awsSecretAccessKey,
			},
		},
	},
	{
		Name: awsDefaultRegion,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ecrSaSecretName,
				},
				Key: awsDefaultRegion,
			},
		},
	},
}

type ECR struct {
	clientset       kubernetes.Interface
	isSecretExists  *bool
	secretNamespace string
}

func CreateECR(clientset kubernetes.Interface, secretNamespace string) *ECR {
	return &ECR{
		clientset:       clientset,
		secretNamespace: secretNamespace,
	}
}

func (e *ECR) ShouldAdd() bool {
	if e.isSecretExists == nil {
		found := isSecretExists(e.clientset, ecrSaSecretName, e.secretNamespace)
		e.isSecretExists = &found
	}

	return *e.isSecretExists
}

// Klar is using google SDK to pull the user name ans password required to pull the image.
// We need to set the following env variables from the `ecrSaSecretName` secret:
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
