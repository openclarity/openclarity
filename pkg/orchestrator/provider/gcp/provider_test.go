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

package gcp

import (
	"reflect"
	"testing"

	"github.com/openclarity/vmclarity/api/models"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/google/go-cmp/cmp"

	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_convertLabelsToTags(t *testing.T) {
	tests := []struct {
		name string
		args map[string]string
		want []models.Tag
	}{
		{
			name: "sanity",
			args: map[string]string{
				"valid-tag": "valid-value",
			},
			want: []models.Tag{{
				Key: "valid-tag", Value: "valid-value",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertLabelsToTags(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertLabelsToTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
							DeviceName: utils.PointerTo("device1"),
							Boot:       utils.PointerTo(true),
						},
						{
							DeviceName: utils.PointerTo("device2"),
							Boot:       utils.PointerTo(false),
						},
					},
				},
			},
			want: &computepb.AttachedDisk{
				DeviceName: utils.PointerTo("device1"),
				Boot:       utils.PointerTo(true),
			},
			wantErr: false,
		},
		{
			name: "not found",
			args: args{
				vm: &computepb.Instance{
					Disks: []*computepb.AttachedDisk{
						{
							DeviceName: utils.PointerTo("device1"),
							Boot:       utils.PointerTo(false),
						},
						{
							DeviceName: utils.PointerTo("device2"),
							Boot:       utils.PointerTo(false),
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

func Test_getZonesLastPart(t *testing.T) {
	type args struct {
		zones []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{
				zones: []string{},
			},
			want: []string{},
		},
		{
			name: "get two zones",
			args: args{
				zones: []string{"https://www.googleapis.com/compute/v1/projects/gcp-etigcp-nprd-12855/zones/us-central1-c", "https://www.googleapis.com/compute/v1/projects/gcp-etigcp-nprd-12855/zones/us-central1-a"},
			},
			want: []string{"us-central1-c", "us-central1-a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getZonesLastPart(tt.args.zones)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("getZonesLastPart() mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
