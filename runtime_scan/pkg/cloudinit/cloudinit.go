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

package cloudinit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/openclarity/vmclarity/runtime_scan/pkg/types"
)

func GenerateCloudInit(scannerConfig *types.ScannerConfig, deviceName string) (*string, error) {
	vars := make(map[string]interface{})
	// parse cloud-init template
	tmpl, err := template.New("cloud-init").Parse(cloudInitTmpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cloud-init template: %v", err)
	}
	vars["Volume"] = deviceName
	vars["ScannerImage"] = scannerConfig.ScannerImage
	vars["ScannerCommand"] = scannerConfig.ScannerCommand
	vars["DirToScan"] = scannerConfig.DirectoryToScan

	scannerJobConfigB, err := json.Marshal(scannerConfig.ScannerJobConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}
	vars["Config"] = bytes.NewBuffer(scannerJobConfigB).String()
	var tmplExB bytes.Buffer
	if err := tmpl.Execute(&tmplExB, vars); err != nil {
		return nil, fmt.Errorf("failed to execute cloud-init template: %v", err)
	}

	cloudInit := tmplExB.String()
	return &cloudInit, nil
}
