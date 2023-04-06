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

package presenter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openclarity/vmclarity/shared/pkg/families"
	"github.com/openclarity/vmclarity/shared/pkg/families/exploits"
	"github.com/openclarity/vmclarity/shared/pkg/families/malware"
	"github.com/openclarity/vmclarity/shared/pkg/families/results"
	"github.com/openclarity/vmclarity/shared/pkg/families/sbom"
	"github.com/openclarity/vmclarity/shared/pkg/families/secrets"
	"github.com/openclarity/vmclarity/shared/pkg/families/vulnerabilities"
)

type DefaultPresenter struct {
	Writer

	FamiliesConfig *families.Config
}

func (p *DefaultPresenter) ExportSbomResult(_ context.Context, res *results.Results, _ families.RunErrors) error {
	sbomResults, err := results.GetResult[*sbom.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get sbom results: %w", err)
	}

	outputFormat := p.FamiliesConfig.SBOM.AnalyzersConfig.Analyzer.OutputFormat
	sbomBytes, err := sbomResults.EncodeToBytes(outputFormat)
	if err != nil {
		return fmt.Errorf("failed to encode sbom results to bytes: %w", err)
	}

	err = p.Write(sbomBytes, "sbom.cdx")
	if err != nil {
		return fmt.Errorf("failed to output sbom results: %w", err)
	}
	return nil
}

func (p *DefaultPresenter) ExportVulResult(_ context.Context, res *results.Results, _ families.RunErrors) error {
	vulnerabilitiesResults, err := results.GetResult[*vulnerabilities.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get sbom results: %w", err)
	}

	bytes, err := json.Marshal(vulnerabilitiesResults.MergedResults)
	if err != nil {
		return fmt.Errorf("failed to output vulnerabilities results: %w", err)
	}
	err = p.Write(bytes, "vulnerabilities.json")
	if err != nil {
		return fmt.Errorf("failed to output vulnerabilities results: %w", err)
	}
	return nil
}

func (p *DefaultPresenter) ExportSecretsResult(_ context.Context, res *results.Results, _ families.RunErrors) error {
	secretsResults, err := results.GetResult[*secrets.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get secrets results: %w", err)
	}

	bytes, err := json.Marshal(secretsResults)
	if err != nil {
		return fmt.Errorf("failed to output secrets results: %w", err)
	}
	err = p.Write(bytes, "secrets.json")
	if err != nil {
		return fmt.Errorf("failed to output secrets results: %w", err)
	}
	return nil
}

func (p *DefaultPresenter) ExportMalwareResult(_ context.Context, res *results.Results, _ families.RunErrors) error {
	malwareResults, err := results.GetResult[*malware.MergedResults](res)
	if err != nil {
		return fmt.Errorf("failed to get malware results: %w", err)
	}

	bytes, err := json.Marshal(malwareResults)
	if err != nil {
		return fmt.Errorf("failed to marshal malware results: %w", err)
	}
	err = p.Write(bytes, "malware.json")
	if err != nil {
		return fmt.Errorf("failed to output  %w", err)
	}
	return nil
}

func (p *DefaultPresenter) ExportExploitsResult(_ context.Context, res *results.Results, _ families.RunErrors) error {
	exploitsResults, err := results.GetResult[*exploits.Results](res)
	if err != nil {
		return fmt.Errorf("failed to get exploits results: %w", err)
	}

	bytes, err := json.Marshal(exploitsResults)
	if err != nil {
		return fmt.Errorf("failed to marshal exploits results: %w", err)
	}
	err = p.Write(bytes, "exploits.json")
	if err != nil {
		return fmt.Errorf("failed to output exploits results: %w", err)
	}
	return nil
}

func (p *DefaultPresenter) ExportMisconfigurationResult(context.Context, *results.Results, families.RunErrors) error {
	// TODO: implement
	return nil
}
