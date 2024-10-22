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
	"os"

	"github.com/openclarity/openclarity/scanner/common"

	stereoscopeFile "github.com/anchore/stereoscope/pkg/file"
	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	"github.com/aquasecurity/trivy/pkg/fanal/types"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	log "github.com/sirupsen/logrus"
)

func SourceToTrivySource(sourceType common.InputType) (artifact.TargetKind, error) {
	switch sourceType {
	case common.IMAGE:
		return artifact.TargetContainerImage, nil
	case common.ROOTFS:
		return artifact.TargetRootfs, nil
	case common.DIR, common.FILE, common.CSV:
		return artifact.TargetFilesystem, nil
	case common.SBOM:
		return artifact.TargetSBOM, nil
	case common.DOCKERARCHIVE, common.OCIARCHIVE, common.OCIDIR:
	default:
	}
	return "Unknown", fmt.Errorf("unable to convert source type %v to trivy type", sourceType)
}

type CleanupFunc func(log *log.Entry)

func UntarToTempDirectory(tar string) (string, CleanupFunc, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", func(_ *log.Entry) {}, fmt.Errorf("unable to create temp directory: %w", err)
	}
	cleanup := func(log *log.Entry) {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.WithError(err).Errorf("unable to remove temp directory %s", tmpDir)
		}
	}

	f, err := os.Open(tar)
	if err != nil {
		return "", cleanup, fmt.Errorf("unable to untar %s to temp directory: %w", tar, err)
	}
	err = stereoscopeFile.UntarToDirectory(f, tmpDir)
	if err != nil {
		return "", cleanup, fmt.Errorf("unable to untar %s to temp directory: %w", tar, err)
	}
	return tmpDir, cleanup, nil
}

func SetTrivyImageOptions(sourceType common.InputType, userInput string, trivyOptions trivyFlag.Options) (trivyFlag.Options, CleanupFunc, error) {
	trivyOptions.ImageOptions = trivyFlag.ImageOptions{
		ImageSources: types.AllImageSources,
	}

	cleanup := func(_ *log.Entry) {}
	switch sourceType {
	// Docker Archive and OCI directories are natively supported by Trivy
	// just needs to set the ImageOptions Input to the tar/directory.
	case common.DOCKERARCHIVE, common.OCIDIR:
		trivyOptions.ImageOptions.Input = userInput

	// OCI Archive isn't natively supported, so we'll convert it to an OCI
	// directory first and then configure trivy as above.
	case common.OCIARCHIVE:
		var err error
		var tmpDir string
		tmpDir, cleanup, err = UntarToTempDirectory(userInput)
		if err != nil {
			return trivyOptions, cleanup, fmt.Errorf("unable to untar %s to temp directory: %w", userInput, err)
		}
		trivyOptions.ImageOptions.Input = tmpDir

	case common.IMAGE, common.ROOTFS, common.DIR, common.FILE, common.SBOM, common.CSV:
		// Nothing to do here, setting the target in ScanOptions is
		// enough.
	}

	return trivyOptions, cleanup, nil
}
