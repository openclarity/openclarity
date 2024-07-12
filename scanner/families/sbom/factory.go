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

package sbom

import (
	"github.com/openclarity/vmclarity/scanner/families/sbom/syft"
	"github.com/openclarity/vmclarity/scanner/families/sbom/trivy"
	"github.com/openclarity/vmclarity/scanner/families/sbom/types"
	"github.com/openclarity/vmclarity/scanner/families/sbom/windows"
	"github.com/openclarity/vmclarity/scanner/internal/scan_manager"
)

// Factory receives parent config that contains shared fields to enable override.
var Factory = scan_manager.NewFactory[types.Config, *types.ScannerResult]()

func init() {
	Factory.Register(syft.AnalyzerName, syft.New)
	Factory.Register(trivy.AnalyzerName, trivy.New)
	Factory.Register(windows.AnalyzerName, windows.New)
}
