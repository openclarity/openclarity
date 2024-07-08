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
	"github.com/aquasecurity/trivy/pkg/flag"
	"github.com/google/go-containerregistry/pkg/name"
)

func GetTrivyDBOptions() (flag.DBOptions, error) {
	dbRepository, err := name.ParseReference(flag.DBRepositoryFlag.Default, name.WithDefaultTag(""))
	if err != nil {
		return flag.DBOptions{}, err
	}
	javaRepository, err := name.ParseReference(flag.JavaDBRepositoryFlag.Default, name.WithDefaultTag(""))
	if err != nil {
		return flag.DBOptions{}, err
	}
	return flag.DBOptions{
		DBRepository:     dbRepository,   // Use the default trivy source for the vuln DB
		JavaDBRepository: javaRepository, // Use the default trivy source for the java DB
		NoProgress:       true,           // Disable the interactive progress bar
	}, nil
}
