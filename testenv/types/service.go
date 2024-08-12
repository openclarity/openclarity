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

package types

import (
	"fmt"
)

type Service interface {
	fmt.Stringer

	GetID() string
	GetNamespace() string
	GetApplicationName() string
	GetComponentName() string
	GetState() ServiceState
}

type ServiceState string

func (s ServiceState) String() string {
	switch s {
	case ServiceStateReady:
		return string(ServiceStateReady)
	case ServiceStateNotReady:
		return string(ServiceStateNotReady)
	case ServiceStateDegraded:
		return string(ServiceStateDegraded)
	case ServiceStateUnknown:
		fallthrough
	default:
		return string(ServiceStateUnknown)
	}
}

const (
	// Service is running and ready.
	ServiceStateReady ServiceState = "Ready"
	// Service is not running or running with errors.
	ServiceStateDegraded ServiceState = "Degraded"
	// Service is running but not ready yet.
	ServiceStateNotReady ServiceState = "NotReady"
	// Service state cannot be determined.
	ServiceStateUnknown ServiceState = "Unknown"
)

// Service is a collection of Service implementation.
type Services []Service

func (s Services) IDs() []string {
	ids := make([]string, len(s))
	for i, srv := range s {
		ids[i] = srv.GetID()
	}

	return ids
}

func (s Services) States() []string {
	states := make([]string, len(s))
	for i, srv := range s {
		states[i] = srv.GetState().String()
	}

	return states
}
