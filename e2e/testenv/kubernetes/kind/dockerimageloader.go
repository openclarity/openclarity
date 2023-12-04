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

package kind

import (
	"context"
	"fmt"
	"io"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"

	"github.com/openclarity/vmclarity/e2e/testenv/utils"
	"github.com/openclarity/vmclarity/e2e/testenv/utils/fanout"
)

const (
	ImageLoaderTypeDocker ImageLoaderType = "docker"
)

type DockerImageLoader struct {
	docker *dockerclient.Client
	images []string
}

func (l *DockerImageLoader) imageIDsFromRepoTags(ctx context.Context, repoTags []string) ([]string, error) {
	imageIDs := make([]string, 0, len(l.images))
	for _, repoTag := range repoTags {
		result, err := l.docker.ImageList(ctx, dockertypes.ImageListOptions{
			Filters: filters.NewArgs(
				filters.Arg("reference", repoTag)),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list image %s: %w", repoTag, err)
		}

		if len(result) <= 0 {
			return nil, fmt.Errorf("failed to find image: %s", repoTag)
		}

		imageIDs = append(imageIDs, result[0].ID)
	}

	return imageIDs, nil
}

func (l *DockerImageLoader) Load(ctx context.Context, nodes []nodes.Node) error {
	logger := utils.GetLoggerFromContextOrDiscard(ctx).WithFields(map[string]interface{}{
		"images": l.images,
		"nodes":  nodes,
	})

	if len(nodes) == 0 {
		logger.Warn("skipping loading container images as no target nodes provided")
		return nil
	}

	logger.Debug("loading images to nodes")

	// Get image IDs for Image RepoTags
	mapping := make(imageIDRepoTagMapping, 0)
	for _, image := range l.images {
		ids, err := l.imageIDsFromRepoTags(ctx, []string{image})
		if err != nil || len(ids) == 0 {
			return fmt.Errorf("failed to get ImageID for RepoTag %s: %w", image, err)
		}

		mapping[ids[0]] = image
	}

	imageData, err := l.docker.ImageSave(ctx, mapping.IDs())
	if err != nil {
		return fmt.Errorf("failed to save images from local Docker: %w", err)
	}
	defer func(imagesTar io.ReadCloser) {
		err = imagesTar.Close()
		if err != nil {
			logger.Error("failed to close container image(s) archive data stream")
		}
	}(imageData)

	nodeLoaders := []func(r io.Reader) error{}
	for _, node := range nodes {
		nodeLoaders = append(nodeLoaders, newNodeLoader(node, mapping))
	}

	if err = fanout.FanOut(ctx, imageData, nodeLoaders); err != nil {
		return fmt.Errorf("failed to load image(s): %w", err)
	}

	return nil
}

func (l DockerImageLoader) Type() ImageLoaderType {
	return ImageLoaderTypeDocker
}

func (l DockerImageLoader) String() string {
	return string(ImageLoaderTypeDocker)
}

func NewDockerImageLoader(images []string, timeout time.Duration) (ImageLoader, error) {
	docker, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithTimeout(timeout),
		dockerclient.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	return &DockerImageLoader{docker, images}, nil
}

type imageIDRepoTagMapping map[string]string

func (m imageIDRepoTagMapping) IDs() []string {
	if m == nil {
		return nil
	}

	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}

	return ids
}

type nodeLoaderFn func(r io.Reader) error

func newNodeLoader(node nodes.Node, mapping imageIDRepoTagMapping) nodeLoaderFn {
	return func(r io.Reader) error {
		if err := nodeutils.LoadImageArchive(node, r); err != nil {
			return fmt.Errorf("failed to load image from stream: %w", err)
		}

		for imageID, repoTag := range mapping {
			tags, err := nodeutils.ImageTags(node, imageID)
			if err != nil {
				return fmt.Errorf("failed to get repoTags for imageID %s: %w", imageID, err)
			}

			exists := false
			for tag := range tags {
				if tag == repoTag {
					exists = true
				}
			}
			if exists {
				continue
			}

			if err = nodeutils.ReTagImage(node, imageID, repoTag); err != nil {
				return fmt.Errorf("failed to set %s tag for %s imageID: %w", repoTag, imageID, err)
			}
		}

		return nil
	}
}
