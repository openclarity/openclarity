package k8s

import (
	"context"
	"github.com/containers/image/docker/reference"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/credentialprovider"
	credprovsecrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"
	"strings"
)

const MaxK8sJobName = 63

func CreateClientset() (kubernetes.Interface, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func GetPodImagePullSecrets(clientset kubernetes.Interface, pod corev1.Pod) []*corev1.Secret {
	var secrets []*corev1.Secret
	for _, secretName := range pod.Spec.ImagePullSecrets {
		secret, err := clientset.CoreV1().Secrets(pod.Namespace).Get(context.TODO(), secretName.Name, metav1.GetOptions{})
		if err != nil {
			log.Warnf("Failed to get secret %s in namespace %s. %+v", secretName.Name, pod.Namespace, err)
			continue
		}
		secrets = append(secrets, secret)
	}

	return secrets
}

func GetMatchingSecretName(secrets []*corev1.Secret, imageName string) string {
	for _, secret := range secrets {
		slice := []corev1.Secret{*secret}
		var singleSecretKeyRing = credentialprovider.NewDockerKeyring()
		singleSecretKeyRing, err := credprovsecrets.MakeDockerKeyring(slice, singleSecretKeyRing)
		if err != nil {
			return ""
		}
		namedImageRef, err := reference.ParseNormalizedNamed(imageName)
		if err != nil {
			return ""
		}
		_, credentialsExist := singleSecretKeyRing.Lookup(namedImageRef.Name())
		if credentialsExist {
			return secret.Name
		}
	}

	return ""
}

// example: for "docker-pullable://gcr.io/development-infra-208909/kubei@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa"
// returns 6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
func ParseImageHash(imageID string) string {
	index := strings.LastIndex(imageID, ":")
	if index == -1 {
		return ""
	}

	return imageID[index+1:]
}
