package scanner

import (
	"fmt"
	klar "github.com/Portshift/klar/kubernetes"
	"github.com/Portshift/kubei/pkg/config"
	"github.com/Portshift/kubei/pkg/utils/k8s"
	"github.com/Portshift/kubei/pkg/utils/proxyconfig"
	stringutils "github.com/Portshift/kubei/pkg/utils/string"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (s *Scanner) jobBatchManagement() {
	defer func() {
		s.Lock()
		s.status = Idle
		s.Unlock()
	}()

	s.Lock()
	imageToScanData := s.imageToScanData
	numberOfWorkers := s.scanConfig.MaxScanParallelism
	imagesStartedToScan := &s.progress.ImagesStartedToScan
	imagesCompletedToScan := &s.progress.ImagesCompletedToScan
	s.Unlock()

	if len(imageToScanData) == 0 {
		log.WithFields(s.logFields).Info("Nothing to scan")
		return
	}

	//queue of jobs
	q := make(chan *scanData)
	// done channel takes the result of the job
	done := make(chan bool)

	fullScanDone := make(chan bool)

	for i := 0; i < numberOfWorkers; i++ {
		go s.worker(q, i, done, s.killSignal)
	}

	go func() {
		for c := 0; c < len(imageToScanData); c++ {
			select {
			case <-done:
				atomic.AddUint32(imagesCompletedToScan, 1)
			case <-s.killSignal:
				log.WithFields(s.logFields).Debugf("Scan process was canceled - stop waiting for finished jobs")
				return
			}
		}

		fullScanDone <- true
	}()

	for _, data := range imageToScanData {
		go func(data *scanData, ks chan bool) {
			select {
			case q <- data:
				atomic.AddUint32(imagesStartedToScan, 1)
			case <-ks:
				log.WithFields(s.logFields).Debugf("Scan process was canceled. image name=%v, scanUUID=%v", data.imageName, data.scanUUID)
				return
			}
		}(data, s.killSignal)
	}

	select {
	case <-s.killSignal:
		log.WithFields(s.logFields).Info("Scan process was canceled")
	case <-fullScanDone:
		log.WithFields(s.logFields).Infof("All jobs has finished")
	}
}

func (s *Scanner) worker(queue chan *scanData, worknumber int, done, ks chan bool) {
	for {
		select {
		case data := <-queue:
			job, err := s.runJob(data)
			if err != nil {
				log.WithFields(s.logFields).Errorf("failed to run job: %v", err)
				s.Lock()
				data.success = false
				data.completed = true
				s.Unlock()
			} else {
				s.waitForResult(data, ks)
			}

			s.deleteJobIfNeeded(job, data.success, data.completed)

			select {
			case done <- true:
			case <-ks:
				log.WithFields(s.logFields).Infof("Image scan was canceled. image=%v", data.imageName)
			}
		case <-ks:
			log.WithFields(s.logFields).Debugf("worker #%v halted", worknumber)
			return
		}
	}
}

func (s *Scanner) waitForResult(data *scanData, ks chan bool) {
	log.WithFields(s.logFields).Infof("Waiting for result. image=%+v", data.imageName)
	ticker := time.NewTicker(s.scanConfig.JobResultTimeout)
	select {
	case <-data.resultChan:
		log.WithFields(s.logFields).Infof("Image scanned result has arrived. image=%v", data.imageName)
	case <-ticker.C:
		log.WithFields(s.logFields).Warnf("job was timeout. image=%v", data.imageName)
		s.Lock()
		data.success = false
		data.timeout = true
		data.completed = true
		s.Unlock()
	case <-ks:
		log.WithFields(s.logFields).Infof("Image scan was canceled. image=%v", data.imageName)
	}
}

func (s *Scanner) runJob(data *scanData) (*batchv1.Job, error) {
	job := s.createJob(data)
	log.WithFields(s.logFields).Debugf("Created job=%+v", job)

	log.WithFields(s.logFields).Infof("Running job %s/%s to scan image %s", job.GetNamespace(), job.GetName(), data.imageName)
	_, err := s.clientset.BatchV1().Jobs(job.GetNamespace()).Create(job)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %s/%s. %v", job.GetNamespace(), job.GetName(), err)
	}

	return job, nil
}

func (s *Scanner) deleteJobIfNeeded(job *batchv1.Job, isSuccessfulJob, isCompletedJob bool) {
	if job == nil {
		return
	}

	// delete uncompleted jobs - scan process was canceled
	if !isCompletedJob {
		s.deleteJob(job)
		return
	}

	switch s.scanConfig.DeleteJobPolicy {
	case config.DeleteJobPolicyAll:
		s.deleteJob(job)
	case config.DeleteJobPolicySuccessful:
		if isSuccessfulJob {
			s.deleteJob(job)
		}
	}
}

func (s *Scanner) deleteJob(job *batchv1.Job) {
	dpb := metav1.DeletePropagationBackground
	deleteOptions := &metav1.DeleteOptions{PropagationPolicy: &dpb}

	log.WithFields(s.logFields).Infof("Deleting job %s/%s", job.GetNamespace(), job.GetName())
	err := s.clientset.BatchV1().Jobs(job.GetNamespace()).Delete(job.GetName(), deleteOptions)
	if err != nil && !errors.IsNotFound(err) {
		log.WithFields(s.logFields).Errorf("failed to delete job: %s/%s. %v", job.GetNamespace(), job.GetName(), err)
	}
}

// Due to K8s names constraint we will take the image name w/o the tag and repo
func getSimpleImageName(imageName string) string {
	repoEnd := strings.LastIndex(imageName, "/")
	imageName = imageName[repoEnd+1 :]

	digestStart := strings.LastIndex(imageName, "@")
	// remove digest if exists
	if digestStart != -1 {
		return imageName[:digestStart]
	}

	tagStart := strings.LastIndex(imageName, ":")
	// remove tag if exists
	if tagStart != -1 {
		return imageName[:tagStart]
	}

	return imageName
}

// Job names require their names to follow the DNS label standard as defined in RFC 1123
// Note: job name is added as a label to the pod template spec so it should follow the DNS label standard and not just DNS-1123 subdomain
//
// This means the name must:
// * contain at most 63 characters
// * contain only lowercase alphanumeric characters or ‘-’
// * start with an alphanumeric character
// * end with an alphanumeric character
func createJobName(imageName string) string {
	jobName := jobContainerName+"-"+getSimpleImageName(imageName)+"-"+uuid.NewV4().String()

	// contain at most 63 characters
	jobName = stringutils.TruncateString(jobName, k8s.MaxK8sJobName)

	// contain only lowercase alphanumeric characters or ‘-’
	jobName = strings.ToLower(jobName)
	jobName = strings.ReplaceAll(jobName, "_", "-")

	// no need to validate start, we are using 'jobContainerName'

	// end with an alphanumeric character
	jobName = strings.TrimRight(jobName, "-")

	return jobName
}

const jobContainerName = "klar-scanner"

func (s *Scanner) createContainer(imageName, secretName string, scanUUID string) corev1.Container {
	env := []corev1.EnvVar{
		{Name: "CLAIR_ADDR", Value: s.config.ClairAddress},
		{Name: "CLAIR_OUTPUT", Value: s.scanConfig.SeverityThreshold},
		{Name: "KLAR_TRACE", Value: strconv.FormatBool(s.config.KlarTrace)},
		{Name: "RESULT_SERVICE_PATH", Value: s.config.KlarResultServicePath},
		{Name: "SCAN_UUID", Value: scanUUID},
	}

	env = s.appendProxyEnvConfig(env)

	if secretName != "" {
		log.WithFields(s.logFields).Debugf("Adding private registry credentials to image: %s", imageName)
		env = append(env, corev1.EnvVar{
			Name: klar.ImagePullSecretEnvVar, ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: corev1.DockerConfigJsonKey,
				},
			},
		})
	}

	return corev1.Container{
		Name:  jobContainerName,
		Image: s.scanConfig.KlarImageName,
		Args: []string{
			imageName,
		},
		Env: env,
	}
}

func (s *Scanner) appendProxyEnvConfig(env []corev1.EnvVar) []corev1.EnvVar {
	if s.config.ScannerHttpsProxy == "" && s.config.ScannerHttpProxy == "" {
		return env
	}

	if s.config.ScannerHttpsProxy != "" {
		env = append(env, corev1.EnvVar{
			Name: proxyconfig.HttpsProxyEnvCaps, Value: s.config.ScannerHttpsProxy,
		})
	}

	if s.config.ScannerHttpProxy != "" {
		env = append(env, corev1.EnvVar{
			Name: proxyconfig.HttpProxyEnvCaps, Value: s.config.ScannerHttpProxy,
		})
	}

	env = append(env, corev1.EnvVar{
		Name: proxyconfig.NoProxyEnvCaps, Value: s.config.KlarResultServiceAddress + "," + s.config.ClairAddress,
	})

	return env
}

func (s *Scanner) createJob(data *scanData) *batchv1.Job {
	var ttlSecondsAfterFinished int32 = 300
	var backOffLimit int32 = 0

	// We will scan each image once, based on the first pod context. The result will be applied for all other pods with this image.
	podContext := data.contexts[0]

	labels := map[string]string{
		"app": jobContainerName,
		ignorePodScanLabelKey: ignorePodScanLabelValue,
	}
	annotations := map[string]string{
		"sidecar.istio.io/inject": "false",
		"sidecar.portshift.io/inject": "false",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createJobName(data.imageName),
			Namespace: podContext.namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{s.createContainer(data.imageName, podContext.imagePullSecret, data.scanUUID)},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
}
