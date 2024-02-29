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
	"fmt"

	"github.com/heimdalr/dag"

	"github.com/openclarity/vmclarity/workflow/internal/dispatcher"
	"github.com/openclarity/vmclarity/workflow/types"
)

type Workflow[S any, R types.Runnable[S]] struct {
	tasks map[string]R
	order []string
}

func (w *Workflow[S, R]) Run(ctx context.Context, state S) error {
	d, stop := dispatcher.New[S, R](w.order)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	from := 0
	for !d.Finished() {
		var err error

		for idx, taskID := range w.order[from:] {
			task := w.tasks[taskID]

			var ok bool
			ok, err = d.Dispatch(ctx, task, state)
			if err != nil {
				return fmt.Errorf("failed to dispatch task: %s", taskID)
			}
			if !ok {
				break
			}
			from = idx
		}

		var next bool
		for !next {
			select {
			case <-ctx.Done():
				return fmt.Errorf("failed to run workflow: %w", ctx.Err())
			case <-d.Next():
				next = true
				break
			default:
				if d.Finished() {
					next = true
					break
				}
			}
		}
	}

	//nolint:wrapcheck
	return d.Result()
}

func New[S any, R types.Runnable[S]](runnables []R) (*Workflow[S, R], error) {
	tasks := make(map[string]R, len(runnables))
	graph := dag.NewDAG()

	for _, runnable := range runnables {
		id := runnable.ID()
		tasks[id] = runnable

		if err := graph.AddVertexByID(id, id); err != nil {
			return nil, fmt.Errorf("failed to add task with %s ID to graph: %w", id, err)
		}
	}

	for _, runnable := range runnables {
		id := runnable.ID()

		for _, dep := range runnable.Dependencies() {
			if err := graph.AddEdge(dep, id); err != nil {
				return nil, fmt.Errorf("failed to add for task with %s ID to graph: %w", id, err)
			}
		}
	}

	v := &visitor{
		Order: make([]string, 0, graph.GetOrder()),
	}

	graph.OrderedWalk(v)

	return &Workflow[S, R]{
		tasks: tasks,
		order: v.Order,
	}, nil
}

type visitor struct {
	Order []string
}

func (v *visitor) Visit(vertex dag.Vertexer) {
	if v.Order == nil {
		v.Order = make([]string, 0)
	}
	id, _ := vertex.Vertex()
	v.Order = append(v.Order, id)
}
