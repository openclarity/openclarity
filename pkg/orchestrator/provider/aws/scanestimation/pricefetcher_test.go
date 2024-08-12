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

package scanestimation

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pricing/types"

	"github.com/openclarity/vmclarity/pkg/shared/utils"
)

func Test_createGetProductsFilters(t *testing.T) {
	type args struct {
		usageType   string
		serviceCode string
		operation   string
	}
	tests := []struct {
		name string
		args args
		want []types.Filter
	}{
		{
			name: "no operation",
			args: args{
				usageType:   "sampleUsageType",
				serviceCode: "sampleServiceCode",
				operation:   "",
			},
			want: []types.Filter{
				{
					Field: utils.PointerTo("ServiceCode"),
					Type:  "TERM_MATCH",
					Value: utils.PointerTo("sampleServiceCode"),
				},
				{
					Field: utils.PointerTo("usagetype"),
					Type:  "TERM_MATCH",
					Value: utils.PointerTo("sampleUsageType"),
				},
			},
		},
		{
			name: "with operation",
			args: args{
				usageType:   "sampleUsageType",
				serviceCode: "sampleServiceCode",
				operation:   "sampleOperation",
			},
			want: []types.Filter{
				{
					Field: utils.PointerTo("ServiceCode"),
					Type:  "TERM_MATCH",
					Value: utils.PointerTo("sampleServiceCode"),
				},
				{
					Field: utils.PointerTo("usagetype"),
					Type:  "TERM_MATCH",
					Value: utils.PointerTo("sampleUsageType"),
				},
				{
					Field: utils.PointerTo("operation"),
					Type:  "TERM_MATCH",
					Value: utils.PointerTo("sampleOperation"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createGetProductsFilters(tt.args.usageType, tt.args.serviceCode, tt.args.operation); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createGetProductsFilters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPricePerUnitFromJsonPriceList(t *testing.T) {
	type args struct {
		jsonPriceList string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPricePerUnitFromJSONPriceList(tt.args.jsonPriceList)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPricePerUnitFromJSONPriceList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPricePerUnitFromJSONPriceList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
