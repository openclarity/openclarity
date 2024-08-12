// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package cisdocker

import (
	"testing"
	"time"

	dockle_config "github.com/Portshift/dockle/config"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	logrusTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/openclarity/vmclarity/scanner/common"
	"github.com/openclarity/vmclarity/scanner/families/misconfiguration/cisdocker/config"
)

func TestCreateDockleConfig(t *testing.T) {
	loggerDebug, _ := logrusTest.NewNullLogger()
	logEntryDebug := loggerDebug.WithField("test", "valueToMisconfiguration")
	logEntryDebug.Logger.SetLevel(logrus.DebugLevel)

	loggerInfo, _ := logrusTest.NewNullLogger()
	logEntryInfo := loggerInfo.WithField("test", "valueToMisconfiguration")
	logEntryInfo.Logger.SetLevel(logrus.InfoLevel)

	tests := []struct {
		name   string
		logger *logrus.Entry
		config config.Config
		input  string
		want   *dockle_config.Config
	}{
		{
			name:   "Scanner with defaults",
			logger: logEntryDebug,
			config: config.Config{},
			input:  "node:slim",
			want: &dockle_config.Config{
				Debug:      true,
				Timeout:    2 * time.Minute,
				Username:   "",
				Password:   "",
				Insecure:   false,
				NonSSL:     false,
				ImageName:  "node:slim",
				LocalImage: true,
			},
		},
		{
			name:   "Scanner with configuration",
			logger: logEntryInfo,
			config: config.Config{
				Timeout: 1 * time.Minute,
				Registry: &common.Registry{
					SkipVerifyTLS: true,
					UseHTTP:       true,
					Auths: []common.RegistryAuth{{
						Username: "testuser",
						Password: "testpassword",
					}},
				},
			},
			input: "node:latest",
			want: &dockle_config.Config{
				Debug:      false,
				Timeout:    1 * time.Minute,
				Username:   "testuser",
				Password:   "testpassword",
				Insecure:   true,
				NonSSL:     true,
				ImageName:  "node:latest",
				LocalImage: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createDockleConfig(tt.logger, common.IMAGE, tt.input, tt.config)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("NewReportParser() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
