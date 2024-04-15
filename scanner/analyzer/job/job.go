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
	"github.com/openclarity/vmclarity/scanner/analyzer/syft"
	"github.com/openclarity/vmclarity/scanner/analyzer/trivy"
	"github.com/openclarity/vmclarity/scanner/analyzer/windows"
	"github.com/openclarity/vmclarity/scanner/job_manager"
)

var Factory = job_manager.NewJobFactory()

func init() {
	Factory.Register(trivy.AnalyzerName, trivy.New)
	Factory.Register(syft.AnalyzerName, syft.New)
	Factory.Register(windows.AnalyzerName, windows.New)
}
