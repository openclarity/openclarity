package orchestrator

import (
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"time"
)

func (orc *Orchestrator) jobBatchManagement() {
	if len(orc.imageToScanData) == 0 {
		log.Info("Nothing to scan")
		return
	}
	//channel for terminating the workers
	killsignal := make(chan bool)

	//queue of jobs
	q := make(chan *scanData)
	// done channel takes the result of the job
	done := make(chan bool)

	numberOfWorkers := 10 //os.Getenv("MAX_PARALLELISM") // TODO: get from config
	for i := 0; i < numberOfWorkers; i++ {
		go orc.worker(q, i, done, killsignal)
	}

	for _, data := range orc.imageToScanData {
		go func(data *scanData) {
			q <- data
		}(data)
	}

	// a deadlock occurs if c >= numberOfJobs
	for c := 0; c < len(orc.imageToScanData); c++ {
		<-done
	}

	log.Infof("all jobs has finished")

	// cleaning workers
	close(killsignal)
}

func (orc *Orchestrator) worker(queue chan *scanData, worknumber int, done, ks chan bool) {
	for {
		select {
		case data := <-queue:
			err := orc.runJob(data)
			if err != nil {
				log.Errorf("failed to run job: %v", err)
				// todo: do we need to report it to the webapp?
			} else {
				waitForResult(data)
			}
			done <- true
		case <-ks:
			log.Infof("worker #%v halted", worknumber)
			return
		}
	}
}

func waitForResult(data *scanData) {
	log.Infof("Waiting for result. image=%+v", data.imageName)
	ticker := time.NewTicker(5 * time.Minute) // todo: should be configurable
	select {
	case <-data.resultChan:
		log.Infof("Image scanned result has arrived. image=%v", data.imageName)
	case <-ticker.C:
		log.Warnf("job was timeout. image=%v", data.imageName)
	}
}

func (orc *Orchestrator) runJob(data *scanData) error {
	job := orc.createJobNew(data)

	log.Infof("Creating job=%+v", job)
	//_, err := orc.ExecutionConfig.Clientset.BatchV1().Jobs(job.GetNamespace()).Create(job)
	//if err != nil {
	//	return fmt.Errorf("failed to create job: %v/%v. %v", job.GetName(), job.GetNamespace(), err)
	//}

	return nil
}


const maxK8sJobName = 63
// TODO: move to utils
func truncateString(str string, maxsize int) string {
	if len(str) > maxsize {
		return str[0:maxsize]
	}

	return str
}

func createJobName(imageName string) string {
	return truncateString("klar-scan-" + imageName + "-" + uuid.NewV4().String(), maxK8sJobName)
}

const jobContainerName = "klar-scanner"
func (orc *Orchestrator) createContainer(imageName, secretName string) corev1.Container {
	env := []corev1.EnvVar{
		{Name: "CLAIR_ADDR", Value: "clair.kubei" }, // TODO: Get from config
		{Name: "CLAIR_OUTPUT", Value: orc.ExecutionConfig.ClairOutput},  // TODO: Get from config
		{Name: "KLAR_TRACE", Value: strconv.FormatBool(orc.ExecutionConfig.KlarTrace)},  // TODO: Get from config
		{Name: "WHITELIST_FILE", Value: orc.ExecutionConfig.WhitelistFile},  // TODO: Get from config
	}
	if secretName != "" {
		log.Debugf("Adding private registry credentials to image: %s", imageName)
		env = append(env, corev1.EnvVar{
			Name: "K8S_IMAGE_PULL_SECRET", ValueFrom: &corev1.EnvVarSource{ // todo: K8S_IMAGE_PULL_SECRET should be const on klar
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
		Image: "rafiportshift/portshift-klar:1.0.0", // TODO: XXX
		Args:  []string{
			imageName, // image to scan
			"kubei.kubei" , // result service address // todo: should be full path including port
		},
		Env:   env,
	}
}

// todo: change name
func (orc *Orchestrator) createJobNew(data *scanData) *batchv1.Job {
	var ttlSecondsAfterFinished int32 = 300
	var backOffLimit int32 = 0

	k8sContext := data.k8sContext[0]

	labels := map[string]string{"kubeiShouldScan": "false"}
	annotations := map[string]string{"sidecar.istio.io/inject": "false"}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createJobName(data.imageName),
			Namespace: k8sContext.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: annotations,
					Labels:      labels,
				},
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{orc.createContainer(data.imageName, k8sContext.Secret)},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
}
