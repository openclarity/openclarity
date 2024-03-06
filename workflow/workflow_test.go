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

package workflow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	gtypes "github.com/onsi/gomega/types"

	"github.com/openclarity/vmclarity/workflow/types"
)

const (
	testTaskID1 string = "test-task-01"
	testTaskID2 string = "test-task-02"
	testTaskID3 string = "test-task-03"
	testTaskID4 string = "test-task-04"
	testTaskID5 string = "test-task-05"
)

type TestState struct {
	order []string

	mu *sync.RWMutex
}

func (s *TestState) Add(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.order = append(s.order, id)
}

func NewTestState() *TestState {
	return &TestState{
		order: make([]string, 0),
		mu:    &sync.RWMutex{},
	}
}

//nolint:maintidx
func TestWorkflow(t *testing.T) {
	tests := []struct {
		Name    string
		Tasks   []*types.Task[*TestState]
		State   *TestState
		Timeout time.Duration

		ExpectedErrorMatcher gtypes.GomegaMatcher
		ExpectedOrderMatcher gtypes.GomegaMatcher
	}{
		{
			Name: "Linear workflow",
			Tasks: []*types.Task[*TestState]{
				{
					Name: testTaskID1,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID1)
						return nil
					},
				},
				{
					Name: testTaskID2,
					Deps: []string{testTaskID1},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID2)
						return nil
					},
				},
				{
					Name: testTaskID3,
					Deps: []string{testTaskID2},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID3)
						return nil
					},
				},
				{
					Name: testTaskID4,
					Deps: []string{testTaskID3},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID4)
						return nil
					},
				},
				{
					Name: testTaskID5,
					Deps: []string{testTaskID4},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID5)
						return nil
					},
				},
			},
			State:                NewTestState(),
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedOrderMatcher: BeEquivalentTo(
				[]string{testTaskID1, testTaskID2, testTaskID3, testTaskID4, testTaskID5},
			),
		},
		{
			Name: "Parallel workflow",
			Tasks: []*types.Task[*TestState]{
				{
					Name: testTaskID1,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID1)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID2,
					Deps: []string{testTaskID1},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID2)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID3,
					Deps: []string{testTaskID2},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID3)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID4,
					Deps: []string{testTaskID5, testTaskID3},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID4)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID5,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID5)
						time.Sleep(4 * time.Second)
						return nil
					},
				},
			},
			State:                NewTestState(),
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedOrderMatcher: SatisfyAny(
				BeEquivalentTo([]string{testTaskID1, testTaskID5, testTaskID2, testTaskID3, testTaskID4}),
				BeEquivalentTo([]string{testTaskID5, testTaskID1, testTaskID2, testTaskID3, testTaskID4}),
			),
		},
		{
			Name: "Parallel workflow with error",
			Tasks: []*types.Task[*TestState]{
				{
					Name: testTaskID1,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID1)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID2,
					Deps: []string{testTaskID1},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID2)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID3,
					Deps: []string{testTaskID2},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID3)
						time.Sleep(time.Second)
						return errors.New(testTaskID4)
					},
				},
				{
					Name: testTaskID4,
					Deps: []string{testTaskID5, testTaskID3},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID4)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID5,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID5)
						time.Sleep(4 * time.Second)
						return nil
					},
				},
			},
			State:                NewTestState(),
			ExpectedErrorMatcher: HaveOccurred(),
			ExpectedOrderMatcher: SatisfyAny(
				BeEquivalentTo([]string{testTaskID1, testTaskID5, testTaskID2, testTaskID3}),
				BeEquivalentTo([]string{testTaskID5, testTaskID1, testTaskID2, testTaskID3}),
			),
		},
		{
			Name: "Parallel workflow with timeout",
			Tasks: []*types.Task[*TestState]{
				{
					Name: testTaskID1,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID1)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID2,
					Deps: []string{testTaskID1},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID2)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID3,
					Deps: []string{testTaskID2},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID3)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID4,
					Deps: []string{testTaskID5, testTaskID3},
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID4)
						time.Sleep(time.Second)
						return nil
					},
				},
				{
					Name: testTaskID5,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID5)
						time.Sleep(5 * time.Second)
						return nil
					},
				},
			},
			State:                NewTestState(),
			Timeout:              4 * time.Second,
			ExpectedErrorMatcher: HaveOccurred(),
			ExpectedOrderMatcher: SatisfyAny(
				BeEquivalentTo([]string{testTaskID1, testTaskID5, testTaskID2, testTaskID3}),
				BeEquivalentTo([]string{testTaskID5, testTaskID1, testTaskID2, testTaskID3}),
			),
		},
		{
			Name: "No dependency workflow",
			Tasks: []*types.Task[*TestState]{
				{
					Name: testTaskID1,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID1)
						return nil
					},
				},
				{
					Name: testTaskID2,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID2)
						return nil
					},
				},
				{
					Name: testTaskID3,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID3)
						return nil
					},
				},
				{
					Name: testTaskID4,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID4)
						return nil
					},
				},
				{
					Name: testTaskID5,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID5)
						return nil
					},
				},
			},
			State:                NewTestState(),
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedOrderMatcher: ContainElements(
				[]string{testTaskID1, testTaskID2, testTaskID3, testTaskID4, testTaskID5},
			),
		},
		{
			Name: "Single task workflow",
			Tasks: []*types.Task[*TestState]{
				{
					Name: testTaskID1,
					Deps: nil,
					Fn: func(ctx context.Context, state *TestState) error {
						state.Add(testTaskID1)
						return nil
					},
				},
			},
			State:                NewTestState(),
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedOrderMatcher: BeEquivalentTo([]string{testTaskID1}),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			ctx := context.Background()
			var cancel context.CancelFunc
			if test.Timeout > 0 {
				ctx, cancel = context.WithTimeout(ctx, test.Timeout)
				defer cancel()
			}

			w, err := New[*TestState, *types.Task[*TestState]](test.Tasks)
			g.Expect(err).ShouldNot(HaveOccurred())

			err = w.Run(ctx, test.State)
			g.Expect(err).Should(test.ExpectedErrorMatcher)

			g.Expect(test.State.order).Should(test.ExpectedOrderMatcher)
		})
	}
}
