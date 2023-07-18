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

package scanner

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/kubeclarity/runtime_scan/api/server/models"
	"github.com/openclarity/kubeclarity/runtime_scan/api/server/restapi/operations"
	_config "github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	_creds "github.com/openclarity/kubeclarity/runtime_scan/pkg/scanner/creds"
	_types "github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
	sliceutils "github.com/openclarity/kubeclarity/runtime_scan/pkg/utils/slice"
	k8sutils "github.com/openclarity/kubeclarity/shared/pkg/utils/k8s"
)

type Scanner struct {
	imageIDToScanData  map[string]*scanData
	progress           _types.ScanProgress
	scannerJobTemplate *batchv1.Job
	scanConfig         *_config.ScanConfig
	killSignal         chan bool
	killChanClosed     atomic.Bool
	clientset          kubernetes.Interface
	logFields          log.Fields
	credentialAdders   []_creds.CredentialAdder
	sync.Mutex
}

func CreateScanner(config *_config.Config, clientset kubernetes.Interface) *Scanner {
	credentialAdders := []_creds.CredentialAdder{
		_creds.CreateBasicRegCred(clientset, config.CredsSecretNamespace),
		_creds.CreateECR(clientset, config.CredsSecretNamespace),
		_creds.CreateGCR(clientset, config.CredsSecretNamespace),
	}
	s := &Scanner{
		progress: _types.ScanProgress{
			Status: _types.Idle,
		},
		scannerJobTemplate: config.ScannerJobTemplate,
		killSignal:         make(chan bool),
		clientset:          clientset,
		logFields:          log.Fields{"scanner id": uuid.NewV4().String()},
		credentialAdders:   credentialAdders,
		Mutex:              sync.Mutex{},
	}

	return s
}

type imagePodContext struct {
	containerName    string
	podName          string
	namespace        string
	imagePullSecrets []string
	imageName        string
	podUID           string
	podLabels        labels.Set
}

type vulnerabilitiesScanResult struct {
	result        []*models.PackageVulnerabilityScan
	layerCommands []*models.ResourceLayerCommand
	success       bool
	completed     bool
	error         *models.ScanError
}

type cisDockerBenchmarkScanResult struct {
	result    []*models.CISDockerBenchmarkCodeInfo
	success   bool
	completed bool
	error     *models.ScanError
}

type scanData struct {
	imageHash                    string
	imageID                      string
	contexts                     []*imagePodContext // All the pods that contain this image hash
	scanUUID                     string
	vulnerabilitiesResult        vulnerabilitiesScanResult
	cisDockerBenchmarkResult     cisDockerBenchmarkScanResult
	shouldScanCISDockerBenchmark bool
	resultChan                   chan bool
	success                      bool
	completed                    bool
	timeout                      bool
	scanErr                      *_types.ScanError
}

func (sd *scanData) getScanErrors() []*_types.ScanError {
	var errors []*_types.ScanError

	if sd.scanErr != nil {
		errors = append(errors, sd.scanErr)
	}

	if sd.vulnerabilitiesResult.error != nil {
		errors = append(errors, &_types.ScanError{
			ErrMsg:    sd.vulnerabilitiesResult.error.Message,
			ErrType:   string(sd.vulnerabilitiesResult.error.Type),
			ErrSource: _types.ScanErrSourceVul,
		})
	}

	if sd.cisDockerBenchmarkResult.error != nil {
		errors = append(errors, &_types.ScanError{
			ErrMsg:    sd.cisDockerBenchmarkResult.error.Message,
			ErrType:   string(sd.cisDockerBenchmarkResult.error.Type),
			ErrSource: _types.ScanErrSourceDockle,
		})
	}

	return errors
}

func (sd *scanData) setVulnerabilitiesResult(result *vulnerabilitiesScanResult) {
	sd.vulnerabilitiesResult = *result
	sd.updateResult()
}

func (sd *scanData) setCISDockerBenchmarkResult(result *cisDockerBenchmarkScanResult) {
	sd.cisDockerBenchmarkResult = *result
	sd.updateResult()
}

func (sd *scanData) updateResult() {
	if sd.vulnerabilitiesResult.completed && (!sd.shouldScanCISDockerBenchmark || sd.cisDockerBenchmarkResult.completed) {
		sd.completed = true
	}
	if sd.vulnerabilitiesResult.success && (!sd.shouldScanCISDockerBenchmark || sd.cisDockerBenchmarkResult.success) {
		sd.success = true
	}
}

const (
	ignorePodScanLabelKey   = "kubeclarityShouldScan"
	ignorePodScanLabelValue = "false"
)

func (s *Scanner) shouldIgnorePod(pod *corev1.Pod) bool {
	if sliceutils.ContainsString(s.scanConfig.IgnoredNamespaces, pod.Namespace) {
		log.WithFields(s.logFields).Infof("Skipping pod scan, namespace is in the ignored namespaces list.  pod=%v ,namespace=%s", pod.Name, pod.Namespace)
		return true
	}
	if pod.Labels != nil && pod.Labels[ignorePodScanLabelKey] == ignorePodScanLabelValue {
		log.WithFields(s.logFields).Infof("Skipping pod scan, pod has an ignore label. pod=%v ,namespace=%s", pod.Name, pod.Namespace)
		return true
	}

	return false
}

// initScan Calculate properties of scan targets
// nolint:cyclop
func (s *Scanner) initScan() error {
	var podsToScan []corev1.Pod

	// Get all target pods
	for _, namespace := range s.scanConfig.TargetNamespaces {
		podList, err := s.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods. namespace=%s: %v", namespace, err)
		}
		podsToScan = append(podsToScan, podList.Items...)
		if namespace == corev1.NamespaceAll {
			break
		}
	}

	imageIDToScanData := make(map[string]*scanData)

	// Populate the image to scanData map from all target pods
	for i, pod := range podsToScan {
		if s.shouldIgnorePod(&podsToScan[i]) {
			continue
		}

		// Fetch all image pull secrets from the pod so that we can
		// inform the scanner where to check for credentials to pull
		// the image to scan.
		imagePullSecretNames := []string{}
		imagePullSecretNamesSet := make(map[string]struct{})
		for _, ips := range pod.Spec.ImagePullSecrets {
			// avoid cases where a pod has the same imagePullSecret more than once.
			if _, ok := imagePullSecretNamesSet[ips.Name]; ok {
				log.WithFields(s.logFields).Warnf("Duplicate image pull secret name: %v", ips.Name)
				continue
			}
			imagePullSecretNamesSet[ips.Name] = struct{}{}
			imagePullSecretNames = append(imagePullSecretNames, ips.Name)
		}

		// Due to scenarios where image name in the `pod.Status.ContainerStatuses` is different
		// from image name in the `pod.Spec.Containers` we will take only image id from `pod.Status.ContainerStatuses`.
		containerNameToImageID := make(map[string]string)
		for _, container := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
			containerNameToImageID[container.Name] = k8sutils.NormalizeImageID(container.ImageID)
		}

		containers := append(pod.Spec.Containers, pod.Spec.InitContainers...)

		for _, container := range containers {
			imageID, ok := containerNameToImageID[container.Name]
			if !ok {
				log.Warnf("Image id is missing. pod=%v, namepspace=%v, container=%v ,image=%v",
					pod.GetName(), pod.GetNamespace(), container.Name, container.Image)
				continue
			}
			imageHash := k8sutils.ParseImageHash(imageID)
			if imageHash == "" {
				log.WithFields(s.logFields).Warnf("Failed to get image hash - ignoring image. "+
					"pod=%v, namepspace=%v, image name=%v", pod.GetName(), pod.GetNamespace(), container.Image)
				continue
			}
			// Create pod context
			podContext := &imagePodContext{
				containerName:    container.Name,
				podName:          pod.GetName(),
				namespace:        pod.GetNamespace(),
				imagePullSecrets: imagePullSecretNames,
				imageName:        container.Image,
				podUID:           string(pod.GetUID()),
				podLabels:        labels.Set(pod.GetLabels()),
			}
			if data, ok := imageIDToScanData[imageID]; !ok {
				// Image added for the first time, create scan data and append pod context
				imageIDToScanData[imageID] = &scanData{
					imageHash:                    imageHash,
					imageID:                      imageID,
					contexts:                     []*imagePodContext{podContext},
					scanUUID:                     uuid.NewV4().String(),
					shouldScanCISDockerBenchmark: s.scanConfig.ShouldScanCISDockerBenchmark,
					resultChan:                   make(chan bool),
				}
			} else {
				// Image already exist in map, just append the pod context
				data.contexts = append(data.contexts, podContext)
			}
		}
	}

	s.imageIDToScanData = imageIDToScanData
	s.progress.ImagesToScan = uint32(len(imageIDToScanData))

	log.WithFields(s.logFields).Infof("Total %d unique images to scan", s.progress.ImagesToScan)

	return nil
}

func (s *Scanner) Scan(scanConfig *_config.ScanConfig) (chan struct{}, error) {
	s.Lock()
	defer s.Unlock()

	s.scanConfig = scanConfig
	log.WithFields(s.logFields).Infof("Start scanning...")
	s.progress.Status = _types.ScanInit
	if err := s.initScan(); err != nil {
		s.progress.SetStatus(_types.ScanInitFailure)
		return nil, fmt.Errorf("failed to initiate scan: %v", err)
	}

	// Create channel for caller to use to track doneness
	done := make(chan struct{})
	if s.progress.ImagesToScan == 0 {
		log.WithFields(s.logFields).Info("Nothing to scan")
		s.progress.SetStatus(_types.NothingToScan)
		// Close done channel to indicate that the scan is no longer running
		close(done)
		return done, nil
	}

	s.progress.SetStatus(_types.Scanning)
	go func() {
		// jobBatchManagement blocks until scan is completed or aborted
		s.jobBatchManagement()
		// close done channel to indicate that the scan is no longer running
		close(done)
	}()
	return done, nil
}

func (s *Scanner) ScanProgress() _types.ScanProgress {
	return _types.ScanProgress{
		ImagesToScan:          s.progress.ImagesToScan,
		ImagesStartedToScan:   atomic.LoadUint32(&s.progress.ImagesStartedToScan),
		ImagesCompletedToScan: atomic.LoadUint32(&s.progress.ImagesCompletedToScan),
		Status:                s.progress.Status,
	}
}

func (s *Scanner) Results() *_types.ScanResults {
	s.Lock()
	defer s.Unlock()

	var imageScanResults []*_types.ImageScanResult

	for _, scanD := range s.imageIDToScanData {
		if !scanD.completed {
			continue
		}
		for _, podContext := range scanD.contexts {
			imageScanResults = append(imageScanResults, &_types.ImageScanResult{
				PodName:                  podContext.podName,
				PodNamespace:             podContext.namespace,
				ImageName:                podContext.imageName,
				ContainerName:            podContext.containerName,
				ImageHash:                scanD.imageHash,
				PodUID:                   podContext.podUID,
				PodLabels:                podContext.podLabels,
				Vulnerabilities:          scanD.vulnerabilitiesResult.result,
				CISDockerBenchmarkResult: scanD.cisDockerBenchmarkResult.result,
				LayerCommands:            scanD.vulnerabilitiesResult.layerCommands,
				Success:                  scanD.success,
				ScanErrors:               scanD.getScanErrors(),
			})
		}
	}

	return &_types.ScanResults{
		ImageScanResults: imageScanResults,
		Progress:         s.ScanProgress(),
	}
}

func (s *Scanner) shouldIgnoreResult(scanD *scanData, resultScanUUID string, imageID string) bool {
	if scanD == nil {
		log.WithFields(s.logFields).Warnf("no scan data for imageID %q, probably an old scan result - ignoring",
			imageID)
		return true
	}

	if resultScanUUID != scanD.scanUUID {
		log.WithFields(s.logFields).Warnf("Scan UUID mismatch, probably an old scan result - ignoring. "+
			"imageID=%v, received=%v, expected=%v", imageID, resultScanUUID, scanD.scanUUID)
		return true
	}

	if scanD.timeout {
		log.WithFields(s.logFields).Warnf("Scan result after timeout - ignoring. imageID=%v, scan uuid=%v",
			imageID, resultScanUUID)
		return true
	}

	if scanD.completed {
		log.WithFields(s.logFields).Warnf("Duplicate result for image scan. imageID=%v, scan uuid=%v",
			imageID, resultScanUUID)
		return true
	}

	return false
}

func (s *Scanner) HandleScanResults(params operations.PostScanScanUUIDResultsParams) error {
	s.Lock()
	defer s.Unlock()

	resourceType := params.Body.ResourceVulnerabilityScan.Resource.ResourceType
	if resourceType != models.ResourceTypeIMAGE {
		return fmt.Errorf("resource type %q is not supported", resourceType)
	}

	imageID := params.Body.ImageID
	scanUUID := string(params.ScanUUID)

	scanD, ok := s.imageIDToScanData[imageID]
	if !ok || s.shouldIgnoreResult(scanD, scanUUID, imageID) {
		log.WithFields(s.logFields).Warnf("Ignoring vulnerabilities result for imageID %q", imageID)
		return nil
	}

	vulnerabilitiesResult := &vulnerabilitiesScanResult{
		result:        params.Body.ResourceVulnerabilityScan.PackageVulnerabilities,
		layerCommands: params.Body.ResourceVulnerabilityScan.ResourceLayerCommands,
		success:       params.Body.ResourceVulnerabilityScan.Status == models.ScanStatusSUCCESS,
		completed:     true,
		error:         params.Body.ResourceVulnerabilityScan.Error,
	}

	log.WithFields(s.logFields).Tracef("Vulnerabilities result recevied for imageID %q. result=%+v",
		imageID, params.Body.ResourceVulnerabilityScan)

	scanD.setVulnerabilitiesResult(vulnerabilitiesResult)
	log.WithFields(s.logFields).Infof("Vulnerabilities result was set for imageID %q", imageID)

	if scanD.vulnerabilitiesResult.success && len(params.Body.ResourceVulnerabilityScan.PackageVulnerabilities) == 0 {
		log.WithFields(s.logFields).Infof("No vulnerabilities found on imageID %q.", imageID)
	}

	if !scanD.vulnerabilitiesResult.success {
		log.WithFields(s.logFields).Warnf("Vulnerabilities scan of imageID %q has failed: %v",
			imageID, scanD.vulnerabilitiesResult.error)
	}

	if !scanD.completed {
		log.WithFields(s.logFields).Infof("Total scan is not yet completed for imageID %q", imageID)
		return nil
	}

	select {
	case scanD.resultChan <- true:
	default:
		log.WithFields(s.logFields).Warnf("Failed to notify upon received result scan. imageID=%v, scan-uuid=%v",
			imageID, scanUUID)
	}

	return nil
}

func (s *Scanner) ShouldHandleScanContentAnalysis(params operations.PostScanScanUUIDContentAnalysisParams) bool {
	s.Lock()
	defer s.Unlock()

	resourceType := params.Body.ResourceContentAnalysis.Resource.ResourceType
	if resourceType != models.ResourceTypeIMAGE {
		log.Infof("Resource type %q is not supported", resourceType)
		return false
	}

	imageID := params.Body.ImageID
	scanUUID := string(params.ScanUUID)

	scanD, ok := s.imageIDToScanData[imageID]
	if !ok || s.shouldIgnoreResult(scanD, scanUUID, imageID) {
		log.WithFields(s.logFields).Warnf("Ignoring content analysis for imageID %q", imageID)
		return false
	}

	return true
}

func (s *Scanner) HandleCISDockerBenchmarkScanResult(params operations.PostScanScanUUIDCisDockerBenchmarkResultsParams) error {
	s.Lock()
	defer s.Unlock()

	imageID := params.Body.ImageID
	scanUUID := string(params.ScanUUID)

	scanD, ok := s.imageIDToScanData[imageID]
	if !ok || s.shouldIgnoreResult(scanD, scanUUID, imageID) {
		log.WithFields(s.logFields).Warnf("Ignoring cis docker benchmark result for imageID %q", imageID)
		return nil
	}

	scanResult := &cisDockerBenchmarkScanResult{
		result:    params.Body.CisDockerBenchmarkScanResult.Result,
		success:   params.Body.CisDockerBenchmarkScanResult.Status == models.ScanStatusSUCCESS,
		completed: true,
		error:     params.Body.CisDockerBenchmarkScanResult.Error,
	}

	scanD.setCISDockerBenchmarkResult(scanResult)
	log.WithFields(s.logFields).Infof("CIS docker benchmark result was set for imageID %q", imageID)

	if scanD.cisDockerBenchmarkResult.success && len(scanD.cisDockerBenchmarkResult.result) == 0 {
		log.WithFields(s.logFields).Infof("No checkpoints found on imageID %q.", imageID)
	}

	if !scanD.cisDockerBenchmarkResult.success {
		log.WithFields(s.logFields).Warnf("CIS docker benchmark scan of imageID %q has failed: %v",
			imageID, scanD.cisDockerBenchmarkResult.error)
	}

	if !scanD.completed {
		log.WithFields(s.logFields).Infof("Total scan is not yet completed for imageID %q", imageID)
		return nil
	}

	select {
	case scanD.resultChan <- true:
	default:
		log.WithFields(s.logFields).Warnf("Failed to notify upon received result scan. imageID=%v, scan-uuid=%v",
			imageID, scanUUID)
	}

	return nil
}

func (s *Scanner) Clear() {
	s.Lock()
	defer s.Unlock()

	log.WithFields(s.logFields).Infof("Clearing...")
	// Make sure to close the channel only once.
	if !s.killChanClosed.Load() {
		close(s.killSignal)
		s.killChanClosed.Store(true)
	}
}
