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

package common

import (
	"fmt"
)

type ConflictError struct {
	Reason string
}

type BadRequestError struct {
	Reason string
}

func (ec *ConflictError) Error() string {
	return fmt.Sprintf("Unable to create due to conflict, %v", ec.Reason)
}

func (bre *BadRequestError) Error() string {
	return fmt.Sprintf("Object validation failed: %v", bre.Reason)
}
