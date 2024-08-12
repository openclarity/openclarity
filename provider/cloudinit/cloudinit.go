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

package cloudinit

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

//go:embed cloud-init.tmpl.yaml
var cloudInitTemplate string

func New(data any) (string, error) {
	tmpl, err := template.New("cloud-init").Funcs(sprig.FuncMap()).Parse(cloudInitTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse cloud-init template: %w", err)
	}

	var b bytes.Buffer
	if err = tmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("failed to generate cloud-init from template: %w", err)
	}

	return b.String(), nil
}
