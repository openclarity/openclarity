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

package dispatcher

import (
	"fmt"
)

type TaskState int8

const (
	PENDING TaskState = iota
	RUNNING
	DONE
	FAILED
)

func (s TaskState) String() string {
	switch s {
	case PENDING:
		return "Pending"
	case RUNNING:
		return "Running"
	case DONE:
		return "Done"
	case FAILED:
		return "Failed"
	default:
		return "Unknown"
	}
}

type TaskStatus struct {
	State TaskState
	Err   error
}

func (s TaskStatus) Finished() bool {
	switch s.State {
	case DONE, FAILED:
		return true
	case PENDING, RUNNING:
		fallthrough
	default:
		return false
	}
}

func (s TaskStatus) FromError(err error) TaskStatus {
	s.Err = err
	if s.Err != nil {
		s.State = FAILED
	} else {
		s.State = DONE
	}

	return s
}

func (s TaskStatus) String() string {
	var err string
	if s.Err != nil {
		err = s.Err.Error()
	}

	return fmt.Sprintf("State: %s Error: %s", s.State, err)
}
