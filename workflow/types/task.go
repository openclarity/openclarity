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
	"context"
)

var _ Runnable[any] = &Task[any]{}

type Task[S any] struct {
	Name string
	Deps []string
	Fn   func(context.Context, S) error
}

func (t Task[S]) Dependencies() []string {
	return t.Deps
}

func (t Task[S]) Run(ctx context.Context, state S) error {
	if t.Fn != nil {
		return t.Fn(ctx, state)
	}

	return nil
}

func (t Task[S]) ID() string {
	return t.Name
}
