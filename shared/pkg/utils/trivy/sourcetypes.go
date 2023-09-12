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
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquasecurity/trivy/pkg/commands/artifact"

	"github.com/openclarity/kubeclarity/shared/pkg/utils"
)

func KubeclaritySourceToTrivySource(sourceType utils.SourceType) (artifact.TargetKind, error) {
	switch sourceType {
	case utils.IMAGE:
		return artifact.TargetContainerImage, nil
	case utils.DOCKERARCHIVE, utils.OCIARCHIVE:
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

// nolint:cyclop
func UntarToDirectory(tarPath, destDirectory string) error {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("unable to open archive file: %w", err)
	}

	tr := tar.NewReader(tarFile)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		rel := filepath.FromSlash(hdr.Name)
		abs := filepath.Join(destDirectory, rel)
		mode := hdr.FileInfo().Mode()

		switch hdr.Typeflag {
		case tar.TypeReg:
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return fmt.Errorf("unable to open file %s for writing: %w", abs, err)
			}

			n, err := io.CopyN(wf, tr, hdr.Size)
			if err != nil {
				return fmt.Errorf("error writing data to file %s: %w", abs, err)
			}

			err = wf.Close()
			if err != nil {
				return fmt.Errorf("error closing file %s: %w", abs, err)
			}

			if n != hdr.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, hdr.Size)
			}
		case tar.TypeDir:
			// nolint:gomnd
			err := os.MkdirAll(abs, 0o755)
			if err != nil {
				return fmt.Errorf("unable to create directory %s: %w", abs, err)
			}
		case tar.TypeXGlobalHeader:
			// ignore
		default:
			return fmt.Errorf("file entry %s contained unsupported file type", abs)
		}
	}

	return nil
}
