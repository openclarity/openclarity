package scanner

import (
	"k8s.io/apimachinery/pkg/util/validation"
	"testing"
)

func Test_getSimpleImageName(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid image name with tag and repo",
			args: args{
				imageName: "docker.io/nginx:1.10",
			},
			want: "nginx",
		},
		{
			name: "valid image name with digest with repo",
			args: args{
				imageName: "docker.io/nginx@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2",
			},
			want: "nginx",
		},
		{
			name: "valid image name with digest no repo",
			args: args{
				imageName: "nginx@sha256:45b23dee08af5e43a7fea6c4cf9c25ccf269ee113168c19722f87876677c5cb2",
			},
			want: "nginx",
		},
		{
			name: "no tag",
			args: args{
				imageName: "docker.io/nginx",
			},
			want: "nginx",
		},
		{
			name: "no tag with port",
			args: args{
				imageName: "docker.io:8080/nginx",
			},
			want: "nginx",
		},
		{
			name: "repo with port",
			args: args{
				imageName: "docker.io:8080/nginx:1.10",
			},
			want: "nginx",
		},
		{
			name: "no repo no tag",
			args: args{
				imageName: "nginx",
			},
			want: "nginx",
		},
		{
			name: "valid image name with digest with repo with tag",
			args: args{
				imageName: "solsson/kafka:2.2.1@sha256:450c6fdacae3f89ca28cecb36b2f120aad9b19583d68c411d551502ee8d0b09b",
			},
			want: "kafka",
		},
		{
			name: "name ends with '/' - invalid reference format",
			args: args{
				imageName: "docker.io:8080/not/valid/:222",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSimpleImageName(tt.args.imageName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSimpleImageName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getSimpleImageName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createJobName(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "trim right '-' that left from the uuid after name was truncated due to max len",
			args: args{
				imageName: "stackdriver-logging-agent",
			},
		},
		{
			name: "underscore",
			args: args{
				imageName: "under_score",
			},
		},
		{
			name: "invalid image name",
			args: args{
				imageName: "InvAliD",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createJobName(tt.args.imageName)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("createJobName() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			errs := validation.IsDNS1123Label(got)
			if len(errs) != 0 {
				t.Errorf("createJobName() = name is not valid. got=%v, errs=%+v", got, errs)
			}
		})
	}
}
