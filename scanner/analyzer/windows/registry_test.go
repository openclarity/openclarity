// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package windows

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

// Windows 10 registry data was obtained from
// https://github.com/AndrewRathbun/VanillaWindowsRegistryHives. Older Windows
// distributions were tested locally and their registries not uploaded due to
// security reasons.
//
// TODO(ramizpolic): Add more test cases for other Windows versions. The testdata
// size will grow, so finding another way to fake the registry data could be
// useful to avoid size issues and manual per-distro registry creation/testing.

func TestRegistry(t *testing.T) {
	// from https://github.com/AndrewRathbun/VanillaWindowsRegistryHives/tree/d12ba60d8dd283a4a17b1a02295356a6bed093cf/Windows10/21H2/W10_21H2_Pro_20211012_19044.1288
	registryFilePath := "testdata/W10_21H2_Pro_20211012_19044.SOFTWARE"

	// when
	reg, err := NewRegistry(registryFilePath, log.NewEntry(&log.Logger{}))
	assert.NilError(t, err)

	bom, err := reg.GetBOM()
	assert.NilError(t, err)

	// check basic details
	assert.Equal(t, bom.SerialNumber, "urn:uuid:ec61c342-9f62-593b-8589-824ecc574a26")
	assert.Equal(t, bom.Metadata.Component.Name, "Windows 10 Pro")
	assert.Equal(t, bom.Metadata.Component.Version, "10.0.19044")

	// check apps and updates details
	hasApp := func(reqAppName string) cmp.Comparison {
		return func() cmp.Result {
			for _, app := range *bom.Components {
				if reqAppName == app.Name {
					return cmp.ResultSuccess
				}
			}
			return cmp.ResultFailure(fmt.Sprintf("BOM components do not contain app %q", reqAppName))
		}
	}

	assert.Assert(t, hasApp("KB5003242"))
	assert.Assert(t, hasApp("Microsoft Edge"))
}
