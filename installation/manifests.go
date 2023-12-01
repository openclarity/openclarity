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

package installation

import (
	"embed"

	"github.com/openclarity/vmclarity/pkg/shared/manifest"
)

//go:embed all:aws
var awsManifestFS embed.FS

var AWSManifestBundle = &manifest.Bundle{
	Prefix:  "aws",
	FS:      awsManifestFS,
	Matcher: manifest.DefaultMatcher,
}

//go:embed all:azure
var azureManifestFS embed.FS

var AzureManifestBundle = &manifest.Bundle{
	Prefix:  "azure",
	FS:      azureManifestFS,
	Matcher: manifest.DefaultMatcher,
}

//go:embed all:docker
var dockerManifestFS embed.FS

var DockerManifestBundle = &manifest.Bundle{
	Prefix:  "docker",
	FS:      dockerManifestFS,
	Matcher: manifest.DefaultMatcher,
}

//go:embed all:gcp
var gcpManifestFS embed.FS

var GCPManifestBundle = &manifest.Bundle{
	Prefix:  "gcp",
	FS:      gcpManifestFS,
	Matcher: manifest.DefaultMatcher,
}

//go:embed all:kubernetes/helm/vmclarity
var helmManifestFS embed.FS

var HelmManifestBundle = &manifest.Bundle{
	Prefix:  "kubernetes/helm/vmclarity",
	FS:      helmManifestFS,
	Matcher: manifest.DefaultMatcher,
}
