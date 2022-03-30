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

	"github.com/cisco-open/kubei/shared/pkg/analyzer/cdx_gomod"
	"github.com/cisco-open/kubei/shared/pkg/analyzer/syft"
	"github.com/cisco-open/kubei/shared/pkg/config"
	"github.com/cisco-open/kubei/shared/pkg/job_manager"
)

func CreateAnalyzerJob(analyzerName string, config *config.Config, logger *logrus.Entry, resultChan chan job_manager.Result) job_manager.Job {
	switch analyzerName {
	case syft.AnalyzerName:
		return syft.New(config, logger, resultChan)
	case cdx_gomod.AnalyzerName:
		return cdx_gomod.New(config, logger, resultChan)
	default:
		logger.Fatalf("Unknown analyzer: %v", analyzerName)
	}

	return nil
}
