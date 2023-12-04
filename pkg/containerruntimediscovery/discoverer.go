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

package containerruntimediscovery

import (
	"context"
	"errors"
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
	"github.com/openclarity/vmclarity/pkg/shared/log"
)

type Discoverer interface {
	Images(ctx context.Context) ([]models.ContainerImageInfo, error)
	Containers(ctx context.Context) ([]models.ContainerInfo, error)
}

type DiscovererFactory func(ctx context.Context) (Discoverer, error)

var discovererFactories map[string]DiscovererFactory = map[string]DiscovererFactory{
	"docker":     NewDockerDiscoverer,
	"containerd": NewContainerdDiscoverer,
}

// NewDiscoverer tries to create all registered discoverers and returns the
// first that succeeds, if none of them succeed then we return all the errors
// for the caller to evalulate.
func NewDiscoverer(ctx context.Context) (Discoverer, error) {
	logger := log.GetLoggerFromContextOrDefault(ctx)
	errs := []error{}

	for name, factory := range discovererFactories {
		discoverer, err := factory(ctx)
		if err == nil {
			logger.Infof("Loaded %s discoverer", name)
			return discoverer, nil
		}
		errs = append(errs, fmt.Errorf("failed to create %s discoverer: %w", name, err))
	}
	return nil, errors.Join(errs...)
}
