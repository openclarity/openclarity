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
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type Data struct {
	Volume           string // Volume to mount e.g. /dev/sdc
	ScannerCLIConfig string // Scanner families configuration file yaml
	ScannerImage     string // Scanner container image to use
	ServerAddress    string // IP address of VMClarity backend for export
	ScanResultID     string // ScanResult ID to export the results to
}

func GenerateCloudInit(data Data) (string, error) {
	// parse cloud-init template
	tmpl, err := template.New("cloud-init").Funcs(sprig.FuncMap()).Parse(cloudInitTmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse cloud-init template: %v", err)
	}

	// execute template using data
	var tmplExB bytes.Buffer
	if err := tmpl.Execute(&tmplExB, data); err != nil {
		return "", fmt.Errorf("failed to execute cloud-init template: %v", err)
	}
	return tmplExB.String(), nil
}
