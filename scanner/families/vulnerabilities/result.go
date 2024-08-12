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

package vulnerabilities

import (
	"errors"
	"fmt"

	apitypes "github.com/openclarity/vmclarity/api/types"

	"github.com/openclarity/vmclarity/scanner/families/types"
	"github.com/openclarity/vmclarity/scanner/scanner"
	"github.com/openclarity/vmclarity/scanner/utils/image_helper"
)

type Results struct {
	Metadata      types.Metadata
	MergedResults *scanner.MergedResults
}

func (*Results) IsResults() {}

func (r *Results) GetSourceImageInfo() (*apitypes.ContainerImageInfo, error) {
	if r.MergedResults == nil {
		return nil, errors.New("missing merged results")
	}

	sourceImage := image_helper.ImageInfo{}
	if err := sourceImage.FromMetadata(r.MergedResults.Source.Metadata); err != nil {
		return nil, fmt.Errorf("failed to load source image from metadata: %w", err)
	}

	containerImageInfo, err := sourceImage.ToContainerImageInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to convert container image: %w", err)
	}

	return containerImageInfo, nil
}
