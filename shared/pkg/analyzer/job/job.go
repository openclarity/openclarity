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
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/cdx_gomod"
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/syft"
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/trivy"
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer/types"
	"github.com/openclarity/kubeclarity/shared/pkg/config"
	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
)

var Factory = job_manager.NewJobFactory[*config.Config, utils.SourceInput, types.Results]()

func init() {
	Factory.Register(trivy.AnalyzerName, trivy.New)
	Factory.Register(syft.AnalyzerName, syft.New)
	Factory.Register(cdx_gomod.AnalyzerName, cdx_gomod.New)
}
