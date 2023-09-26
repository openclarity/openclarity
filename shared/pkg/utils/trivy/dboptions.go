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

	"github.com/aquasecurity/trivy/pkg/flag"
)

func GetTrivyDBOptions() (flag.DBOptions, error) {
	// Get the Trivy CVE DB URL default value from the trivy
	// configuration, we may want to make this configurable in the
	// future.
	dbRepoDefaultValue, ok := flag.DBRepositoryFlag.Default.(string)
	if !ok {
		return flag.DBOptions{}, fmt.Errorf("unable to get trivy DB repo config")
	}

	// Get the Trivy JAVA DB URL default value from the trivy
	// configuration, we may want to make this configurable in the
	// future.
	javaDBRepoDefaultValue, ok := flag.JavaDBRepositoryFlag.Default.(string)
	if !ok {
		return flag.DBOptions{}, fmt.Errorf("unable to get trivy java DB repo config")
	}

	return flag.DBOptions{
		DBRepository:     dbRepoDefaultValue,     // Use the default trivy source for the vuln DB
		JavaDBRepository: javaDBRepoDefaultValue, // Use the default trivy source for the java DB
		NoProgress:       true,                   // Disable the interactive progress bar
	}, nil
}
