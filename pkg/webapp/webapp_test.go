package webapp

import (
	"reflect"
	"testing"
)

func Test_sortDockerfileVulnerabilities(t *testing.T) {
	type args struct {
		data []*containerDockerfileVulnerability
	}
	tests := []struct {
		name string
		args args
		want []*containerDockerfileVulnerability
	}{
		{
			name: "sort fatal before warning",
			args: args{
				data: []*containerDockerfileVulnerability{
					{
						containerInfo: containerInfo{
							Pod:       "warn pod",
						},
						DockerfileVulnerability: &dockerfileVulnerability{
							Level:       "WARN",
						},
					},
					{
						containerInfo: containerInfo{
							Pod:       "fatal pod",
						},
						DockerfileVulnerability: &dockerfileVulnerability{
							Level:       "FATAL",
						},
					},
				},
			},
			want: []*containerDockerfileVulnerability{
				{
					containerInfo: containerInfo{
						Pod:       "fatal pod",
					},
					DockerfileVulnerability: &dockerfileVulnerability{
						Level:       "FATAL",
					},
				},
				{
					containerInfo: containerInfo{
						Pod:       "warn pod",
					},
					DockerfileVulnerability: &dockerfileVulnerability{
						Level:       "WARN",
					},
				},
			},
		},
		{
			name: "no need to sort",
			args: args{
				data: []*containerDockerfileVulnerability{
					{
						containerInfo: containerInfo{
							Pod:       "fatal pod",
						},
						DockerfileVulnerability: &dockerfileVulnerability{
							Level:       "FATAL",
						},
					},
					{
						containerInfo: containerInfo{
							Pod:       "warn pod",
						},
						DockerfileVulnerability: &dockerfileVulnerability{
							Level:       "WARN",
						},
					},
				},
			},
			want: []*containerDockerfileVulnerability{
				{
					containerInfo: containerInfo{
						Pod:       "fatal pod",
					},
					DockerfileVulnerability: &dockerfileVulnerability{
						Level:       "FATAL",
					},
				},
				{
					containerInfo: containerInfo{
						Pod:       "warn pod",
					},
					DockerfileVulnerability: &dockerfileVulnerability{
						Level:       "WARN",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortDockerfileVulnerabilities(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortDockerfileVulnerabilities() = %v, want %v", got, tt.want)
			}
		})
	}
}
