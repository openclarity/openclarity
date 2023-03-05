package image

import "testing"

func TestIsLocalImage(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "docker://sha256 prefix",
			args: args{
				imageID: "docker://sha256:12bae74413f7240099ba68a4b44c55541fa94c51c676681c2988a7571e6891eb",
			},
			want: true,
		},
		{
			name: "sha256: prefix",
			args: args{
				imageID: "sha256:12bae74413f7240099ba68a4b44c55541fa94c51c676681c2988a7571e6891eb",
			},
			want: true,
		},
		{
			name: "good",
			args: args{
				imageID: "gke.gcr.io/proxy-agent@sha256:d5ae8affd1ca510a4bfd808e14a563c573510a70196ad5b04fdf0fb5425abf35",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLocalImage(tt.args.imageID); got != tt.want {
				t.Errorf("IsLocalImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
