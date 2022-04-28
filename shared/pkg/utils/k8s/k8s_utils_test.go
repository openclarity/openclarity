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

package k8s

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseImageHash(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid image id",
			args: args{
				imageID: "docker-pullable://gcr.io/development-infra-208909/kubei@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "no image hash",
			args: args{
				imageID: "docker-pullable://gcr.io/development-infra-208909/kubei@sha256:",
			},
			want: "",
		},
		{
			name: "no image id",
			args: args{
				imageID: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseImageHash(tt.args.imageID); got != tt.want {
				t.Errorf("ParseImageHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMatchingSecretName(t *testing.T) {
	notSpecificImageName := "gcr.io/not/specific"
	specificImageName := "gcr.io/more/specific"
	partialPathsImageName := "gcr.io/partial/path"
	dockerHubImageName := "repo/image"

	dockerconfigjsonB := []byte("{\"auths\":{\"gcr.io\":{\"username\":\"gcr\",\"password\":\"io\",\"auth\":\"Z2NyOmlv\"},\"gcr.io/more/specific\":{\"username\":\"gcr\",\"password\":\"io/more/specific\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}")
	partialPathsDockerconfigjsonB := []byte("{\"auths\":{\"gcr.io/partial\":{\"username\":\"gcr\",\"password\":\"partial\",\"auth\":\"Z2NyOmlvL21vcmUvc3BlY2lmaWM=\"},\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}")
	noMatchDockerconfigjsonB := []byte("{\"auths\":{\"http://foo.example.com\":{\"username\":\"foo\",\"password\":\"bar\",\"auth\":\"Zm9vOmJhcg==\"}}}")
	matchDockerIODockerconfigjsonB := []byte("{\"auths\":{\"https://index.docker.io/v1/\":{\"username\":\"test-user\",\"password\":\"test-pass)\",\"email\":\"test@test\",\"auth\":\"dGVzdC11c2VyOnRlc3QtcGFzcw==\"}}}")

	type args struct {
		secrets   []*corev1.Secret
		imageName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "match specific image name",
			args: args{
				secrets: []*corev1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "regcred",
							Namespace: "default",
						},
						Data: map[string][]byte{
							corev1.DockerConfigJsonKey: dockerconfigjsonB,
						},
						Type: corev1.SecretTypeDockerConfigJson,
					},
				},
				imageName: specificImageName,
			},
			want: "regcred",
		},
		{
			name: "match partial paths image name",
			args: args{
				secrets: []*corev1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "regcred",
							Namespace: "default",
						},
						Data: map[string][]byte{
							corev1.DockerConfigJsonKey: partialPathsDockerconfigjsonB,
						},
						Type: corev1.SecretTypeDockerConfigJson,
					},
				},
				imageName: partialPathsImageName,
			},
			want: "regcred",
		},
		{
			name: "match not specific image name",
			args: args{
				secrets: []*corev1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "no-match",
							Namespace: "default",
						},
						Data: map[string][]byte{
							corev1.DockerConfigJsonKey: noMatchDockerconfigjsonB,
						},
						Type: corev1.SecretTypeDockerConfigJson,
					},
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "regcred",
							Namespace: "default",
						},
						Data: map[string][]byte{
							corev1.DockerConfigJsonKey: dockerconfigjsonB,
						},
						Type: corev1.SecretTypeDockerConfigJson,
					},
				},
				imageName: notSpecificImageName,
			},
			want: "regcred",
		},
		{
			name: "no match",
			args: args{
				secrets: []*corev1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "no-match",
							Namespace: "default",
						},
						Data: map[string][]byte{
							corev1.DockerConfigJsonKey: noMatchDockerconfigjsonB,
						},
						Type: corev1.SecretTypeDockerConfigJson,
					},
				},
				imageName: notSpecificImageName,
			},
			want: "",
		},
		{
			name: "not docker config json secret",
			args: args{
				secrets: []*corev1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "no-docker-config",
							Namespace: "default",
						},
						Data: map[string][]byte{
							"bla": []byte("bla"),
						},
					},
				},
				imageName: notSpecificImageName,
			},
			want: "",
		},
		{
			name: "match docker hub image",
			args: args{
				secrets: []*corev1.Secret{
					{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: "v1",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "regcred",
							Namespace: "default",
						},
						Data: map[string][]byte{
							corev1.DockerConfigJsonKey: matchDockerIODockerconfigjsonB,
						},
						Type: corev1.SecretTypeDockerConfigJson,
					},
				},
				imageName: dockerHubImageName,
			},
			want: "regcred",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMatchingSecretName(tt.args.secrets, tt.args.imageName); got != tt.want {
				t.Errorf("GetMatchingSecretName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseImageID(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "image id with docker-pullable prefix",
			args: args{
				imageID: "docker-pullable://gcr.io/development-infra-208909/kubei@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "gcr.io/development-infra-208909/kubei@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "image id without docker-pullable prefix",
			args: args{
				imageID: "gcr.io/development-infra-208909/kubei@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
			},
			want: "gcr.io/development-infra-208909/kubei@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa",
		},
		{
			name: "no image id",
			args: args{
				imageID: "",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseImageID(tt.args.imageID); got != tt.want {
				t.Errorf("ParseImageID() = %v, want %v", got, tt.want)
			}
		})
	}
}
