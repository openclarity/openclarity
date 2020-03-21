package common

import (
	"github.com/Portshift-Admin/klar/clair"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type K8ContextServiceInterface interface {
	GetK8ContextFromContainer(orchestratorImageK8ExtendedContextMap ImageK8ExtendedContextMap, pod *corev1.Pod, imageNamespacesMap ImageNamespacesMap, namespacedImageSecretMap NamespacedImageSecretMap, containerImagesSet map[ContainerImageName]bool, totalContainers int) (ImageNamespacesMap, NamespacedImageSecretMap, map[ContainerImageName]bool, int)
	GetPodImagePullSecrets(pod corev1.Pod) []corev1.Secret
}

type K8ContextService struct {
	ExecutionConfig        *ExecutionConfiguration
	K8ContextSecretService K8ContextSecretServiceInterface
}

type K8ContextSecretServiceInterface interface {
	GetMatchingSecretName(secrets []corev1.Secret, container corev1.Container) string
}

type K8ContextSecretService struct{}

type ContainerImageName string

type ImageNamespacesMap map[string][]ContainerImageName

type ImageK8ExtendedContextMap map[ContainerImageName][]*K8ExtendedContext

type NamespacedImageSecretMap map[string]string

type ViewData struct {
	Vulnerabilities      []*ExtendedContextualVulnerability `json:"vulnerabilities,omitempty"`
	Total                int                                `json:"total"`
	TotalDefcon1         int                                `json:"totalDefcon1"`
	TotalCritical        int                                `json:"totalCritical"`
	TotalHigh            int                                `json:"totalHigh"`
	ShowGoMsg            bool                               `json:"showGoMsg"`
	ShowGoWarning        bool                               `json:"ShowGoWarning"`
	LastScannedNamespace string                             `json:"lastScannedNamespace"`
}

type ExecutionConfiguration struct {
	Clientset        *kubernetes.Clientset `json:"clientset"`
	Parallelism      int                   `json:"parallelism"`
	KubeiNamespace   string                `json:"kubeiNamespace"`
	TargetNamespace  string                `json:"targetNamespace"`
	ClairOutput      string                `json:"clairOutput"`
	WhitelistFile    string                `json:"whitelistFile"`
	IgnoreNamespaces []string              `json:"IgnoreNamespaces"`
	KlarTrace        bool                  `json:"klarTrace"`
}

type ContextualVulnerability struct {
	Vulnerability *clair.Vulnerability `json:"vulnerabilities"`
	Image         string               `json:"image"`
}

type ExtendedContextualVulnerability struct {
	Vulnerability *clair.Vulnerability `json:"vulnerability"`
	Pod           string               `json:"pod"`
	Container     string               `json:"container"`
	Image         string               `json:"image"`
	Namespace     string               `json:"namespace"`
}

type K8ExtendedContext struct {
	Namespace string `json:"namespace"`
	Container string `json:"container"`
	Pod       string `json:"pod"`
	Secret    string `json:"secret"`
}
