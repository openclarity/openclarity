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

package findingkey

import (
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
)

func GenerateFindingKey(findingInfo *models.Finding_FindingInfo) (string, error) {
	value, err := findingInfo.ValueByDiscriminator()
	if err != nil {
		return "", fmt.Errorf("failed to value by discriminator from finding info: %v", err)
	}

	switch info := value.(type) {
	case models.ExploitFindingInfo:
		return GenerateExploitFindingUniqueKey(info), nil
	case models.VulnerabilityFindingInfo:
		return GenerateVulnerabilityKey(info).String(), nil
	case models.MalwareFindingInfo:
		return GenerateMalwareKey(info).String(), nil
	case models.MisconfigurationFindingInfo:
		return GenerateMisconfigurationKey(info).String(), nil
	case models.RootkitFindingInfo:
		return GenerateRootkitKey(info).String(), nil
	case models.SecretFindingInfo:
		return GenerateSecretKey(info).String(), nil
	case models.PackageFindingInfo:
		return GeneratePackageKey(info).String(), nil
	default:
		return "", fmt.Errorf("unsupported finding info type %T", value)
	}
}
