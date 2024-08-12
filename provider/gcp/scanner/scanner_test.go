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

package scanner

import (
	"reflect"
	"testing"

	"cloud.google.com/go/compute/apiv1/computepb"

	"github.com/openclarity/vmclarity/core/to"
)

func Test_getInstanceBootDisk(t *testing.T) {
	type args struct {
		vm *computepb.Instance
	}
	tests := []struct {
		name    string
		args    args
		want    *computepb.AttachedDisk
		wantErr bool
	}{
		{
			name: "found",
			args: args{
				vm: &computepb.Instance{
					Disks: []*computepb.AttachedDisk{
						{
							DeviceName: to.Ptr("device1"),
							Boot:       to.Ptr(true),
						},
						{
							DeviceName: to.Ptr("device2"),
							Boot:       to.Ptr(false),
						},
					},
				},
			},
			want: &computepb.AttachedDisk{
				DeviceName: to.Ptr("device1"),
				Boot:       to.Ptr(true),
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				vm: &computepb.Instance{
					Disks: []*computepb.AttachedDisk{
						{
							DeviceName: to.Ptr("device1"),
							Boot:       to.Ptr(false),
						},
						{
							DeviceName: to.Ptr("device2"),
							Boot:       to.Ptr(false),
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInstanceBootDisk(tt.args.vm)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInstanceBootDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getInstanceBootDisk() got = %v, want %v", got, tt.want)
			}
		})
	}
}
