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
	"github.com/openclarity/kubeclarity/shared/v2/pkg/job_manager"
	"github.com/openclarity/kubeclarity/shared/v2/pkg/scanner/dependency_track"
	"github.com/openclarity/kubeclarity/shared/v2/pkg/scanner/grype"
	"github.com/openclarity/kubeclarity/shared/v2/pkg/scanner/trivy"
)

var Factory = job_manager.NewJobFactory()

func init() {
	Factory.Register(grype.ScannerName, grype.New)
	Factory.Register(dependency_track.ScannerName, dependency_track.New)
	Factory.Register(trivy.ScannerName, trivy.New)
}
