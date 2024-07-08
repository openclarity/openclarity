// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package formatter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Checkmarx/kics/pkg/model"

	"github.com/openclarity/vmclarity/plugins/sdk-go/types"
)

var mapKICSSeverity = map[model.Severity]types.MisconfigurationSeverity{
	model.SeverityHigh:   types.MisconfigurationSeverityHigh,
	model.SeverityMedium: types.MisconfigurationSeverityMedium,
	model.SeverityLow:    types.MisconfigurationSeverityLow,
	model.SeverityInfo:   types.MisconfigurationSeverityInfo,
	model.SeverityTrace:  types.MisconfigurationSeverityInfo,
}

func FormatJSONOutput(rawOutputDir string) (model.Summary, error) {
	var summaryJSON model.Summary
	err := decodeFile(filepath.Join(rawOutputDir, "kics.json"), &summaryJSON)
	if err != nil {
		return model.Summary{}, fmt.Errorf("failed to decode kics.json: %w", err)
	}

	return summaryJSON, nil
}

func FormatVMClarityOutput(summaryJSON model.Summary) (*[]types.Misconfiguration, error) {
	var misconfigurations []types.Misconfiguration
	for _, q := range summaryJSON.Queries {
		for _, file := range q.Files {
			misconfigurations = append(misconfigurations, types.Misconfiguration{
				Id:          types.Ptr(file.SimilarityID),
				Location:    types.Ptr(file.FileName + "#" + strconv.Itoa(file.Line)),
				Category:    types.Ptr(q.Category + ":" + string(file.IssueType)),
				Message:     types.Ptr(file.KeyActualValue),
				Description: types.Ptr(q.Description),
				Remediation: types.Ptr(file.KeyExpectedValue),
				Severity:    types.Ptr(mapKICSSeverity[q.Severity]),
			})
		}
	}

	return types.Ptr(misconfigurations), nil
}

func FormatSarifOutput(rawOutputDir string) (*interface{}, error) {
	var summarySarif interface{}
	err := decodeFile(filepath.Join(rawOutputDir, "kics.sarif"), &summarySarif)
	if err != nil {
		return nil, fmt.Errorf("failed to decode kics.sarif: %w", err)
	}

	return types.Ptr(summarySarif), nil
}

func decodeFile(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode file: %w", err)
	}

	return nil
}
