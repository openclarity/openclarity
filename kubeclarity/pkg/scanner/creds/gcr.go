package creds

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

const (
	GcrSaSecretName      = "gcr-sa"
	gcrSaSecretFileName  = "sa.json"
	gcrVolumeName        = "gcr-sa"
	gcrVolumeMountPath   = "/etc/gcr"
	googleAppCredsEnvVar = "GOOGLE_APPLICATION_CREDENTIALS"
)

type GCR struct {
	credsCommon
}

// ensure type implement the requisite interface
var _ CredentialAdder = &GCR{}

func CreateGCR(clientset kubernetes.Interface, secretNamespace string) *GCR {
	return &GCR{
		credsCommon: credsCommon{
			clientset:       clientset,
			secretNamespace: secretNamespace,
		},
	}
}

func (g *GCR) ShouldAdd() bool {
	if g.isSecretExists == nil {
		found := isSecretExists(g.clientset, GcrSaSecretName, g.secretNamespace)
		g.isSecretExists = &found
	}

	return *g.isSecretExists
}

// Klar is using google SDK to pull the user name and password required to pull the image.
// We need to do the following:
// 1. Create a volume that holds the `gcrSaSecretFileName` data
// 2. Mount the volume into each container to a specific path (`gcrVolumeMountPath`/`gcrSaSecretFileName`)
// 3. Set `GOOGLE_APPLICATION_CREDENTIALS` to point to the mounted file
func (g *GCR) Add(job *batchv1.Job) {
	job.Namespace = g.secretNamespace
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: gcrVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: GcrSaSecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  gcrSaSecretFileName,
						Path: gcrSaSecretFileName,
					},
				},
			},
		},
	})
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      gcrVolumeName,
			ReadOnly:  true,
			MountPath: gcrVolumeMountPath,
		})
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  googleAppCredsEnvVar,
			Value: strings.Join([]string{gcrVolumeMountPath, gcrSaSecretFileName}, "/"),
		})
	}
}
