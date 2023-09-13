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

package export

import (
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/api/client/client"
	"github.com/openclarity/kubeclarity/api/client/client/operations"
	"github.com/openclarity/kubeclarity/api/client/models"
	"github.com/openclarity/kubeclarity/shared/pkg/analyzer"
	"github.com/openclarity/kubeclarity/shared/pkg/utils"
	cdx_helper "github.com/openclarity/kubeclarity/shared/pkg/utils/cyclonedx_helper"
)

func Export(apiClient *client.KubeClarityAPIs, mergedResults *analyzer.MergedResults, id string) error {
	// In the case of the source is a file, if the syft version below 0.32, it doesn't set the metatdata component.
	if mergedResults.SrcMetaData.Component == nil {
		return fmt.Errorf("failed to export analysis result, metadata component is empty")
	}

	hash, err := cdx_helper.GetComponentHash(mergedResults.SrcMetaData.Component)
	if err != nil {
		return fmt.Errorf("unable to get hash from src metadata: %w", err)
	}

	// create ApplicationContentAnalysis from mergedResults
	body := createApplicationContentAnalysis(mergedResults, hash)
	// create post parameters
	postParams := operations.NewPostApplicationsContentAnalysisIDParams().WithID(id).WithBody(body)

	_, err = apiClient.Operations.PostApplicationsContentAnalysisID(postParams)
	if err != nil {
		return fmt.Errorf("failed to post analysis report: %v", err)
	}

	return nil
}

func createApplicationContentAnalysis(m *analyzer.MergedResults, hash string) *models.ApplicationContentAnalysis {
	return &models.ApplicationContentAnalysis{
		Resources: []*models.ResourceContentAnalysis{
			{
				Packages: createPackagesContentAnalysis(m),
				Resource: &models.ResourceInfo{
					ResourceHash: hash,
					ResourceName: m.SrcMetaData.Component.Name,
					ResourceType: getResourceType(m),
				},
			},
		},
	}
}

func createPackagesContentAnalysis(m *analyzer.MergedResults) []*models.PackageContentAnalysis {
	packageContentAnalysis := make([]*models.PackageContentAnalysis, 0, len(m.MergedComponentByKey))
	for _, component := range m.MergedComponentByKey {
		packageContentAnalysis = append(packageContentAnalysis, &models.PackageContentAnalysis{
			Analyzers: component.AnalyzerInfo,
			Package:   getPackageInfo(component.Component),
		})
	}
	return packageContentAnalysis
}

func getResourceType(m *analyzer.MergedResults) models.ResourceType {
	switch m.Source {
	case utils.IMAGE, utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		return models.ResourceTypeIMAGE
	case utils.DIR:
		return models.ResourceTypeDIRECTORY
	case utils.FILE:
		return models.ResourceTypeFILE
	case utils.ROOTFS:
		return models.ResourceTypeROOTFS
	case utils.SBOM:
		log.Errorf("SBOM is unsupported resource type")
	default:
		log.Errorf("Unknown resource type %s", m.Source)
	}
	return ""
}

func getPackageInfo(component cdx.Component) *models.PackageInfo {
	// In the case of packages there can be only one license in the licence list
	var license string
	licenses := cdx_helper.GetComponentLicenses(component)
	if len(licenses) > 0 {
		license = licenses[0]
	}
	return &models.PackageInfo{
		Name:     component.Name,
		Version:  component.Version,
		License:  license,
		Language: cdx_helper.GetComponentLanguage(component),
	}
}
