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
	"errors"
	"fmt"

	"github.com/aquasecurity/trivy/pkg/flag"
	"github.com/google/go-containerregistry/pkg/name"
)

func GetTrivyDBOptions() (flag.DBOptions, error) {
	// Get the Trivy CVE DB URL default value from the trivy
	// configuration, we may want to make this configurable in the
	// future.
	dbRepoDefaultValue := flag.DBRepositoryFlag.Default
	if len(dbRepoDefaultValue) == 0 {
		return flag.DBOptions{}, errors.New("unable to get trivy DB repo config")
	}

	dbRepositories := []name.Reference{}
	for _, value := range dbRepoDefaultValue {
		dbRepository, err := name.ParseReference(value, name.WithDefaultTag(""))
		if err != nil {
			return flag.DBOptions{}, fmt.Errorf("invalid db repository: %w", err)
		}
		dbRepositories = append(dbRepositories, dbRepository)
	}

	// Get the Trivy JAVA DB URL default value from the trivy
	// configuration, we may want to make this configurable in the
	// future.
	javaDBRepoDefaultValue := flag.JavaDBRepositoryFlag.Default
	if len(javaDBRepoDefaultValue) == 0 {
		return flag.DBOptions{}, errors.New("unable to get trivy java DB repo config")
	}

	javaDBRepositories := []name.Reference{}
	for _, value := range javaDBRepoDefaultValue {
		javaDBRepository, err := name.ParseReference(value, name.WithDefaultTag(""))
		if err != nil {
			return flag.DBOptions{}, fmt.Errorf("invalid db repository: %w", err)
		}
		javaDBRepositories = append(javaDBRepositories, javaDBRepository)
	}

	return flag.DBOptions{
		DBRepositories:     dbRepositories,     // Use the default trivy source for the vuln DB
		JavaDBRepositories: javaDBRepositories, // Use the default trivy source for the java DB
		NoProgress:         true,               // Disable the interactive progress bar
	}, nil
}
