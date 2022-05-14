// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package rest

import (
	"fmt"
	"gotest.tools/assert"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/openclarity/kubeclarity/api/server/models"
	"github.com/openclarity/kubeclarity/backend/pkg/runtime_scanner"
	"github.com/openclarity/kubeclarity/runtime_scan/pkg/orchestrator"
	_types "github.com/openclarity/kubeclarity/runtime_scan/pkg/types"
)

func Test_setOrUpdateScanFailDataForImage(t *testing.T) {
	type args struct {
		imageNameToScanFailData map[string]*scanFailData
		result                  *_types.ImageScanResult
	}
	tests := []struct {
		name string
		args args
		want map[string]*scanFailData
	}{
		{
			name: "new image name",
			args: args{
				imageNameToScanFailData: make(map[string]*scanFailData),
				result: &_types.ImageScanResult{
					PodName:         "PodName",
					PodNamespace:    "PodNamespace",
					PodUID:          "PodUID",
					PodLabels:       map[string]string{"key": "label"},
					ContainerName:   "ContainerName",
					ImageName:       "ImageName",
					ImageHash:       "ImageHash",
					Vulnerabilities: nil,
					Success:         false,
					ScanErrors: []*_types.ScanError{
						{
							ErrMsg:    "ErrMsg",
							ErrType:   "ErrType",
							ErrSource: _types.ScanErrSourceVul,
						},
						{
							ErrMsg:    "ErrMsg2",
							ErrType:   "ErrType",
							ErrSource: _types.ScanErrSourceVul,
						},
					},
				},
			},
			want: map[string]*scanFailData{
				"ImageName": {
					errorMessages: []string{"ErrMsg", "ErrMsg2"},
					podData: []podData{
						{
							name:      "PodName",
							namespace: "PodNamespace",
						},
					},
				},
			},
		},
		{
			name: "image name already exits",
			args: args{
				imageNameToScanFailData: map[string]*scanFailData{
					"ImageName": {
						errorMessages: []string{"ErrMsg", "ErrMsg2"},
						podData: []podData{
							{
								name:      "PodName",
								namespace: "PodNamespace",
							},
						},
					},
				},
				result: &_types.ImageScanResult{
					PodName:         "PodName2",
					PodNamespace:    "PodNamespace2",
					PodUID:          "PodUid2",
					PodLabels:       map[string]string{"key": "label"},
					ContainerName:   "ContainerName2",
					ImageName:       "ImageName",
					ImageHash:       "ImageHash",
					Vulnerabilities: nil,
					Success:         false,
					ScanErrors: []*_types.ScanError{
						{
							ErrMsg:    "ErrMsg",
							ErrType:   "ErrType",
							ErrSource: _types.ScanErrSourceVul,
						},
						{
							ErrMsg:    "ErrMsg2",
							ErrType:   "ErrType",
							ErrSource: _types.ScanErrSourceVul,
						},
					},
				},
			},
			want: map[string]*scanFailData{
				"ImageName": {
					errorMessages: []string{"ErrMsg", "ErrMsg2"},
					podData: []podData{
						{
							name:      "PodName",
							namespace: "PodNamespace",
						},
						{
							name:      "PodName2",
							namespace: "PodNamespace2",
						},
					},
				},
			},
		},
		{
			name: "different image name",
			args: args{
				imageNameToScanFailData: map[string]*scanFailData{
					"ImageName": {
						errorMessages: []string{"ErrMsg", "ErrMsg2"},
						podData: []podData{
							{
								name:      "PodName",
								namespace: "PodNamespace",
							},
						},
					},
				},
				result: &_types.ImageScanResult{
					PodName:         "PodName",
					PodNamespace:    "PodNamespace",
					PodUID:          "PodUID",
					PodLabels:       map[string]string{"key": "label"},
					ContainerName:   "ContainerName",
					ImageName:       "ImageName2",
					ImageHash:       "ImageHash2",
					Vulnerabilities: nil,
					Success:         false,
					ScanErrors: []*_types.ScanError{
						{
							ErrMsg:    "ErrMsg",
							ErrType:   "ErrType",
							ErrSource: _types.ScanErrSourceVul,
						},
						{
							ErrMsg:    "ErrMsg2",
							ErrType:   "ErrType",
							ErrSource: _types.ScanErrSourceVul,
						},
					},
				},
			},
			want: map[string]*scanFailData{
				"ImageName": {
					errorMessages: []string{"ErrMsg", "ErrMsg2"},
					podData: []podData{
						{
							name:      "PodName",
							namespace: "PodNamespace",
						},
					},
				},
				"ImageName2": {
					errorMessages: []string{"ErrMsg", "ErrMsg2"},
					podData: []podData{
						{
							name:      "PodName",
							namespace: "PodNamespace",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := setOrUpdateScanFailDataForImage(tt.args.imageNameToScanFailData, tt.args.result); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setOrUpdateScanFailDataForImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAppNameFromResult(t *testing.T) {
	type args struct {
		result *_types.ImageScanResult
	}
	tests := []struct {
		name        string
		args        args
		wantAppName string
	}{
		{
			name: "using app label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"app": "test"},
				},
			},
			wantAppName: "test",
		},
		{
			name: "using app.kubernetes.io/name label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"app.kubernetes.io/name": "test"},
				},
			},
			wantAppName: "test",
		},
		{
			name: "k8s-app label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"k8s-app": "test"},
				},
			},
			wantAppName: "test",
		},
		{
			name: "name label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"name": "test"},
				},
			},
			wantAppName: "test",
		},
		{
			name: "using pod name",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"no-app-label": "test"},
				},
			},
			wantAppName: "PodName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAppName := getAppNameFromResult(tt.args.result); gotAppName != tt.wantAppName {
				t.Errorf("getAppNameFromResult() = %v, want %v", gotAppName, tt.wantAppName)
			}
		})
	}
}

func Test_getAppVersionFromResult(t *testing.T) {
	type args struct {
		result *_types.ImageScanResult
	}
	tests := []struct {
		name           string
		args           args
		wantAppVersion string
	}{
		{
			name: "using version label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"version": "test"},
				},
			},
			wantAppVersion: "test",
		},
		{
			name: "using app.kubernetes.io/version label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"app.kubernetes.io/version": "test"},
				},
			},
			wantAppVersion: "test",
		},
		{
			name: "no version label",
			args: args{
				result: &_types.ImageScanResult{
					PodName:   "PodName",
					PodLabels: map[string]string{"no-app-version": "test"},
				},
			},
			wantAppVersion: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotAppVersion := getAppVersionFromResult(tt.args.result); gotAppVersion != tt.wantAppVersion {
				t.Errorf("getAppVersionFromResult() = %v, want %v", gotAppVersion, tt.wantAppVersion)
			}
		})
	}
}

func Test_getApplicationLabelsFromResults(t *testing.T) {
	type args struct {
		results []*_types.ImageScanResult
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "sanity",
			args: args{
				results: []*_types.ImageScanResult{
					{
						PodLabels: map[string]string{
							"k1": "v1",
							"k2": "v2",
							"k3": "v3",
						},
					},
					{
						PodLabels: map[string]string{
							"k11": "v11",
							"k22": "v22",
							"k33": "v33",
						},
					},
				},
			},
			want: []string{"k1=v1", "k2=v2", "k3=v3", "k11=v11", "k22=v22", "k33=v33"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getApplicationLabelsFromResults(tt.args.results)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getApplicationLabelsFromResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getApplicationEnvironmentsFromResults(t *testing.T) {
	type args struct {
		results []*_types.ImageScanResult
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "sanity",
			args: args{
				results: []*_types.ImageScanResult{
					{
						PodNamespace: "ns1",
					},
					{
						PodNamespace: "ns2",
					},
				},
			},
			want: []string{"ns1", "ns2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getApplicationEnvironmentsFromResults(tt.args.results)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getApplicationEnvironmentsFromResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFailures(t *testing.T) {
	type args struct {
		imageNameToScanFailData map[string]*scanFailData
	}
	tests := []struct {
		name         string
		args         args
		wantFailures []string
	}{
		{
			name: "sanity",
			args: args{
				imageNameToScanFailData: map[string]*scanFailData{
					"imageName": {
						errorMessages: []string{"err1", "err2"},
						podData: []podData{
							{
								name:      "name1",
								namespace: "namespace1",
							},
							{
								name:      "name2",
								namespace: "namespace2",
							},
						},
					},
					"imageName2": {
						errorMessages: []string{"err11", "err22"},
						podData: []podData{
							{
								name:      "name11",
								namespace: "namespace11",
							},
							{
								name:      "name22",
								namespace: "namespace22",
							},
						},
					},
				},
			},
			wantFailures: []string{
				fmt.Sprintf(failureFormat, "imageName", "name1/namespace1, name2/namespace2", "err1, err2"),
				fmt.Sprintf(failureFormat, "imageName2", "name11/namespace11, name22/namespace22", "err11, err22"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFailures := getFailures(tt.args.imageNameToScanFailData)
			sort.Strings(gotFailures)
			sort.Strings(tt.wantFailures)
			if !reflect.DeepEqual(gotFailures, tt.wantFailures) {
				t.Errorf("getFailures() = %v, want %v", gotFailures, tt.wantFailures)
			}
		})
	}
}

func Test_getPodsList(t *testing.T) {
	type args struct {
		podsData []podData
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sanity",
			args: args{
				podsData: []podData{
					{
						name:      "name",
						namespace: "namespace",
					},
					{
						name:      "name1",
						namespace: "namespace1",
					},
				},
			},
			want: "name/namespace, name1/namespace1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPodsList(tt.args.podsData); got != tt.want {
				t.Errorf("getPodsList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_getScanStatusAndScanned(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockVulScanner := orchestrator.NewMockVulnerabilitiesScanner(mockCtrl)

	tests := []struct {
		name             string
		expectVulScanner func(scanner *orchestrator.MockVulnerabilitiesScanner)
		wantStatus       models.RuntimeScanStatus
		wantScanned      int64
		doneApplyingToDB bool
	}{
		{
			name: "scan init failure",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          10,
					ImagesStartedToScan:   0,
					ImagesCompletedToScan: 0,
					Status:                _types.ScanInitFailure,
				})
			},
			wantStatus:  models.RuntimeScanStatusDONE,
			wantScanned: 0,
		},
		{
			name: "scan init",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          10,
					ImagesStartedToScan:   0,
					ImagesCompletedToScan: 0,
					Status:                _types.ScanInit,
				})
			},
			wantStatus:  models.RuntimeScanStatusINPROGRESS,
			wantScanned: 0,
		},
		{
			name: "scanning",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          10,
					ImagesStartedToScan:   4,
					ImagesCompletedToScan: 6,
					Status:                _types.Scanning,
				})
			},
			wantStatus:  models.RuntimeScanStatusINPROGRESS,
			wantScanned: 60,
		},
		{
			name: "idle",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          0,
					ImagesStartedToScan:   0,
					ImagesCompletedToScan: 0,
					Status:                _types.Idle,
				})
			},
			wantStatus:  models.RuntimeScanStatusNOTSTARTED,
			wantScanned: 0,
		},
		{
			name: "nothing to scan",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          0,
					ImagesStartedToScan:   0,
					ImagesCompletedToScan: 0,
					Status:                _types.NothingToScan,
				})
			},
			wantStatus:  models.RuntimeScanStatusDONE,
			wantScanned: 0,
		},
		{
			name: "done scanning - finalizing - applying to DB",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          10,
					ImagesStartedToScan:   10,
					ImagesCompletedToScan: 10,
					Status:                _types.DoneScanning,
				})
			},
			wantStatus:       models.RuntimeScanStatusFINALIZING,
			wantScanned:      100,
			doneApplyingToDB: false,
		},
		{
			name: "done scanning - done applying to DB",
			expectVulScanner: func(scanner *orchestrator.MockVulnerabilitiesScanner) {
				scanner.EXPECT().ScanProgress().Return(_types.ScanProgress{
					ImagesToScan:          10,
					ImagesStartedToScan:   10,
					ImagesCompletedToScan: 10,
					Status:                _types.DoneScanning,
				})
			},
			wantStatus:       models.RuntimeScanStatusDONE,
			wantScanned:      100,
			doneApplyingToDB: true,
		},
	}
	rts := runtime_scanner.CreateRuntimeScanner(mockVulScanner, nil, nil)
	for _, tt := range tests {
		tt.expectVulScanner(mockVulScanner)
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				runtimeScanner: rts,
				State: State{
					doneApplyingToDB: tt.doneApplyingToDB,
				},
			}
			got, got1 := s.getScanStatusAndScanned()
			if got != tt.wantStatus {
				t.Errorf("getScanStatusAndScanned() got = %v, wantStatus %v", got, tt.wantStatus)
			}
			if got1 != tt.wantScanned {
				t.Errorf("getScanStatusAndScanned() got1 = %v, wantScanned %v", got1, tt.wantScanned)
			}
		})
	}
}

func int64Ptr(val int64) *int64 {
	ret := val
	return &ret
}

func createTime(t *testing.T, ti string) time.Time {
	tim, err := time.Parse(time.RFC3339, ti)
	assert.NilError(t, err)
	return tim
}

func Test_getIntervalAndStartTimeFromByDaysScheduleScanConfig(t *testing.T) {
	timeNow, err := time.Parse(time.RFC3339, "2022-05-08T18:23:21+00:00")
	assert.NilError(t, err)

	type args struct {
		timeNow    time.Time
		scanConfig *models.ByDaysScheduleScanConfig
	}
	tests := []struct {
		name          string
		args          args
		wantInterval  time.Duration
		wantStartTime time.Time
	}{
		{
			name: "time of day to start scan (17:20) is before time now (18:23)",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.ByDaysScheduleScanConfig{
					DaysInterval: 2,
					TimeOfDay: &models.TimeOfDay{
						Hour:   int64Ptr(17),
						Minute: int64Ptr(20),
					},
				},
			},
			wantInterval:  2 * secondsInDay * time.Second,
			wantStartTime: createTime(t, "2022-05-08T17:20:00+00:00"),
		},
		{
			name: "time of day to start scan (19:20) is after time now (18:23)",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.ByDaysScheduleScanConfig{
					DaysInterval: 2,
					TimeOfDay: &models.TimeOfDay{
						Hour:   int64Ptr(19),
						Minute: int64Ptr(20),
					},
				},
			},
			wantInterval:  2 * secondsInDay * time.Second,
			wantStartTime: createTime(t, "2022-05-08T19:20:00+00:00"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getIntervalAndStartTimeFromByDaysScheduleScanConfig(tt.args.timeNow, tt.args.scanConfig)
			if got != tt.wantInterval {
				t.Errorf("getIntervalAndStartTimeFromByDaysScheduleScanConfig() got = %v, wantInterval %v", got, tt.wantInterval)
			}
			if !got1.Equal(tt.wantStartTime) {
				t.Errorf("getIntervalAndStartTimeFromByDaysScheduleScanConfig() got1 = %v, wantStartTime %v", got1, tt.wantStartTime)
			}
		})
	}
}

func Test_getIntervalAndStartTimeFromByHoursScheduleScanConfig(t *testing.T) {
	timeNow, err := time.Parse(time.RFC3339, "2022-05-08T18:23:21+00:00")
	assert.NilError(t, err)

	type args struct {
		timeNow    time.Time
		scanConfig *models.ByHoursScheduleScanConfig
	}
	tests := []struct {
		name         string
		args         args
		wantInterval time.Duration
		wantTime     time.Time
	}{
		{
			name: "sanity",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.ByHoursScheduleScanConfig{
					HoursInterval: 3,
				},
			},
			wantInterval: 3 * secondsInHour * time.Second,
			wantTime:     timeNow,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getIntervalAndStartTimeFromByHoursScheduleScanConfig(tt.args.timeNow, tt.args.scanConfig)
			if got != tt.wantInterval {
				t.Errorf("getIntervalAndStartTimeFromByHoursScheduleScanConfig() got = %v, wantInterval %v", got, tt.wantInterval)
			}
			if !got1.Equal(tt.wantTime) {
				t.Errorf("getIntervalAndStartTimeFromByHoursScheduleScanConfig() got1 = %v, wantTime %v", got1, tt.wantTime)
			}
		})
	}
}

func Test_getIntervalAndStartTimeFromWeeklyScheduleScanConfig(t *testing.T) {
	// monday
	timeNow, err := time.Parse(time.RFC3339, "2022-05-09T18:23:21+00:00")
	assert.NilError(t, err)

	type args struct {
		timeNow    time.Time
		scanConfig *models.WeeklyScheduleScanConfig
	}
	tests := []struct {
		name         string
		args         args
		wantInterval time.Duration
		wantTime     time.Time
	}{
		{
			name: "Start in the next day",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.WeeklyScheduleScanConfig{
					DayInWeek: 3,
					TimeOfDay: &models.TimeOfDay{
						Hour:   int64Ptr(19),
						Minute: int64Ptr(20),
					},
				},
			},
			wantInterval: secondsInWeek * time.Second,
			wantTime:     createTime(t, "2022-05-10T19:20:00+00:00"),
		},
		{
			name: "Start in earlier day from now",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.WeeklyScheduleScanConfig{
					DayInWeek: 1,
					TimeOfDay: &models.TimeOfDay{
						Hour:   int64Ptr(19),
						Minute: int64Ptr(20),
					},
				},
			},
			wantInterval: secondsInWeek * time.Second,
			wantTime:     createTime(t, "2022-05-08T19:20:00+00:00"),
		},
		{
			name: "Start in the same day, later hour",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.WeeklyScheduleScanConfig{
					DayInWeek: 2,
					TimeOfDay: &models.TimeOfDay{
						Hour:   int64Ptr(19),
						Minute: int64Ptr(20),
					},
				},
			},
			wantInterval: secondsInWeek * time.Second,
			wantTime:     createTime(t, "2022-05-09T19:20:00+00:00"),
		},
		{
			name: "Start in the same day, hour has pass",
			args: args{
				timeNow: timeNow,
				scanConfig: &models.WeeklyScheduleScanConfig{
					DayInWeek: 2,
					TimeOfDay: &models.TimeOfDay{
						Hour:   int64Ptr(15),
						Minute: int64Ptr(20),
					},
				},
			},
			wantInterval: secondsInWeek * time.Second,
			wantTime:     createTime(t, "2022-05-09T15:20:00+00:00"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getIntervalAndStartTimeFromWeeklyScheduleScanConfig(tt.args.timeNow, tt.args.scanConfig)
			if got != tt.wantInterval {
				t.Errorf("getIntervalAndStartTimeFromWeeklyScheduleScanConfig() got = %v, wantInterval %v", got, tt.wantInterval)
			}
			if !got1.Equal(tt.wantTime) {
				t.Errorf("getIntervalAndStartTimeFromWeeklyScheduleScanConfig() got1 = %v, wantTime %v", got1, tt.wantTime)
			}
		})
	}
}
