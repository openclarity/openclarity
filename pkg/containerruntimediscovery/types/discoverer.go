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

package types

import (
	"context"
	"io"

	"github.com/openclarity/vmclarity/api/models"
)

type DiscovererFactory func() (Discoverer, error)

type Discoverer interface {
	Images(ctx context.Context) ([]models.ContainerImageInfo, error)
	Image(ctx context.Context, imageID string) (models.ContainerImageInfo, error)
	ExportImage(ctx context.Context, imageID string) (io.ReadCloser, error)

	Containers(ctx context.Context) ([]models.ContainerInfo, error)
	Container(ctx context.Context, containerID string) (models.ContainerInfo, error)
	ExportContainer(ctx context.Context, containerID string) (io.ReadCloser, func(), error)

	Ready(ctx context.Context) (bool, error)
}
