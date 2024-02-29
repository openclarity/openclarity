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

type Stats struct {
	pending int
	running int
	failed  int
	done    int
}

func (s *Stats) Update(status *TaskStatus) {
	switch status.State {
	case PENDING:
		s.pending += 1
	case RUNNING:
		s.running += 1
		s.pending -= 1
	case DONE:
		s.done += 1
		s.pending -= 1
	case FAILED:
		s.failed += 1
		s.pending -= 1
	default:
		return
	}
}

func (s *Stats) Failed() int {
	return s.failed
}

func (s *Stats) InProgress() int {
	return s.pending + s.running
}
