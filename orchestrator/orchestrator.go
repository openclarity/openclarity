package orchestrator

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubei/common"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Orchestrator struct {
	ImageK8ExtendedContextMap common.ImageK8ExtendedContextMap
	DataUpdateLock            *sync.Mutex
	ExecutionConfig           *common.ExecutionConfiguration
	scanIssuesMessages        *[]string
	batchCompletedScansCount  *int32
	k8ContextService          common.K8ContextServiceInterface
}

type OrchestratorInterface interface {
	Scan()
}

const line = "_________________________________________________________________________________________________________"
const webappServiceName = "kubei-service"

func (orc *Orchestrator) getPodsImagesDetails(pods []corev1.Pod) (common.ImageNamespacesMap, common.NamespacedImageSecretMap, error) {
	log.Infof("There are %d pods in the given namespaces scope", len(pods))
	totalContainers := 0
	imageNamespacesMap := make(common.ImageNamespacesMap)
	namespacedImageSecretMap := make(common.NamespacedImageSecretMap)
	containerImagesSet := make(map[common.ContainerImageName]bool)
	for _, pod := range pods {
		imageNamespacesMap, namespacedImageSecretMap, containerImagesSet, totalContainers = orc.k8ContextService.GetK8ContextFromContainer(orc.ImageK8ExtendedContextMap, &pod, imageNamespacesMap, namespacedImageSecretMap, containerImagesSet, totalContainers)
	}

	log.Infof("There are %d containers in the given namespaces scope", totalContainers)
	log.Infof("There are %d different images in the given namespaces scope", len(containerImagesSet))
	return imageNamespacesMap, namespacedImageSecretMap, nil
}

func (orc *Orchestrator) getImageDetails() (common.ImageNamespacesMap, common.NamespacedImageSecretMap, error) {
	podList, err := orc.ExecutionConfig.Clientset.CoreV1().Pods(orc.ExecutionConfig.TargetNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list pods in namespace %s: %v", orc.ExecutionConfig.TargetNamespace, err)
	}

	pods := podList.Items
	imageNamespacesMap, namespacedImageSecretMap, err := orc.getPodsImagesDetails(pods)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get image details in namespace %s: %v", orc.ExecutionConfig.TargetNamespace, err)
	}

	orc.printAllImages()

	return imageNamespacesMap, namespacedImageSecretMap, nil
}

func (orc *Orchestrator) waitForServiceAccount(serviceAccountName string, namespace string) bool {
	for i := 0; i < 30; i++ { //30 * 1s = 30s = 0.5m
		response, _ := orc.ExecutionConfig.Clientset.CoreV1().ServiceAccounts(namespace).Get(serviceAccountName, metav1.GetOptions{})
		if response != nil {
			log.Debugf("Service account kubei in namespace %s is ready!", namespace)
			return true
		}
		time.Sleep(1 * time.Second)
	}

	return false
}

func (orc *Orchestrator) runJobsBatch(totalImages int, batch []string, batchNum int, startPoint int, imageNamespace string, scannedImageNames []string, namespacedImageSecretMap common.NamespacedImageSecretMap) error {
	log.Infof("Processing batch %d of namespace %s. batch size is: %d. total images in namespace: %d.", batchNum, imageNamespace, len(batch), totalImages)

	err := orc.createJob(imageNamespace, batchNum, batch, startPoint, scannedImageNames, namespacedImageSecretMap)

	if err != nil {
		return err
	}
	return nil
}

func (orc *Orchestrator) createJob(imageNamespace string, batchNum int, batch []string, startPoint int, scannedImageNames []string, namespacedImageSecretMap common.NamespacedImageSecretMap) error {
	var ttlSecondsAfterFinished int32 = 300
	jobName := orc.getKubernetesCompliantJobName(imageNamespace, batchNum)

	containers := orc.buildContainersPart(imageNamespace, batch, startPoint, scannedImageNames, namespacedImageSecretMap)
	var backOffLimit int32
	backOffLimit = 0
	jobDefinition := orc.createJobDefinition(jobName, imageNamespace, containers, backOffLimit, ttlSecondsAfterFinished)
	_, err := orc.ExecutionConfig.Clientset.BatchV1().Jobs(imageNamespace).Create(jobDefinition)
	if err != nil {
		log.Errorf("failed to create jobs in namespace: %s. %v", imageNamespace, err)
		return err
	}
	return nil
}

func (orc *Orchestrator) createJobDefinition(jobName string, imageNamespace string, containers []corev1.Container, backOffLimit int32, ttlSecondsAfterFinished int32) *batchv1.Job {
	labels := map[string]string{"kubeiShouldScan": "false"}
	annotations := map[string]string{"sidecar.istio.io/inject": "false"}
	jobDefinition := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: imageNamespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: apiv1.PodSpec{
					Containers:    containers,
					RestartPolicy: apiv1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
	return jobDefinition
}

func (orc *Orchestrator) buildContainersPart(imageNamespace string, batch []string, startPoint int, scannedImageNames []string, namespacedImageSecretMap common.NamespacedImageSecretMap) []corev1.Container {
	var containers []apiv1.Container
	log.Infof("Processing batch of images:[")
	orc.printBatch(batch, startPoint)
	for _, image := range batch {
		if common.ContainsString(scannedImageNames, image) {
			continue
		}
		k8ExtendedContexts := orc.ImageK8ExtendedContextMap[common.ContainerImageName(image)]
		if k8ExtendedContexts != nil {
			secretName := namespacedImageSecretMap[image+"_"+imageNamespace]
			containerName := orc.getKubernetesCompliantContainerName(common.ContainerImageName(image))
			clairServiceAddress := "clairsvc." + orc.ExecutionConfig.KubeiNamespace
			env := []corev1.EnvVar{
				{Name: "CLAIR_ADDR", Value: clairServiceAddress},
				{Name: "CLAIR_OUTPUT", Value: orc.ExecutionConfig.ClairOutput},
				{Name: "KLAR_TRACE", Value: strconv.FormatBool(orc.ExecutionConfig.KlarTrace)},
				{Name: "WHITELIST_FILE", Value: orc.ExecutionConfig.WhitelistFile},
			}
			if secretName != "" {
				log.Debugf("Adding private registry credentials to image: %s", image)
				env = append(env, corev1.EnvVar{
					Name: "K8S_IMAGE_PULL_SECRET", ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: corev1.DockerConfigJsonKey,
						},
					},
				})
			}
			orchestratorPodNamespace := os.Getenv("MY-POD-NAMESPACE")
			container := apiv1.Container{
				Name:  containerName,
				Image: "rafiportshift/portshift-klar:1.0.0",
				Args:  []string{image, webappServiceName + "." + orchestratorPodNamespace},
				Env:   env,
			}
			containers = append(containers, container)
		} else {
			log.Warnf("image %s didn't have a context. will not be scanned", image)
		}

		scannedImageNames = append(scannedImageNames, image)
	}

	return containers
}

func (orc *Orchestrator) getDistinctUnscannedImagesForBatch(batch []common.ContainerImageName, scannedImageNames []string) []string {
	var distinctUnscannedImages []string
	for _, imageName := range batch {
		if !common.ContainsString(scannedImageNames, string(imageName)) && !common.ContainsString(distinctUnscannedImages, string(imageName)) {
			distinctUnscannedImages = append(distinctUnscannedImages, string(imageName))
		}
	}
	return distinctUnscannedImages
}

func (orc *Orchestrator) getKubernetesCompliantContainerName(imageName common.ContainerImageName) string {
	startIndex := strings.LastIndex(string(imageName), "/") + 1
	endIndex := strings.LastIndex(string(imageName), ":")
	if endIndex == -1 {
		endIndex = len(string(imageName))
	}
	simpleImageName := string(imageName)[startIndex:endIndex]       ////kubernetes constraint + we just wanted the image name
	simpleImageName = strings.ReplaceAll(simpleImageName, "_", "-") //kubernetes constraint
	containerName := "klar-" + simpleImageName + "-" + uuid.NewV4().String()
	return containerName[0:orc.minInt(len(containerName), 63)] //kubernetes constraint
}

func (orc *Orchestrator) minInt(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func (orc *Orchestrator) getKubernetesCompliantJobName(imageNamespace string, batchNum int) string {
	jobName := "klar-job-" + imageNamespace + "-" + strconv.Itoa(batchNum) + "-" + uuid.NewV4().String()
	return jobName[0:orc.minInt(len(jobName), 63)] //kubernetes constraint
}

func (orc *Orchestrator) waitForJobs(batchSize int, timeoutTime time.Time, batchNum int, imageNamespace string) {
	for int(atomic.LoadInt32(orc.batchCompletedScansCount)) < batchSize {
		log.Infof("scanning... num of completed scans: %d. batch size: %d", int(atomic.LoadInt32(orc.batchCompletedScansCount)), batchSize)
		time.Sleep(2 * time.Second)
		if time.Now().After(timeoutTime) {
			log.Warnf("batch %d in namespace %s has timed out! moving to next batch", batchNum, imageNamespace)
			break
		}
	}
	log.Infof("scanning... num of completed scans: %d. batch size: %d", batchSize, batchSize)
	log.Infof("batch %d in namespace %s is done! ", batchNum, imageNamespace)
	*orc.batchCompletedScansCount = 0
}

func (orc *Orchestrator) runJobs(imageNames []string, imageNamespace string, scannedImageNames []string, namespacedImageSecretMap common.NamespacedImageSecretMap) {
	totalImages := len(imageNames)
	batchNum := 1
	for i := 0; i < totalImages; i += orc.ExecutionConfig.Parallelism {
		j := i + orc.ExecutionConfig.Parallelism
		if j > totalImages {
			j = totalImages
		}

		batch := imageNames[i:j]
		log.Infof(line)

		beforeExecutionTime := time.Now()
		timeoutTime := beforeExecutionTime.Add(30 * time.Minute) //todo maybe should be configurable?
		err := orc.runJobsBatch(totalImages, batch, batchNum, i, imageNamespace, scannedImageNames, namespacedImageSecretMap)
		if err == nil {
			orc.waitForJobs(len(batch), timeoutTime, batchNum, imageNamespace)
		} else {
			log.Errorf("failed to run batch. %+v", err)
		}
		batchNum++
		log.Infof(line)
	}
}

func (orc *Orchestrator) printAllImages() {
	imageNames := make([]string, 0, len(orc.ImageK8ExtendedContextMap))
	for k := range orc.ImageK8ExtendedContextMap {
		imageNames = append(imageNames, string(k))
	}
	log.Infof("ALL images:")
	orc.printBatch(imageNames, 0)
}

func (orc *Orchestrator) printBatch(batch []string, startPoint int) {
	log.Infof("[")
	for index, imageName := range batch {
		log.Infof("   %d)%s", startPoint+index+1, imageName)
	}
	log.Infof("]")
}

func (orc *Orchestrator) createKubeiServiceAccount(imageNamespace string) error {
	kubeiServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubei",
			Namespace: imageNamespace,
		},
	}

	_, err := orc.ExecutionConfig.Clientset.CoreV1().ServiceAccounts(imageNamespace).Create(kubeiServiceAccount)
	if err != nil {
		return fmt.Errorf("failed to create service account kubei in namespace %s: %v", imageNamespace, err)
	}

	ready := orc.waitForServiceAccount(kubeiServiceAccount.Name, imageNamespace)
	if !ready {
		return fmt.Errorf("failed to create service account. Creation has timed out")
	} else {
		log.Debugf("created service account kubei in namespace %s", imageNamespace)
	}

	return nil
}

func (orc *Orchestrator) executeScan() error {
	var scannedImageNames []string

	imageNamespacesMap, namespacedImageSecretMap, err := orc.getImageDetails()
	if err != nil {
		return err
	}

	for imageNamespace, imageNames := range imageNamespacesMap { //scan by namespace
		if imageNamespace != orc.ExecutionConfig.KubeiNamespace { //if not our own namespace
			distinctUnscannedImages := orc.getDistinctUnscannedImagesForBatch(imageNames, scannedImageNames)
			orc.runJobs(distinctUnscannedImages, imageNamespace, scannedImageNames, namespacedImageSecretMap)
		}

	}
	return nil
}

/******************************************************* PUBLIC *******************************************************/

func Init(executionConfig *common.ExecutionConfiguration, dataUpdateLock *sync.Mutex, imageK8ExtendedContextMap common.ImageK8ExtendedContextMap, scanIssuesMessages *[]string, batchCompletedScansCount *int32) *Orchestrator {
	return &Orchestrator{
		ImageK8ExtendedContextMap: imageK8ExtendedContextMap,
		DataUpdateLock:            dataUpdateLock,
		ExecutionConfig:           executionConfig,
		scanIssuesMessages:        scanIssuesMessages,
		batchCompletedScansCount:  batchCompletedScansCount,
		k8ContextService: &common.K8ContextService{
			ExecutionConfig:        executionConfig,
			K8ContextSecretService: &common.K8ContextSecretService{},
		},
	}
}

func (orc *Orchestrator) Scan() {
	err := orc.executeScan()
	if err != nil {
		log.Errorf("Failed to execute scan. %v", err)
		return
	}

	orc.DataUpdateLock.Lock()
	defer orc.DataUpdateLock.Unlock()
	if len(*orc.scanIssuesMessages) == 0 {
		fmt.Println() // fmt since we just want an empty line
		log.Infof("Scan has completed Successfully!!!")
	} else {
		fmt.Println() // fmt since we just want an empty line
		log.Warnf("Scan has completed with some issues.")
		log.Warnf("Summary:")
		for index, warning := range *orc.scanIssuesMessages {
			log.Warnf("   %d)%s", index+1, warning)
		}
		var slice []string
		orc.scanIssuesMessages = &slice
	}
}
