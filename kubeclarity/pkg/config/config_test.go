package config

import "testing"

func Test_getServiceAddressAndGetServiceFullPath(t *testing.T) {
	type args struct {
		serviceFullPath string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "full service path",
			args: args{
				serviceFullPath: "http://kubei.kubei:8081/result/",
			},
			want: "kubei.kubei",
		},
		{
			name: "w/o scheme",
			args: args{
				serviceFullPath: "kubei.kubei:8081/result/",
			},
			want: "kubei.kubei",
		},
		{
			name: "w/o path",
			args: args{
				serviceFullPath: "http://kubei.kubei:8081",
			},
			want: "kubei.kubei",
		},
		{
			name: "w/o port",
			args: args{
				serviceFullPath: "http://kubei.kubei/result/",
			},
			want: "kubei.kubei",
		},
		{
			name: "w/o port and path",
			args: args{
				serviceFullPath: "http://kubei.kubei",
			},
			want: "kubei.kubei",
		},
		{
			name: "w/o scheme, port and path",
			args: args{
				serviceFullPath: "kubei.kubei",
			},
			want: "kubei.kubei",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getServiceAddress(getServiceFullPath(tt.args.serviceFullPath)); got != tt.want {
				t.Errorf("getServiceAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
