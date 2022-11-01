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

package trivy

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/commands/artifact"

	"github.com/openclarity/kubeclarity/shared/v2/pkg/utils"
)

func KubeclaritySourceToTrivySource(sourceType utils.SourceType) (artifact.TargetKind, error) {
	switch sourceType {
	case utils.IMAGE:
		return artifact.TargetContainerImage, nil
	case utils.ROOTFS:
		return artifact.TargetRootfs, nil
	case utils.DIR, utils.FILE:
		return artifact.TargetFilesystem, nil
	case utils.SBOM:
		return artifact.TargetSBOM, nil
	}
	return artifact.TargetKind("Unknown"), fmt.Errorf("unable to convert source type %v to trivy type", sourceType)
}
