package scanner

import (
	"context"
	"fmt"
	"github.com/Portshift/klar/clair"
	"github.com/Portshift/klar/forwarding"
	"github.com/Portshift/kubei/pkg/config"
	"github.com/Portshift/kubei/pkg/scanner/creds"
	"github.com/Portshift/kubei/pkg/types"
	k8s_utils "github.com/Portshift/kubei/pkg/utils/k8s"
	slice_utils "github.com/Portshift/kubei/pkg/utils/slice"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sync"
	"sync/atomic"
)

type Status string

const (
	Idle            Status = "Idle"
	ScanInit        Status = "ScanInit"
	ScanInitFailure Status = "ScanInitFailure"
	Scanning        Status = "Scanning"
)

type Scanner struct {
	imageToScanData  map[string]*scanData
	progress         types.ScanProgress
	status           Status
	config           *config.Config
	scanConfig       *config.ScanConfig
	killSignal       chan bool
	clientset        kubernetes.Interface
	logFields        log.Fields
	credentialAdders []creds.CredentialAdder
	sync.Mutex
}

func CreateScanner(config *config.Config, clientset kubernetes.Interface) *Scanner {
	s := &Scanner{
		progress:         types.ScanProgress{},
		status:           Idle,
		config:           config,
		killSignal:       make(chan bool),
		clientset:        clientset,
		logFields:        log.Fields{"scanner id": uuid.NewV4().String()},
		credentialAdders: []creds.CredentialAdder{
			creds.CreateECR(clientset, config.CredsSecretNamespace),
			creds.CreateGCR(clientset, config.CredsSecretNamespace),
		},
		Mutex:            sync.Mutex{},
	}

	return s
}

type imagePodContext struct {
	containerName   string
	podName         string
	namespace       string
	imagePullSecret string
	imageHash       string
	podUid          string
}

type scanData struct {
	imageName  string
	contexts   []*imagePodContext // All the pods that contain this image
	scanUUID   string
	result     []*clair.Vulnerability
	resultChan chan bool
	success    bool
	completed  bool
	timeout    bool
	scanErrMsg string
}

const (
	ignorePodScanLabelKey   = "kubeiShouldScan"
	ignorePodScanLabelValue = "false"
)

func (s *Scanner) shouldIgnorePod(pod *corev1.Pod) bool {
	if slice_utils.ContainsString(s.scanConfig.IgnoredNamespaces, pod.Namespace) {
		log.WithFields(s.logFields).Infof("Skipping pod scan, namespace is in the ignored namespaces list.  pod=%v ,namespace=%s", pod.Name, pod.Namespace)
		return true
	}
	if pod.Labels != nil && pod.Labels[ignorePodScanLabelKey] == ignorePodScanLabelValue {
		log.WithFields(s.logFields).Infof("Skipping pod scan, pod has an ignore label. pod=%v ,namespace=%s", pod.Name, pod.Namespace)
		return true
	}

	return false
}

func (s *Scanner) initScan() error {
	s.status = ScanInit

	// Get all target pods
	podList, err := s.clientset.CoreV1().Pods(s.scanConfig.TargetNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods. namespace=%s: %v", s.scanConfig.TargetNamespace, err)
	}

	imageToScanData := make(map[string]*scanData)

	// Populate the image to scanData map from all target pods
	for _, pod := range podList.Items {
		if s.shouldIgnorePod(&pod) {
			continue
		}
		secrets := k8s_utils.GetPodImagePullSecrets(s.clientset, pod)

		// Due to scenarios where image name in the `pod.Status.ContainerStatuses` is different
		// from image name in the `pod.Spec.Containers` we will take only image id from `pod.Status.ContainerStatuses`.
		containerNameToImageId := make(map[string]string)
		for _, container := range pod.Status.ContainerStatuses {
			containerNameToImageId[container.Name] = container.ImageID
		}

		for _, container := range pod.Spec.Containers {
			// Create pod context
			podContext := &imagePodContext{
				containerName:   container.Name,
				podName:         pod.GetName(),
				podUid:          string(pod.GetUID()),
				namespace:       pod.GetNamespace(),
				imagePullSecret: k8s_utils.GetMatchingSecretName(secrets, container.Image),
				imageHash:       getImageHash(containerNameToImageId, container),
			}
			if data, ok := imageToScanData[container.Image]; !ok {
				// Image added for the first time, create scan data and append pod context
				imageToScanData[container.Image] = &scanData{
					imageName:  container.Image,
					contexts:   []*imagePodContext{podContext},
					scanUUID:   uuid.NewV4().String(),
					resultChan: make(chan bool),
				}
			} else {
				// Image already exist in map, just append the pod context
				data.contexts = append(data.contexts, podContext)
			}
		}
	}

	s.imageToScanData = imageToScanData
	s.progress = types.ScanProgress{
		ImagesToScan:          uint32(len(imageToScanData)),
		ImagesStartedToScan:   0,
		ImagesCompletedToScan: 0,
	}

	log.WithFields(s.logFields).Infof("Total %d unique images to scan", s.progress.ImagesToScan)

	return nil
}

func getImageHash(containerNameToImageId map[string]string, container corev1.Container) string {
	imageID, ok := containerNameToImageId[container.Name]
	if !ok {
		log.Warnf("Image id is missing. container=%v ,image=%v", container.Name, container.Image)
		return ""
	}

	imageHash := k8s_utils.ParseImageHash(imageID)
	if imageHash == "" {
		log.Warnf("Failed to parse image hash. container=%v ,image=%v, image id=%v", container.Name, container.Image, imageID)
		return ""
	}

	return imageHash
}

func (s *Scanner) Scan(scanConfig *config.ScanConfig) error {
	s.Lock()
	defer s.Unlock()

	s.scanConfig = scanConfig
	log.WithFields(s.logFields).Infof("Start scanning...")
	err := s.initScan()
	if err != nil {
		s.status = ScanInitFailure
		return fmt.Errorf("failed to initiate scan: %v", err)
	}

	s.status = Scanning
	go s.jobBatchManagement()

	return nil
}

func (s *Scanner) ScanProgress() types.ScanProgress {
	return types.ScanProgress{
		ImagesToScan:          s.progress.ImagesToScan,
		ImagesStartedToScan:   atomic.LoadUint32(&s.progress.ImagesStartedToScan),
		ImagesCompletedToScan: atomic.LoadUint32(&s.progress.ImagesCompletedToScan),
	}
}

func (s *Scanner) Results() *types.ScanResults {
	s.Lock()
	defer s.Unlock()

	var imageScanResults []*types.ImageScanResult

	for _, scanD := range s.imageToScanData {
		if !scanD.completed {
			continue
		}
		for _, podContext := range scanD.contexts {
			imageScanResults = append(imageScanResults, &types.ImageScanResult{
				PodName:         podContext.podName,
				PodNamespace:    podContext.namespace,
				ImageName:       scanD.imageName,
				ContainerName:   podContext.containerName,
				ImageHash:       podContext.imageHash,
				PodUid:          podContext.podUid,
				Vulnerabilities: scanD.result,
				Success:         scanD.success,
				ScanErrMsg:      scanD.scanErrMsg,
			})
		}
	}

	return &types.ScanResults{
		ImageScanResults: imageScanResults,
		Progress:         s.ScanProgress(),
	}
}

func (s *Scanner) HandleResult(result *forwarding.ImageVulnerabilities) error {
	s.Lock()
	defer s.Unlock()

	scanD, ok := s.imageToScanData[result.Image]
	if !ok || scanD == nil {
		log.WithFields(s.logFields).Warnf("no scan data for image '%v', probably an old scan result - ignoring", result.Image)
		return nil
	}

	if result.ScanUUID != scanD.scanUUID {
		log.WithFields(s.logFields).Warnf("Scan UUID mismatch, probably an old scan result - ignoring. image=%v, received=%v, expected=%v", result.Image, result.ScanUUID, scanD.scanUUID)
		return nil
	}

	if scanD.timeout {
		log.WithFields(s.logFields).Warnf("Scan result after timeout - ignoring. image=%v, scan uuid=%v", result.Image, result.ScanUUID)
		return nil
	}

	if scanD.completed {
		log.WithFields(s.logFields).Warnf("Duplicate result for image scan. image=%v, scan uuid=%v", result.Image, result.ScanUUID)
		return nil
	}

	scanD.completed = true
	scanD.result = result.Vulnerabilities
	scanD.success = result.Success
	scanD.scanErrMsg = result.ScanErrMsg

	if scanD.success && scanD.result == nil {
		log.WithFields(s.logFields).Infof("No vulnerabilities found on image %v.", result.Image)
	}
	if !scanD.success {
		log.WithFields(s.logFields).Warnf("Scan of image %v has failed: %v", result.Image, scanD.scanErrMsg)
	}

	select {
	case scanD.resultChan <- true:
	default:
		log.WithFields(s.logFields).Warnf("Failed to notify upon received result scan. image=%v, scan-uuid=%v", result.Image, result.ScanUUID)
	}

	return nil
}

func (s *Scanner) Clear() {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Clearing...")
	close(s.killSignal)
}
