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

package server

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	dbtypes "github.com/openclarity/vmclarity/api/server/database/types"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars map[string]string

		ExpectedNewErrorMatcher types.GomegaMatcher
		ExpectedConfig          *Config
	}{
		{
			Name: "Custom config",
			EnvVars: map[string]string{
				"VMCLARITY_APISERVER_LISTEN_ADDRESS":      "example.com:8889",
				"VMCLARITY_APISERVER_HEALTHCHECK_ADDRESS": "example.com:18888",
				"VMCLARITY_APISERVER_DATABASE_DRIVER":     "POSTGRES",
				"VMCLARITY_APISERVER_DB_NAME":             "testname",
				"VMCLARITY_APISERVER_DB_USER":             "testuser",
				"VMCLARITY_APISERVER_DB_PASS":             "testpass",
				"VMCLARITY_APISERVER_DB_HOST":             "postgresql",
				"VMCLARITY_APISERVER_DB_PORT":             "5432",
				"VMCLARITY_APISERVER_ENABLE_DB_INFO_LOGS": "true",
				"VMCLARITY_APISERVER_LOCAL_DB_PATH":       "/data/vmclarity.db",
			},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ListenAddress:      "example.com:8889",
				HealthCheckAddress: "example.com:18888",
				DatabaseDriver:     dbtypes.DBDriverTypePostgres,
				DBName:             "testname",
				DBUser:             "testuser",
				DBPassword:         "testpass",
				DBHost:             "postgresql",
				DBPort:             "5432",
				EnableDBInfoLogs:   true,
				LocalDBPath:        "/data/vmclarity.db",
			},
		},
		{
			Name:                    "Default config",
			EnvVars:                 map[string]string{},
			ExpectedNewErrorMatcher: Not(HaveOccurred()),
			ExpectedConfig: &Config{
				ListenAddress:      "0.0.0.0:8888",
				HealthCheckAddress: "0.0.0.0:8081",
				DatabaseDriver:     "LOCAL",
				DBName:             "",
				DBUser:             "",
				DBPassword:         "",
				DBHost:             "",
				DBPort:             "",
				EnableDBInfoLogs:   false,
				LocalDBPath:        "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			os.Clearenv()
			for k, v := range test.EnvVars {
				err := os.Setenv(k, v)
				g.Expect(err).Should(Not(HaveOccurred()))
			}

			config, err := NewConfig()

			g.Expect(err).Should(test.ExpectedNewErrorMatcher)
			g.Expect(config).Should(BeEquivalentTo(test.ExpectedConfig))
		})
	}
}
