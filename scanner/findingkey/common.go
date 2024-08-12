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

	apitypes "github.com/openclarity/vmclarity/api/types"
)

func GenerateFindingKey(findingInfo *apitypes.FindingInfo) (string, error) {
	value, err := findingInfo.ValueByDiscriminator()
	if err != nil {
		return "", fmt.Errorf("failed to value by discriminator from finding info: %w", err)
	}

	switch info := value.(type) {
	case apitypes.ExploitFindingInfo:
		return GenerateExploitKey(info).ExploitString(), nil
	case apitypes.VulnerabilityFindingInfo:
		return GenerateVulnerabilityKey(info).VulnerabilityString(), nil
	case apitypes.MalwareFindingInfo:
		return GenerateMalwareKey(info).MalwareString(), nil
	case apitypes.MisconfigurationFindingInfo:
		return GenerateMisconfigurationKey(info).MisconfigurationString(), nil
	case apitypes.RootkitFindingInfo:
		return GenerateRootkitKey(info).RootkitString(), nil
	case apitypes.SecretFindingInfo:
		return GenerateSecretKey(info).SecretString(), nil
	case apitypes.PackageFindingInfo:
		return GeneratePackageKey(info).PackageString(), nil
	default:
		return "", fmt.Errorf("unsupported finding info type %T", value)
	}
}
