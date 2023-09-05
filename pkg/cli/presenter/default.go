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

	"github.com/openclarity/vmclarity/pkg/shared/families"
	"github.com/openclarity/vmclarity/pkg/shared/families/exploits"
	"github.com/openclarity/vmclarity/pkg/shared/families/infofinder"
	"github.com/openclarity/vmclarity/pkg/shared/families/malware"
	"github.com/openclarity/vmclarity/pkg/shared/families/rootkits"
	"github.com/openclarity/vmclarity/pkg/shared/families/sbom"
	"github.com/openclarity/vmclarity/pkg/shared/families/secrets"
	"github.com/openclarity/vmclarity/pkg/shared/families/types"
	"github.com/openclarity/vmclarity/pkg/shared/families/vulnerabilities"
)

type DefaultPresenter struct {
	Writer

	FamiliesConfig *families.Config
}

func (p *DefaultPresenter) ExportFamilyResult(ctx context.Context, res families.FamilyResult) error {
	var err error

	switch res.FamilyType {
	case types.SBOM:
		err = p.ExportSbomResult(ctx, res)
	case types.Vulnerabilities:
		err = p.ExportVulResult(ctx, res)
	case types.Secrets:
		err = p.ExportSecretsResult(ctx, res)
	case types.Exploits:
		err = p.ExportExploitsResult(ctx, res)
	case types.Misconfiguration:
		err = p.ExportMisconfigurationResult(ctx, res)
	case types.Rootkits:
		err = p.ExportRootkitResult(ctx, res)
	case types.Malware:
		err = p.ExportMalwareResult(ctx, res)
	case types.InfoFinder:
		err = p.ExportInfoFinderResult(ctx, res)
	}

	return err
}

func (p *DefaultPresenter) ExportSbomResult(_ context.Context, res families.FamilyResult) error {
	sbomResults, ok := res.Result.(*sbom.Results)
	if !ok {
		return fmt.Errorf("failed to convert to sbom results")
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

func (p *DefaultPresenter) ExportVulResult(_ context.Context, res families.FamilyResult) error {
	vulnerabilitiesResults, ok := res.Result.(*vulnerabilities.Results)
	if !ok {
		return fmt.Errorf("failed to convert to vulnerabilities results")
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

func (p *DefaultPresenter) ExportSecretsResult(_ context.Context, res families.FamilyResult) error {
	secretsResults, ok := res.Result.(*secrets.Results)
	if !ok {
		return fmt.Errorf("failed to convert to secrets results")
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

func (p *DefaultPresenter) ExportMalwareResult(_ context.Context, res families.FamilyResult) error {
	malwareResults, ok := res.Result.(*malware.MergedResults)
	if !ok {
		return fmt.Errorf("failed to convert to malware results")
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

func (p *DefaultPresenter) ExportExploitsResult(_ context.Context, res families.FamilyResult) error {
	exploitsResults, ok := res.Result.(*exploits.Results)
	if !ok {
		return fmt.Errorf("failed to convert to exploits results")
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

func (p *DefaultPresenter) ExportMisconfigurationResult(context.Context, families.FamilyResult) error {
	// TODO: implement
	return nil
}

func (p *DefaultPresenter) ExportRootkitResult(_ context.Context, res families.FamilyResult) error {
	rootkitsResults, ok := res.Result.(*rootkits.Results)
	if !ok {
		return fmt.Errorf("failed to convert to rootkits results")
	}

	bytes, err := json.Marshal(rootkitsResults)
	if err != nil {
		return fmt.Errorf("failed to marshal rootkits results: %w", err)
	}
	err = p.Write(bytes, "rootkits.json")
	if err != nil {
		return fmt.Errorf("failed to output rootkits results: %w", err)
	}
	return nil
}

func (p *DefaultPresenter) ExportInfoFinderResult(_ context.Context, res families.FamilyResult) error {
	if res.Result == nil {
		return nil
	}

	infoFinderResults, ok := res.Result.(*infofinder.Results)
	if !ok {
		return fmt.Errorf("failed to convert to infofinder results")
	}

	bytes, err := json.Marshal(infoFinderResults)
	if err != nil {
		return fmt.Errorf("failed to marshal infofinder results: %w", err)
	}

	if err = p.Write(bytes, "infofinder.json"); err != nil {
		return fmt.Errorf("failed to output infofinder results: %w", err)
	}

	return nil
}
