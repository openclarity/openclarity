package orchestrator

import (
	"fmt"
	klar "github.com/Portshift/klar/kubernetes"
	"github.com/Portshift/kubei/pkg/utils/k8s"
	stringutils "github.com/Portshift/kubei/pkg/utils/string"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (o *Orchestrator) jobBatchManagement() {
	defer func() {
		o.Lock()
		o.status = Idle
		o.Unlock()
	}()

	if len(o.imageToScanData) == 0 {
		log.Info("Nothing to scan")
		return
	}
	//channel for terminating the workers
	killsignal := make(chan bool)

	//queue of jobs
	q := make(chan *scanData)
	// done channel takes the result of the job
	done := make(chan bool)

	numberOfWorkers := o.scanConfig.MaxScanParallelism
	for i := 0; i < numberOfWorkers; i++ {
		go o.worker(q, i, done, killsignal)
	}

	for _, data := range o.imageToScanData {
		go func(data *scanData) {
			q <- data
			atomic.AddUint32(&o.progress.ImagesStartedToScan, 1)
		}(data)
	}

	for c := 0; c < len(o.imageToScanData); c++ {
		<-done
		atomic.AddUint32(&o.progress.ImagesCompletedToScan, 1)
	}

	log.Infof("All jobs has finished")

	// cleaning workers
	close(killsignal)
}

func (o *Orchestrator) worker(queue chan *scanData, worknumber int, done, ks chan bool) {
	for {
		select {
		case data := <-queue:
			err := o.runJob(data)
			if err != nil {
				log.Errorf("failed to run job: %v", err)
				o.Lock()
				data.success = false
				data.completed = true
				o.Unlock()
			} else {
				o.waitForResult(data)
			}
			done <- true
		case <-ks:
			log.Debugf("worker #%v halted", worknumber)
			return
		}
	}
}

func (o *Orchestrator) waitForResult(data *scanData) {
	log.Infof("Waiting for result. image=%+v", data.imageName)
	ticker := time.NewTicker(o.scanConfig.JobResultTimeout)
	select {
	case <-data.resultChan:
		log.Infof("Image scanned result has arrived. image=%v", data.imageName)
	case <-ticker.C:
		log.Warnf("job was timeout. image=%v", data.imageName)
	}
}

func (o *Orchestrator) runJob(data *scanData) error {
	job := o.createJob(data)
	log.Debugf("Created job=%+v", job)

	log.Infof("Running job %s/%s to scan image %s", job.GetNamespace(), job.GetName(), data.imageName)
	_, err := o.clientset.BatchV1().Jobs(job.GetNamespace()).Create(job)
	if err != nil {
		return fmt.Errorf("failed to create job: %v/%v. %v", job.GetName(), job.GetNamespace(), err)
	}

	return nil
}

// Due to K8s names constraint we will take the image name w/o the tag and repo
func getSimpleImageName(imageName string) string {
	repoEnd := strings.LastIndex(imageName, "/")
	imageName = imageName[repoEnd+1 :]

	digestStart := strings.LastIndex(imageName, "@sha256:")
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

func (o *Orchestrator) createContainer(imageName, secretName string, scanUUID string) corev1.Container {
	env := []corev1.EnvVar{
		{Name: "CLAIR_ADDR", Value: o.config.ClairAddress},
		{Name: "CLAIR_OUTPUT", Value: o.scanConfig.SeverityThreshold},
		{Name: "KLAR_TRACE", Value: strconv.FormatBool(o.config.KlarTrace)},
		{Name: "RESULT_SERVICE_PATH", Value: o.config.KlarResultServicePath},
		{Name: "SCAN_UUID", Value: scanUUID},
	}
	if secretName != "" {
		log.Debugf("Adding private registry credentials to image: %s", imageName)
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
		Image: o.scanConfig.KlarImageName,
		Args: []string{
			imageName,
		},
		Env: env,
	}
}

func (o *Orchestrator) createJob(data *scanData) *batchv1.Job {
	var ttlSecondsAfterFinished int32 = 300
	var backOffLimit int32 = 0

	// We will scan each image once, based on the first pod context. The result will be applied for all other pods with this image.
	podContext := data.contexts[0]

	labels := map[string]string{ignorePodScanLabelKey: ignorePodScanLabelValue}
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
					Containers:    []corev1.Container{o.createContainer(data.imageName, podContext.imagePullSecret, data.scanUUID)},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
}
