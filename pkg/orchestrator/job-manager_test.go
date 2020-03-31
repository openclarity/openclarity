package orchestrator

import (
	"k8s.io/apimachinery/pkg/util/validation"
	"testing"
)

func Test_getSimpleImageName(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "",
			args: args{
				imageName: "docker.io/nginx:1.10",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSimpleImageName(tt.args.imageName); got != tt.want {
				t.Errorf("getSimpleImageName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createJobName(t *testing.T) {
	type args struct {
		imageName string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "trim right '-' that left from the uuid after name was truncated due to max len",
			args: args{
				imageName: "stackdriver-logging-agent",
			},
		},
		{
			name: "lower case",
			args: args{
				imageName: "LowerCase",
			},
		},
		{
			name: "underscore",
			args: args{
				imageName: "under_score",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createJobName(tt.args.imageName)
			errs := validation.IsDNS1123Label(got)
			if len(errs) != 0 {
				t.Errorf("createJobName() = name is not valid. got=%v, errs=%+v", got, errs)
			}
		})
	}
}