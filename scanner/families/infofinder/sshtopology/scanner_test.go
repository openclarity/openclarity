// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sshtopology

import (
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/families/infofinder/types"
)

var testScanner = &Scanner{
	logger: log.NewEntry(log.StandardLogger()),
}

func Test_getHomeUserDirs(t *testing.T) {
	type args struct {
		rootDir string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "sanity",
			args: args{
				rootDir: "../testdata/rootfolder",
			},
			want: []string{"../testdata/rootfolder/home/dir1", "../testdata/rootfolder/home/dir2"},
		},
		{
			name: "root folder with root home folder",
			args: args{
				rootDir: "../testdata/rootfolderwithroothome",
			},
			want: []string{"../testdata/rootfolderwithroothome/root", "../testdata/rootfolderwithroothome/home/dir1", "../testdata/rootfolderwithroothome/home/dir2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getHomeUserDirs(tt.args.rootDir)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHomeUserDirs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_getPrivateKeysPaths(t *testing.T) {
	type args struct {
		rootPath  string
		recursive bool
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "not recursive - found nothing",
			args: args{
				rootPath:  "../testdata/rootfolder",
				recursive: false,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "not recursive - found keys in folder",
			args: args{
				rootPath:  "../testdata/rootfolder/home/dir1",
				recursive: false,
			},
			want:    []string{"../testdata/rootfolder/home/dir1/private_key"},
			wantErr: false,
		},
		{
			name: "recursive - found keys in all sub folders",
			args: args{
				rootPath:  "../testdata/rootfolder",
				recursive: true,
			},
			want: []string{
				"../testdata/rootfolder/.ssh/private_key",
				"../testdata/rootfolder/etc/ssh/ssh_dummy_key",
				"../testdata/rootfolder/etc/ssh/ssh_dummy_key2",
				"../testdata/rootfolder/home/dir1/dir3/private_key",
				"../testdata/rootfolder/home/dir1/private_key",
				"../testdata/rootfolder/home/dir2/private_key",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := testScanner
			got, err := s.getPrivateKeysPaths(tt.args.rootPath, tt.args.recursive)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPrivateKeysPaths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getPrivateKeysPaths() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPrivateKey(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "private key",
			args: args{
				path: "../testdata/rootfolder/.ssh/private_key",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not private key",
			args: args{
				path: "../testdata/rootfolder/.ssh/not_a_key",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "empty_file",
			args: args{
				path: "../testdata/rootfolder/.ssh/empty_file",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "folder as an input should return an error",
			args: args{
				path: "../testdata/rootfolder/.ssh",
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "path does not exists should return an error",
			args: args{
				path: "../testdata/dummyrootfolder/.ssh",
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isPrivateKey(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("isPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseSSHKeyGenCommandOutput(t *testing.T) {
	type args struct {
		output   string
		infoType types.InfoType
		path     string
	}
	tests := []struct {
		name string
		args args
		want []types.Info
	}{
		{
			name: "single line output",
			args: args{
				output:   "256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ED25519)",
				infoType: types.SSHKnownHostFingerprint,
				path:     "/home/user/.ssh/known_hosts",
			},
			want: []types.Info{
				{
					Type: types.SSHKnownHostFingerprint,
					Path: "/home/user/.ssh/known_hosts",
					Data: "256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ED25519)",
				},
			},
		},
		{
			name: "single line output with new line",
			args: args{
				output:   "256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ED25519)\n",
				infoType: types.SSHKnownHostFingerprint,
				path:     "/home/user/.ssh/known_hosts",
			},
			want: []types.Info{
				{
					Type: types.SSHKnownHostFingerprint,
					Path: "/home/user/.ssh/known_hosts",
					Data: "256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ED25519)",
				},
			},
		},
		{
			name: "multiple lines output",
			args: args{
				output:   "256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ED25519)\n256 SHA256:cDmm4+e/BNwQpsk/Qhh39i2qiT6HcIs6qTLtIiMWzPg ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ECDSA)",
				infoType: types.SSHKnownHostFingerprint,
				path:     "/home/user/.ssh/known_hosts",
			},
			want: []types.Info{
				{
					Type: types.SSHKnownHostFingerprint,
					Path: "/home/user/.ssh/known_hosts",
					Data: "256 SHA256:gv6snCwAl5+6fY2g5VkmETWb9Mv0zLRkMz8aQyQWAVc ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ED25519)",
				},
				{
					Type: types.SSHKnownHostFingerprint,
					Path: "/home/user/.ssh/known_hosts",
					Data: "256 SHA256:cDmm4+e/BNwQpsk/Qhh39i2qiT6HcIs6qTLtIiMWzPg ec2-3-64-214-52.eu-central-1.compute.amazonaws.com (ECDSA)",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSSHKeyGenFingerprintCommandOutput(tt.args.output, tt.args.infoType, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSSHKeyGenFingerprintCommandOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
