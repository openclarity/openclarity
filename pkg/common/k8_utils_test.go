package common

import (
	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestK8ContextService_GetK8ContextFromContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockK8ContextSecretService := NewMockK8ContextSecretServiceInterface(mockCtrl)
	mockK8ContextSecretService.EXPECT().GetMatchingSecretName(gomock.Any(), gomock.Any()).DoAndReturn(func(secrets []corev1.Secret, container corev1.Container) string {
		if container.Image == "image3" {
			return "secret 2"
		} else {
			return "secret 1"
		}
	}).AnyTimes()

	type args struct {
		orchestratorImageK8ExtendedContextMap ImageK8ExtendedContextMap
		pod                                   *corev1.Pod
		imageNamespacesMap                    ImageNamespacesMap
		namespacedImageSecretMap              NamespacedImageSecretMap
		containerImagesSet                    map[ContainerImageName]bool
		totalContainers                       int
	}

	testPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod 1",
			Namespace: "ns1",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "container1",
					Image: "image1",
				},
				{
					Name:  "container2",
					Image: "image2",
				},
				{
					Name:  "container3",
					Image: "image2",
				},
				{
					Name:  "container3",
					Image: "image3",
				},
			},
		},
		Status: corev1.PodStatus{},
	}
	testInitialArgs := args{
		orchestratorImageK8ExtendedContextMap: make(ImageK8ExtendedContextMap),
		pod:                                   testPod,
		imageNamespacesMap:                    make(ImageNamespacesMap),
		namespacedImageSecretMap:              make(NamespacedImageSecretMap),
		containerImagesSet:                    make(map[ContainerImageName]bool),
		totalContainers:                       0,
	}






	tests := []struct {
		name                     string
		args                     args
		imageNamespacesMap       ImageNamespacesMap
		namespacedImageSecretMap NamespacedImageSecretMap
		containerImagesSet       map[ContainerImageName]bool
		totalContainers          int
	}{
		{
			name: "test image namespace map",
			args:                     testInitialArgs,
			imageNamespacesMap:       ImageNamespacesMap{"ns1": []ContainerImageName{"image1", "image2","image3"}},
			namespacedImageSecretMap: nil,
			containerImagesSet:       nil,
			totalContainers:          -1,
		},
		{
			name: "test namespaced image secret map",
			args:                     testInitialArgs,
			imageNamespacesMap:       nil,
			namespacedImageSecretMap: NamespacedImageSecretMap{"image1_ns1": "secret 1", "image2_ns1": "secret 1", "image3_ns1": "secret 2"},
			containerImagesSet:       nil,
			totalContainers:          -1,
		},
		{
			name: "test container images set (scan each image once)",
			args:                     testInitialArgs,
			imageNamespacesMap:       nil,
			namespacedImageSecretMap: nil,
			containerImagesSet:       map[ContainerImageName]bool{"image1": true, "image2": true, "image3": true},
			totalContainers:          -1,
		},
		{
			name: "test containers count (count container even if same image)",
			args:                     testInitialArgs,
			imageNamespacesMap:       nil,
			namespacedImageSecretMap: nil,
			containerImagesSet:       nil,
			totalContainers:          4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kcs := &K8ContextService{
				K8ContextSecretService: mockK8ContextSecretService,
			}
			got, got1, got2, got3 := kcs.GetK8ContextFromContainer(tt.args.orchestratorImageK8ExtendedContextMap, tt.args.pod, tt.args.imageNamespacesMap, tt.args.namespacedImageSecretMap, tt.args.containerImagesSet, tt.args.totalContainers)
			if tt.imageNamespacesMap!= nil && !reflect.DeepEqual(got, tt.imageNamespacesMap) {
				t.Errorf("GetK8ContextFromContainer() got = %v, imageNamespacesMap %v", got, tt.imageNamespacesMap)
			}
			if  tt.namespacedImageSecretMap!= nil && !reflect.DeepEqual(got1, tt.namespacedImageSecretMap) {
				t.Errorf("GetK8ContextFromContainer() got1 = %v, namespacedImageSecretMap %v", got1, tt.namespacedImageSecretMap)
			}
			if  tt.containerImagesSet!= nil && !reflect.DeepEqual(got2, tt.containerImagesSet) {
				t.Errorf("GetK8ContextFromContainer() got2 = %v, containerImagesSet %v", got2, tt.containerImagesSet)
			}
			if  tt.totalContainers!= -1  && got3 != tt.totalContainers {
				t.Errorf("GetK8ContextFromContainer() got3 = %v, totalContainers %v", got3, tt.totalContainers)
			}
		})
	}
}
