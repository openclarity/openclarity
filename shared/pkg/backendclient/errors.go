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

package backendclient

import (
	"fmt"

	"github.com/openclarity/vmclarity/api/models"
)

type TargetConflictError struct {
	ConflictingTarget *models.Target
	Message           string
}

func (t TargetConflictError) Error() string {
	return fmt.Sprintf("Conflicting Target Found with ID %s: %s", *t.ConflictingTarget.Id, t.Message)
}

type ScanConflictError struct {
	ConflictingScan *models.Scan
	Message         string
}

func (t ScanConflictError) Error() string {
	return fmt.Sprintf("Conflicting Scan Found with ID %s: %s", *t.ConflictingScan.Id, t.Message)
}

type ScanResultConflictError struct {
	ConflictingScanResult *models.TargetScanResult
	Message               string
}

func (t ScanResultConflictError) Error() string {
	return fmt.Sprintf("Conflicting Scan Result Found with ID %s: %s", *t.ConflictingScanResult.Id, t.Message)
}
