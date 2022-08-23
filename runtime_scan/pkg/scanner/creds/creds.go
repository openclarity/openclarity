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

package creds

import (
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
)

type CredentialAdder interface {
	// ShouldAdd returns true if credentials should be added to a scanner job
	ShouldAdd() bool
	// Add adds credentials to a scanner job
	Add(job *batchv1.Job)
	// GetNamespace gets credential namespace
	GetNamespace() string
}

// nolint:structcheck
type credsCommon struct {
	clientset       kubernetes.Interface
	isSecretExists  *bool
	secretNamespace string
}
