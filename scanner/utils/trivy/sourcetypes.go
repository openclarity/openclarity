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

	stereoscopeFile "github.com/anchore/stereoscope/pkg/file"
	"github.com/aquasecurity/trivy/pkg/commands/artifact"
	"github.com/aquasecurity/trivy/pkg/fanal/types"
	trivyFlag "github.com/aquasecurity/trivy/pkg/flag"
	log "github.com/sirupsen/logrus"

	"github.com/openclarity/vmclarity/scanner/utils"
)

func SourceToTrivySource(sourceType utils.SourceType) (artifact.TargetKind, error) {
	switch sourceType {
	case utils.IMAGE:
		return artifact.TargetContainerImage, nil
	case utils.DOCKERARCHIVE, utils.OCIARCHIVE, utils.OCIDIR:
		return artifact.TargetImageArchive, nil
	case utils.ROOTFS:
		return artifact.TargetRootfs, nil
	case utils.DIR, utils.FILE:
		return artifact.TargetFilesystem, nil
	case utils.SBOM:
		return artifact.TargetSBOM, nil
	}
	return artifact.TargetKind("Unknown"), fmt.Errorf("unable to convert source type %v to trivy type", sourceType)
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

func SetTrivyImageOptions(sourceType utils.SourceType, userInput string, trivyOptions trivyFlag.Options) (trivyFlag.Options, CleanupFunc, error) {
	trivyOptions.ImageOptions = trivyFlag.ImageOptions{
		ImageSources: types.AllImageSources,
	}

	cleanup := func(_ *log.Entry) {}
	switch sourceType {
	// Docker Archive and OCI directories are natively supported by Trivy
	// just needs to set the ImageOptions Input to the tar/directory.
	case utils.DOCKERARCHIVE, utils.OCIDIR:
		trivyOptions.ImageOptions.Input = userInput

	// OCI Archive isn't natively supported, so we'll convert it to an OCI
	// directory first and then configure trivy as above.
	case utils.OCIARCHIVE:
		var err error
		var tmpDir string
		tmpDir, cleanup, err = UntarToTempDirectory(userInput)
		if err != nil {
			return trivyOptions, cleanup, fmt.Errorf("unable to untar %s to temp directory: %w", userInput, err)
		}
		trivyOptions.ImageOptions.Input = tmpDir

	case utils.IMAGE, utils.ROOTFS, utils.DIR, utils.FILE, utils.SBOM:
		// Nothing to do here, setting the target in ScanOptions is
		// enough.
	}

	return trivyOptions, cleanup, nil
}
