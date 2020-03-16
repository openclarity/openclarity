package orchestrator

import (
	"gotest.tools/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubei/common"
	"strings"
	"sync"
	"testing"
)

func TestOrchestrator_buildContainersPart(t *testing.T) {
	envVars := []corev1.EnvVar{
		{Name: "CLAIR_ADDR", Value: "clairsvc.kubei"},
		{Name: "CLAIR_OUTPUT", Value: "HIGH"},
		{Name: "KLAR_TRACE", Value: "false"},
		{Name: "WHITELIST_FILE", Value: ""},
	}

	type fields struct {
		ImageK8ExtendedContextMap common.ImageK8ExtendedContextMap
		DataUpdateLock            *sync.Mutex
		ExecutionConfig           *common.ExecutionConfiguration
		k8ContextService          *common.K8ContextService
		scanIssuesMessages        *[]string
		batchCompletedScansCount  *int32
		k8ContextServiceInterface *common.K8ContextServiceInterface
	}
	type args struct {
		imageNamespace           string
		batch                    []string
		startPoint               int
		scannedImageNames        []string
		namespacedImageSecretMap common.NamespacedImageSecretMap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []corev1.Container
	}{
		{
			name: "test create containers part",
			fields: fields{
				ImageK8ExtendedContextMap: common.ImageK8ExtendedContextMap{
					common.ContainerImageName("image1"): []*common.K8ExtendedContext{
						{
							Namespace: "ns1",
							Container: "container1",
							Pod:       "pod1",
							Secret:    "secret1",
						},
					},
					common.ContainerImageName("image2"): []*common.K8ExtendedContext{
						{
							Namespace: "ns1",
							Container: "container2",
							Pod:       "pod1",
							Secret:    "secret2",
						},
					},
					common.ContainerImageName("image3"): []*common.K8ExtendedContext{
						{
							Namespace: "ns1",
							Container: "container3",
							Pod:       "pod2",
							Secret:    "secret3",
						},
					},
					common.ContainerImageName("image4"): []*common.K8ExtendedContext{
						{
							Namespace: "ns2",
							Container: "container3",
							Pod:       "pod2",
							Secret:    "secret3",
						},
					},
				},
				ExecutionConfig: &common.ExecutionConfiguration{
					Clientset:        nil,
					Parallelism:      0,
					KubeiNamespace:   "kubei",
					TargetNamespace:  "ns1",
					ClairOutput:      "HIGH",
					WhitelistFile:    "",
					IgnoreNamespaces: nil,
					KlarTrace:        false,
				},
			},
			args: args{
				imageNamespace: "ns1",
				batch:          []string{"image1", "image2", "image3"},
				startPoint:     0,
			},
			want: []corev1.Container{
				{
					Name:  "container1",
					Image: "rafiportshift/portshift-klar:1.0.0",
					Env:   envVars,
				},
				{
					Name:  "container2",
					Image: "rafiportshift/portshift-klar:1.0.0",
					Env:   envVars,
				},
				{
					Name:  "container3",
					Image: "rafiportshift/portshift-klar:1.0.0",
					Env:   envVars,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orc := &Orchestrator{
				ImageK8ExtendedContextMap: tt.fields.ImageK8ExtendedContextMap,
				DataUpdateLock:            tt.fields.DataUpdateLock,
				ExecutionConfig:           tt.fields.ExecutionConfig,
				k8ContextService:          tt.fields.k8ContextService,
				scanIssuesMessages:        tt.fields.scanIssuesMessages,
				batchCompletedScansCount:  tt.fields.batchCompletedScansCount,
			}
			got := orc.buildContainersPart(tt.args.imageNamespace, tt.args.batch, tt.args.startPoint, tt.args.scannedImageNames, tt.args.namespacedImageSecretMap)
			assert.Assert(t, strings.HasPrefix(got[0].Name, "klar-image1-"))
			assert.Equal(t, got[0].Image, tt.want[0].Image)
			assert.DeepEqual(t, got[0].Env, tt.want[0].Env)

			assert.Assert(t, strings.HasPrefix(got[1].Name, "klar-image2-"))
			assert.Equal(t, got[1].Image, tt.want[1].Image)
			assert.DeepEqual(t, got[1].Env, tt.want[1].Env)
		})
	}
}

func TestOrchestrator_createJobDefinition(t *testing.T) {
	exampleLabels := map[string]string{"key1": "value1"}
	exampleBackoffLimit := int32(88)
	exampleTtlSecondsAfterFinished := int32(99)
	exampleContainers := []corev1.Container{
		{
			Name:  "some container name",
			Image: "some image ",
			Ports: []corev1.ContainerPort{{ //different from default
				Name:          "port 1",
				HostPort:      44,
				ContainerPort: 55,
				Protocol:      "TCP",
				HostIP:        "1.1.1.1",
			},
				{
					Name:          "port 2",
					HostPort:      13221,
					ContainerPort: 66,
					Protocol:      "HTTP",
					HostIP:        "2.2.2.2",
				},
			},
			StdinOnce: true, //different from default
			TTY:       true, //different from default
		},
	}

	type fields struct {
		ImageK8ExtendedContextMap common.ImageK8ExtendedContextMap
		DataUpdateLock            *sync.Mutex
		ExecutionConfig           *common.ExecutionConfiguration
		k8ContextService          *common.K8ContextService
		scanIssuesMessages        *[]string
		batchCompletedScansCount  *int32
		k8ContextServiceInterface *common.K8ContextServiceInterface
	}
	type args struct {
		jobName                 string
		imageNamespace          string
		labels                  map[string]string
		containers              []corev1.Container
		backOffLimit            int32
		ttlSecondsAfterFinished int32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *batchv1.Job
	}{
		{
			name:   "test create job definition",
			fields: fields{},
			args: args{
				jobName:                 "some jobname",
				imageNamespace:          "some imageNamespace",
				labels:                  exampleLabels,
				containers:              exampleContainers,
				backOffLimit:            exampleBackoffLimit,            //different from default
				ttlSecondsAfterFinished: exampleTtlSecondsAfterFinished, //different from default
			},
			want: &batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "some jobname",
					Namespace: "some imageNamespace",
					Labels:    exampleLabels,
				},
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: v1.ObjectMeta{
							Labels: exampleLabels,
						},
						Spec: corev1.PodSpec{
							ServiceAccountName: "kubei",
							Containers:         exampleContainers,
							RestartPolicy:      corev1.RestartPolicyNever,
						},
					},
					TTLSecondsAfterFinished: &exampleTtlSecondsAfterFinished,
					BackoffLimit:            &exampleBackoffLimit,
				},
				Status: batchv1.JobStatus{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orc := &Orchestrator{
				ImageK8ExtendedContextMap: tt.fields.ImageK8ExtendedContextMap,
				DataUpdateLock:            tt.fields.DataUpdateLock,
				ExecutionConfig:           tt.fields.ExecutionConfig,
				k8ContextService:          tt.fields.k8ContextService,
				scanIssuesMessages:        tt.fields.scanIssuesMessages,
				batchCompletedScansCount:  tt.fields.batchCompletedScansCount,
			}

			got := orc.createJobDefinition(tt.args.jobName, tt.args.imageNamespace, tt.args.labels, tt.args.containers, tt.args.backOffLimit, tt.args.ttlSecondsAfterFinished)
			assert.Equal(t, got.Name, tt.want.Name)
			assert.Equal(t, got.Namespace, tt.want.Namespace)
			assert.DeepEqual(t, got.TypeMeta, tt.want.TypeMeta)
			assert.DeepEqual(t, got.ObjectMeta, tt.want.ObjectMeta)
			assert.DeepEqual(t, got.Spec, tt.want.Spec)
			assert.DeepEqual(t, got.Status, tt.want.Status)
		})
	}
}
