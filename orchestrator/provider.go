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

package orchestrator

import (
	"context"
	"fmt"

	apitypes "github.com/openclarity/vmclarity/api/types"
	"github.com/openclarity/vmclarity/provider"
	"github.com/openclarity/vmclarity/provider/aws"
	"github.com/openclarity/vmclarity/provider/azure"
	"github.com/openclarity/vmclarity/provider/docker"
	"github.com/openclarity/vmclarity/provider/external"
	"github.com/openclarity/vmclarity/provider/gcp"
	"github.com/openclarity/vmclarity/provider/kubernetes"
)

// nolint:wrapcheck
// NewProvider returns an initialized provider.Provider based on the kind apitypes.CloudProvider.
func NewProvider(ctx context.Context, kind apitypes.CloudProvider) (provider.Provider, error) {
	switch kind {
	case apitypes.Azure:
		return azure.New(ctx)
	case apitypes.Docker:
		return docker.New(ctx)
	case apitypes.AWS:
		return aws.New(ctx)
	case apitypes.GCP:
		return gcp.New(ctx)
	case apitypes.External:
		return external.New(ctx)
	case apitypes.Kubernetes:
		return kubernetes.New(ctx)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", kind)
	}
}
