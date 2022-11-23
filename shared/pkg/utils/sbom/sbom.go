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

package sbom

import (
	"fmt"

	"github.com/openclarity/kubeclarity/shared/pkg/converter"
	cdx_helper "github.com/openclarity/kubeclarity/shared/pkg/utils/cyclonedx_helper"
)

func GetTargetNameAndHashFromSBOM(inputSBOMFile string) (string, string, error) {
	cdxBOM, err := converter.GetCycloneDXSBOMFromFile(inputSBOMFile)
	if err != nil {
		return "", "", converter.ErrFailedToGetCycloneDXSBOM
	}

	hash, err := cdx_helper.GetComponentHash(cdxBOM.Metadata.Component)
	if err != nil {
		return "", "", fmt.Errorf("unable to get hash from original SBOM: %w", err)
	}

	return cdxBOM.Metadata.Component.Name, hash, nil
}
