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

package containerrootfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/file"
	"github.com/anchore/stereoscope/pkg/filetree/filenode"
	"github.com/anchore/stereoscope/pkg/image"

	"github.com/openclarity/vmclarity/core/log"
)

const perFileReadLimit = 2 * file.GB

func GetImageWithCleanup(ctx context.Context, src string) (*image.Image, func(), error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	image, err := stereoscope.GetImage(ctx, src)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse image from src %s: %w", src, err)
	}

	cleanup := func() {
		err := image.Cleanup()
		if err != nil {
			logger.WithError(err).Error("unable to clean up image")
		}
	}

	return image, cleanup, nil
}

// nolint:cyclop, gocognit
func ToDirectory(ctx context.Context, image *image.Image, dest string) error {
	logger := log.GetLoggerFromContextOrDefault(ctx)

	err := image.SquashedTree().Walk(func(path file.Path, f filenode.FileNode) error {
		target := filepath.Join(dest, string(path))

		switch f.FileType {
		case file.TypeDirectory:
			// nolint:mnd
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("unable to make directory %s: %w", path, err)
			}
		case file.TypeRegular:
			indexEntry, err := image.FileCatalog.Get(*f.Reference)
			if err != nil {
				return fmt.Errorf("unable to get index entry for file: %w", err)
			}

			output, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, indexEntry.Mode())
			if err != nil {
				return fmt.Errorf("unable to open file %s: %w", target, err)
			}
			defer func() {
				err = output.Close()
				if err != nil {
					logger.WithError(err).Errorf("unable to close file %s", target)
				}
			}()

			content, err := image.OpenReference(*f.Reference)
			if err != nil {
				return fmt.Errorf("unable to open path from squash: %w", err)
			}

			numBytes, err := io.Copy(output, io.LimitReader(content, perFileReadLimit))
			if numBytes >= perFileReadLimit || errors.Is(err, io.EOF) {
				return errors.New("zip read limit hit (potential decompression bomb attack)")
			}
			if err != nil {
				return fmt.Errorf("unable to copy file: %w", err)
			}
		case file.TypeSymLink:
			linkTarget := string(f.LinkPath)
			if f.LinkPath.IsAbsolutePath() {
				linkTarget = filepath.Join(dest, string(f.LinkPath))
			}

			err := os.Symlink(linkTarget, target)
			if err != nil {
				return fmt.Errorf("unable to create symlink: %w", err)
			}
		case file.TypeHardLink, file.TypeCharacterDevice, file.TypeBlockDevice, file.TypeFIFO, file.TypeSocket, file.TypeIrregular:
			logger.Warnf("found unsupported file type %s in container image at %s", f.FileType, path)
		}

		return nil
	}, nil)
	if err != nil {
		return fmt.Errorf("unable to walk squashed tree from container: %w", err)
	}

	return nil
}
