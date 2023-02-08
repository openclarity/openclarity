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
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/containers/image/v5/docker/reference"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openclarity/kubeclarity/runtime_scan/pkg/config"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
	stringsutils "github.com/openclarity/kubeclarity/runtime_scan/pkg/utils/strings"
	shared "github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/utils/k8s"
)

const (
	cisDockerBenchmarkScannerContainerName = "cis-docker-benchmark-scanner"
	localImageIDDockerPrefix               = "docker://sha256"
	localImageIDSHA256Prefix               = "sha256:"
)

// run jobs.
func (s *Scanner) jobBatchManagement() {
	s.Lock()
	imageIDToScanData := s.imageIDToScanData
	numberOfWorkers := s.scanConfig.MaxScanParallelism
	imagesStartedToScan := &s.progress.ImagesStartedToScan
	imagesCompletedToScan := &s.progress.ImagesCompletedToScan
	s.Unlock()

	// queue of scan data
	q := make(chan *scanData)
	// done channel takes the result of the job
	done := make(chan bool)

	fullScanDone := make(chan bool)

	// spawn workers
	for i := 0; i < numberOfWorkers; i++ {
		go s.worker(q, i, done, s.killSignal)
	}

	// wait until scan of all images is done - non blocking. once all done, notify on fullScanDone chan
	go func() {
		for c := 0; c < len(imageIDToScanData); c++ {
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

	// send all scan data on scan data queue, for workers to pick it up.
	for _, data := range imageIDToScanData {
		go func(data *scanData, ks chan bool) {
			select {
			case q <- data:
				atomic.AddUint32(imagesStartedToScan, 1)
			case <-ks:
				log.WithFields(s.logFields).Debugf("Scan process was canceled. imageID=%v, scanUUID=%v", data.imageID, data.scanUUID)
				return
			}
		}(data, s.killSignal)
	}

	// wait for killSignal or fullScanDone
	select {
	case <-s.killSignal:
		log.WithFields(s.logFields).Info("Scan process was canceled")
		s.Lock()
		s.progress.SetStatus(types.ScanAborted)
		s.Unlock()
	case <-fullScanDone:
		log.WithFields(s.logFields).Infof("All jobs has finished")
		s.Lock()
		s.progress.SetStatus(types.DoneScanning)
		s.Unlock()
	}
}

// worker waits for data on the queue, runs a scan job and waits for results from that scan job. Upon completion, done is notified to the caller.
func (s *Scanner) worker(queue chan *scanData, workNumber int, done, ks chan bool) {
	for {
		select {
		case data := <-queue:
			job, err := s.runJob(data)
			if err != nil {
				errMsg := fmt.Errorf("failed to run job: %v", err)
				log.WithFields(s.logFields).Error(errMsg)
				s.Lock()
				data.success = false
				data.scanErr = &types.ScanError{
					ErrMsg:    err.Error(),
					ErrType:   string(types.JobRun),
					ErrSource: types.ScanErrSourceJob,
				}
				data.completed = true
				s.Unlock()
			} else {
				s.waitForResult(data, ks)
			}

			s.deleteJobIfNeeded(job, data.success, data.completed)

			select {
			case done <- true:
			case <-ks:
				log.WithFields(s.logFields).Infof("Image scan was canceled. imageID=%v", data.imageID)
			}
		case <-ks:
			log.WithFields(s.logFields).Debugf("worker #%v halted", workNumber)
			return
		}
	}
}

func (s *Scanner) waitForResult(data *scanData, ks chan bool) {
	log.WithFields(s.logFields).Infof("Waiting for result. imageID=%+v", data.imageID)
	ticker := time.NewTicker(s.scanConfig.JobResultTimeout)
	select {
	case <-data.resultChan:
		log.WithFields(s.logFields).Infof("Image scanned result has arrived. imageID=%v", data.imageID)
	case <-ticker.C:
		errMsg := fmt.Errorf("job has timed out. imageID=%v", data.imageID)
		log.WithFields(s.logFields).Warn(errMsg)
		s.Lock()
		data.success = false
		data.scanErr = &types.ScanError{
			ErrMsg:    errMsg.Error(),
			ErrType:   string(types.JobTimeout),
			ErrSource: types.ScanErrSourceJob,
		}
		data.timeout = true
		data.completed = true
		s.Unlock()
	case <-ks:
		log.WithFields(s.logFields).Infof("Image scan was canceled. imageID=%v", data.imageID)
	}
}

func validateImageID(imageID string) error {
	// image ids with no name and only hash are not pullable from the registry, so we can't scan them.
	if strings.HasPrefix(imageID, localImageIDDockerPrefix) || strings.HasPrefix(imageID, localImageIDSHA256Prefix) {
		return fmt.Errorf("scanning of local docker images is not supported. The Image must be present in the image registry. ImageID=%v", imageID)
	}
	return nil
}

func (s *Scanner) runJob(data *scanData) (*batchv1.Job, error) {
	if err := validateImageID(data.imageID); err != nil {
		return nil, fmt.Errorf("imageID validation failed: %v", err)
	}
	job, err := s.createJob(data)
	if err != nil {
		return nil, fmt.Errorf("failed to create job object. imageID=%v: %v", data.imageID, err)
	}

	log.WithFields(s.logFields).Debugf("Created job=%+v", job)

	log.WithFields(s.logFields).Infof("Running job %s/%s to scan imageID %s", job.GetNamespace(), job.GetName(), data.imageID)
	_, err = s.clientset.BatchV1().Jobs(job.GetNamespace()).Create(context.TODO(), job, metav1.CreateOptions{})
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
	case config.DeleteJobPolicyNever:
		// do nothing
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
	deleteOptions := metav1.DeleteOptions{PropagationPolicy: &dpb}

	log.WithFields(s.logFields).Infof("Deleting job %s/%s", job.GetNamespace(), job.GetName())
	err := s.clientset.BatchV1().Jobs(job.GetNamespace()).Delete(context.TODO(), job.GetName(), deleteOptions)
	if err != nil && !errors.IsNotFound(err) {
		log.WithFields(s.logFields).Errorf("failed to delete job: %s/%s. %v", job.GetNamespace(), job.GetName(), err)
	}
}

// Due to K8s names constraint we will take the image name w/o the tag and repo.
func getSimpleImageName(imageName string) (string, error) {
	ref, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		return "", fmt.Errorf("failed to parse image name. name=%v: %v", imageName, err)
	}

	refName := ref.Name()
	// Take only image name from repo path (ex. solsson/kafka ==> kafka)
	repoEnd := strings.LastIndex(refName, "/")

	return refName[repoEnd+1:], nil
}

// Job names require their names to follow the DNS label standard as defined in RFC 1123
// Note: job name is added as a label to the pod template spec so it should follow the DNS label standard and not just DNS-1123 subdomain
//
// This means the name must:
// * contain at most 63 characters
// * contain only lowercase alphanumeric characters or ‘-’
// * start with an alphanumeric character
// * end with an alphanumeric character.
func createJobName(imageName string) (string, error) {
	simpleName, err := getSimpleImageName(imageName)
	if err != nil {
		return "", err
	}

	jobName := "scanner-" + simpleName + "-" + uuid.NewV4().String()

	// contain at most 63 characters
	jobName = stringsutils.TruncateString(jobName, k8s.MaxK8sJobName)

	// contain only lowercase alphanumeric characters or ‘-’
	jobName = strings.ToLower(jobName)
	jobName = strings.ReplaceAll(jobName, "_", "-")

	// no need to validate start, we are using 'jobName'

	// end with an alphanumeric character
	jobName = strings.TrimRight(jobName, "-")

	return jobName, nil
}

func (s *Scanner) createJob(data *scanData) (*batchv1.Job, error) {
	// We will scan each image once, based on the first pod context. The result will be applied for all other pods with this image.
	podContext := data.contexts[0]

	jobName, err := createJobName(podContext.imageName)
	if err != nil {
		return nil, fmt.Errorf("failed to create job name. namespace=%v, pod=%v, container=%v, image=%v, hash=%v: %v",
			podContext.namespace, podContext.podName, podContext.containerName, podContext.imageName, data.imageHash, err)
	}

	// Set job values on scanner job template
	job := s.scannerJobTemplate.DeepCopy()
	if !data.shouldScanCISDockerBenchmark {
		removeCISDockerBenchmarkScannerFromJob(job)
	}
	job.SetName(jobName)
	job.SetNamespace(podContext.namespace)
	setJobScanUUID(job, data.scanUUID)
	setJobImageIDToScan(job, data.imageID)
	setJobImageHashToScan(job, data.imageHash)
	setJobImageNameToScan(job, podContext.imageName)

	if len(podContext.imagePullSecrets) > 0 {
		for _, secretName := range podContext.imagePullSecrets {
			addJobImagePullSecretVolume(job, secretName)
		}
		setJobImagePullSecretPath(job)
	} else {
		// Use private repo sa credentials only if there is no imagePullSecret
		for _, adder := range s.credentialAdders {
			if adder.ShouldAdd() {
				adder.Add(job)
			}
		}
	}

	return job, nil
}

func removeCISDockerBenchmarkScannerFromJob(job *batchv1.Job) {
	var containers []corev1.Container
	for i := range job.Spec.Template.Spec.Containers {
		container := job.Spec.Template.Spec.Containers[i]
		if container.Name != cisDockerBenchmarkScannerContainerName {
			containers = append(containers, container)
		}
	}
	job.Spec.Template.Spec.Containers = containers
}

const (
	imagePullSecretMountPath    = "/opt/kubeclarity-pull-secrets" // nolint:gosec
	imagePullSecretVolumePrefix = "image-pull-secret-"            // nolint:gosec
)

// Mount image pull secret as a volume into /opt/kubeclarity-pull-secrets so
// that the scanner job can find it. setJobImagePullSecretPath must be used in
// addition to this function to configure IMAGE_PULL_SECRET_PATH environment
// variable.
//  1. Create a volume "image-pull-secret-secretName" that holds the
//     `secretName` data. We don't know if this secret exists so mark it
//     optional so it doesn't block the pod starting.
//  2. Mount the volume into each container to a specific path
//     /opt/kubeclarity-pull-secrets/secretName
func addJobImagePullSecretVolume(job *batchv1.Job, secretName string) {
	volumeName := fmt.Sprintf("%s%s", imagePullSecretVolumePrefix, secretName)
	optional := true
	job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
				Optional:   &optional,
			},
		},
	})
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: path.Join(imagePullSecretMountPath, secretName),
		})
	}
}

// Set the IMAGE_PULL_SECRET_PATH environment variable to
// /opt/kubeclarity-pull-secrets.
func setJobImagePullSecretPath(job *batchv1.Job) {
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]

		// Try to find and update any existing IMAGE_PULL_SECRET_PATH env var.
		for i, eVar := range container.Env {
			if eVar.Name == shared.ImagePullSecretPath {
				container.Env[i].Value = imagePullSecretMountPath
				return
			}
		}

		// We didn't find an existing env var, so add a new one.
		container.Env = append(container.Env, corev1.EnvVar{Name: shared.ImagePullSecretPath, Value: imagePullSecretMountPath})
	}
}

func setJobImageIDToScan(job *batchv1.Job, imageID string) {
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{Name: shared.ImageIDToScan, Value: imageID})
	}
}

func setJobImageHashToScan(job *batchv1.Job, imageHash string) {
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{Name: shared.ImageHashToScan, Value: imageHash})
	}
}

func setJobImageNameToScan(job *batchv1.Job, imageName string) {
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{Name: shared.ImageNameToScan, Value: imageName})
	}
}

func setJobScanUUID(job *batchv1.Job, scanUUID string) {
	for i := range job.Spec.Template.Spec.Containers {
		container := &job.Spec.Template.Spec.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{Name: shared.ScanUUID, Value: scanUUID})
	}
}
