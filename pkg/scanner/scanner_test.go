package scanner

import (
	v1 "k8s.io/api/core/v1"
	"testing"
)

func Test_getImageHash(t *testing.T) {
	var (
		validImageHash  = "6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa"
		validImageId    = "docker-pullable://gcr.io/development-infra-208909/kubei@sha256:" + validImageHash
		notValidImageId = "notValidImageId"
	)
	type args struct {
		containerNameToImageId map[string]string
		container              v1.Container
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "valid image hash",
			args: args{
				containerNameToImageId: map[string]string{"container-name": validImageId},
				container: v1.Container{
					Name:  "container-name",
					Image: "image-name",
				},
			},
			want: validImageHash,
		},
		{
			name: "image id is missing",
			args: args{
				containerNameToImageId: nil,
				container: v1.Container{
					Name:  "container-name",
					Image: "image-name",
				},
			},
			want: "",
		},
		{
			name: "failed to parse image hash",
			args: args{
				containerNameToImageId: map[string]string{"container-name": notValidImageId},
				container: v1.Container{
					Name:  "container-name",
					Image: "image-name",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getImageHash(tt.args.containerNameToImageId, tt.args.container); got != tt.want {
				t.Errorf("getImageHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
