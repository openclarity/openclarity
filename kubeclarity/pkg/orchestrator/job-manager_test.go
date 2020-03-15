package orchestrator

import "testing"

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
				t.Errorf("createJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}