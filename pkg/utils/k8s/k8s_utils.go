package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	credprovsecrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"
)

const MaxK8sJobName = 63

func CreateClientset(kubeconfigPath string) (kubernetes.Interface, error) {
	// Create Kubernetes go-client clientset
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %v", err)
	}

	// Create a rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a rest client: %v", err)
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
		dockerKeyring, err := credprovsecrets.MakeDockerKeyring(slice, nil)
		if err != nil || dockerKeyring == nil {
			return ""
		}
		namedImageRef, err := reference.ParseNormalizedNamed(imageName)
		if err != nil {
			return ""
		}
		_, credentialsExist := dockerKeyring.Lookup(namedImageRef.Name())
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
