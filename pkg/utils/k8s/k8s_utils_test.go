package k8s

import (
	"encoding/json"
	"testing"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/credentialprovider"
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
	dockerconfigjson := credentialprovider.DockerConfigJson{
		Auths: credentialprovider.DockerConfig(map[string]credentialprovider.DockerConfigEntry{
			"http://foo.example.com": {
				Username: "foo",
				Password: "bar",
			},
			"gcr.io": {
				Username: "gcr",
				Password: "io",
			},
			specificImageName: {
				Username: "gcr",
				Password: "io/more/specific",
			},
		}),
	}
	dockerconfigjsonB, err := json.Marshal(dockerconfigjson)
	assert.NilError(t, err)

	noMatchDockerconfigjson := credentialprovider.DockerConfigJson{
		Auths: credentialprovider.DockerConfig(map[string]credentialprovider.DockerConfigEntry{
			"http://foo.example.com": {
				Username: "foo",
				Password: "bar",
			},
		}),
	}
	noMatchDockerconfigjsonB, err := json.Marshal(noMatchDockerconfigjson)
	assert.NilError(t, err)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMatchingSecretName(tt.args.secrets, tt.args.imageName); got != tt.want {
				t.Errorf("GetMatchingSecretName() = %v, want %v", got, tt.want)
			}
		})
	}
}
