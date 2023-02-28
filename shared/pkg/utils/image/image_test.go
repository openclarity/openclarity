package image

import "testing"

func TestValidateLocalImageID(t *testing.T) {
	type args struct {
		imageID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "docker://sha256 prefix",
			args: args{
				imageID: "docker://sha256:12bae74413f7240099ba68a4b44c55541fa94c51c676681c2988a7571e6891eb",
			},
			wantErr: true,
		},
		{
			name: "sha256: prefix",
			args: args{
				imageID: "sha256:12bae74413f7240099ba68a4b44c55541fa94c51c676681c2988a7571e6891eb",
			},
			wantErr: true,
		},
		{
			name: "good",
			args: args{
				imageID: "gke.gcr.io/proxy-agent@sha256:d5ae8affd1ca510a4bfd808e14a563c573510a70196ad5b04fdf0fb5425abf35",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateLocalImageID(tt.args.imageID); (err != nil) != tt.wantErr {
				t.Errorf("ValidateLocalImageID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
