package k8s

import "testing"

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
