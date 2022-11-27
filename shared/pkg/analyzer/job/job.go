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

package job

import (
	"github.com/sirupsen/logrus"

	_ "github.com/openclarity/kubeclarity/shared/pkg/analyzer/cdx_gomod"
	_ "github.com/openclarity/kubeclarity/shared/pkg/analyzer/syft"
	_ "github.com/openclarity/kubeclarity/shared/pkg/analyzer/trivy"
	"github.com/openclarity/kubeclarity/shared/pkg/job_factory"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
)

func CreateAnalyzerJob(scannerName string, conf job_manager.IsConfig, logger *logrus.Entry, resultChan chan job_manager.Result) job_manager.Job {
	createJobFunc, ok := job_factory.GetCreateJobFuncs()[scannerName]
	if !ok {
		logrus.Fatalf("Unregister analyzer: %v", scannerName)
	}
	return createJobFunc(conf, logger, resultChan)
}
