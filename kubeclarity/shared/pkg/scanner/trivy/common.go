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

package trivy

import (
	log "github.com/sirupsen/logrus"

	trivyLog "github.com/aquasecurity/trivy/pkg/log"

	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
)

const ScannerName = "trivy"

func New(c job_manager.IsConfig,
	logger *log.Entry,
	resultChan chan job_manager.Result,
) job_manager.Job {
	conf := c.(*config.Config) // nolint:forcetypeassert

	// For now disable all the logging from trivy
	err := trivyLog.InitLogger(false, true)
	if err != nil {
		logger.Fatalf("Unable to init trivy logger %v", err)
	}

	return &LocalScanner{
		logger:     logger.Dup().WithField("scanner", ScannerName),
		config:     config.CreateLocalScannerTrivyConfigEx(conf.Scanner, conf.Registry),
		resultChan: resultChan,
		localImage: conf.LocalImageScan,
	}
}
