//go:generate $GOPATH/bin/mockgen -destination=./mock_k8ContextServiceSecretInterface.go -package=common kubei/common K8ContextSecretServiceInterface

package common

import (
	"github.com/docker/distribution/reference"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
	credprovsecrets "k8s.io/kubernetes/pkg/credentialprovider/secrets"
)

func AppendStringIfMissing(list []string, candidate string) []string {
	for _, ele := range list {
		if ele == candidate {
			return list
		}
	}
	list = append(list, candidate)
	return list
}

func AppendContainerImageNameIfMissing(list []ContainerImageName, candidate ContainerImageName) []ContainerImageName {
	for _, ele := range list {
		if ele == candidate {
			return list
		}
	}
	list = append(list, candidate)
	return list
}

func ContainsString(list []string, imageName string) bool {
	for _, a := range list {
		if a == imageName {
			return true
		}
	}
	return false
}

func (kcs *K8ContextService) GetK8ContextFromContainer(orchestratorImageK8ExtendedContextMap ImageK8ExtendedContextMap, pod *corev1.Pod, imageNamespacesMap ImageNamespacesMap, namespacedImageSecretMap NamespacedImageSecretMap, containerImagesSet map[ContainerImageName]bool, totalContainers int) (ImageNamespacesMap, NamespacedImageSecretMap, map[ContainerImageName]bool, int) {
	if kcs.shouldIgnore(pod) {
		return imageNamespacesMap, namespacedImageSecretMap, containerImagesSet, totalContainers
	}

	containers := pod.Spec.Containers

	secrets := kcs.GetPodImagePullSecrets(*pod)

	log.Debugf("Getting all container images for pod: %s", pod.Name)
	for _, container := range containers {
		secretName := kcs.K8ContextSecretService.GetMatchingSecretName(secrets, container)
		containerImageName := ContainerImageName(container.Image)
		imageNamespacesMap[pod.Namespace] = AppendContainerImageNameIfMissing(imageNamespacesMap[pod.Namespace], containerImageName)

		k8ExtendedContext := &K8ExtendedContext{
			Namespace: pod.Namespace,
			Container: container.Name,
			Pod:       pod.Name,
			Secret:    secretName,
		}

		contexts := orchestratorImageK8ExtendedContextMap[containerImageName]
		contexts = append(contexts, k8ExtendedContext)
		orchestratorImageK8ExtendedContextMap[containerImageName] = contexts

		namespacedImageSecretMap[string(containerImageName)+"_"+pod.Namespace] = secretName
		containerImagesSet[containerImageName] = true
	}

	totalContainers += len(containers)
	return imageNamespacesMap, namespacedImageSecretMap, containerImagesSet, totalContainers
}

func (kcss *K8ContextSecretService) GetMatchingSecretName(secrets []corev1.Secret, container corev1.Container) string {
	for _, secret := range secrets {
		slice := []corev1.Secret{secret}
		var singleSecretKeyRing = credentialprovider.NewDockerKeyring()
		singleSecretKeyRing, err := credprovsecrets.MakeDockerKeyring(slice, singleSecretKeyRing)
		if err != nil {
			return ""
		}
		namedImageRef, err := reference.ParseNormalizedNamed(container.Image)
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

func (kcs *K8ContextService) GetPodImagePullSecrets(pod corev1.Pod) []corev1.Secret {
	var secrets []corev1.Secret
	for _, secretName := range pod.Spec.ImagePullSecrets {
		secret, err := kcs.ExecutionConfig.Clientset.CoreV1().Secrets(pod.Namespace).Get(secretName.Name, metav1.GetOptions{})
		if err != nil {
			log.Warnf("Failed to get secret %s in namespace %s. %+v", secretName.Name, pod.Namespace, err)
			continue
		}
		secrets = append(secrets, *secret)
	}
	return secrets
}

func (kcs *K8ContextService) shouldIgnore(pod *corev1.Pod) bool {
	if ContainsString(kcs.ExecutionConfig.IgnoreNamespaces, pod.Namespace) {
		log.Infof("Skipping scan of pod: %s from namespace: %s. Namespace is in IGNORE_NAMESPACES list", pod.Name, pod.Namespace)
		return true

	}
	if pod.Labels["kubeiShouldScan"] == "false" {
		log.Debugf("Skipping scan of pod: %s from namespace: %s. Pod has label kubeiShouldScan=false", pod.Name, pod.Namespace)
		return true
	}

	return false
}
