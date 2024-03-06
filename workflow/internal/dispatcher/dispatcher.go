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
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/openclarity/vmclarity/workflow/types"
)

type Dispatcher[S any, R types.Runnable[S]] struct {
	m  map[string]TaskStatus
	mu *sync.RWMutex

	stats *Stats

	exitChan chan struct{}
	nextChan chan struct{}
}

func (d *Dispatcher[S, R]) get(id string) (TaskStatus, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	s, ok := d.m[id]
	if !ok {
		return TaskStatus{}, fmt.Errorf("unknown task with ID: %s", id)
	}

	return s, nil
}

func (d *Dispatcher[S, R]) set(id string, status *TaskStatus) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.m[id] = *status
	d.stats.Update(status)
}

func (d *Dispatcher[S, R]) update(ctx context.Context, id string, status TaskStatus) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	d.set(id, &status)

	if status.Finished() {
		d.mu.RLock()
		defer d.mu.RUnlock()

		select {
		case <-d.exitChan:
			return
		default:
			select {
			case d.nextChan <- struct{}{}:
				return
			default:
			}
		}
	}
}

func (d *Dispatcher[S, R]) stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	close(d.exitChan)
	close(d.nextChan)
}

func (d *Dispatcher[S, R]) Next() <-chan struct{} {
	return d.nextChan
}

func (d *Dispatcher[S, R]) Dispatch(ctx context.Context, task R, state S) (bool, error) {
	id := task.ID()

	taskStatus, err := d.get(id)
	if err != nil {
		return false, fmt.Errorf("failed to get status for task with ID: %s", id)
	}

	if taskStatus.State != PENDING {
		return true, nil
	}

	for _, depID := range task.Dependencies() {
		depStatus, err := d.get(depID)
		if err != nil {
			return false, fmt.Errorf("failed to get status for task with ID: %s", id)
		}
		if !depStatus.Finished() {
			return false, nil
		}
	}

	taskStatus.State = RUNNING
	d.update(ctx, id, taskStatus)

	go func() {
		err = task.Run(ctx, state)
		d.update(ctx, id, taskStatus.FromError(err))
	}()

	return true, nil
}

func (d *Dispatcher[S, R]) Finished() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.stats.failed > 0 || d.stats.InProgress() == 0 {
		return true
	}

	return false
}

func (d *Dispatcher[S, R]) Result() error {
	d.mu.RLock()
	defer d.mu.RUnlock()

	errs := make([]error, 0, len(d.m))
	for _, status := range d.m {
		errs = append(errs, status.Err)
	}

	return errors.Join(errs...)
}

func (d *Dispatcher[S, R]) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var s string
	for id, state := range d.m {
		s += fmt.Sprintf("ID: %s, %s\n", id, state.String())
	}

	return s
}

func New[S any, R types.Runnable[S]](order []string) (*Dispatcher[S, R], func()) {
	d := &Dispatcher[S, R]{
		m:        make(map[string]TaskStatus, len(order)),
		mu:       &sync.RWMutex{},
		stats:    &Stats{},
		exitChan: make(chan struct{}),
		nextChan: make(chan struct{}, 1),
	}

	for _, id := range order {
		status := TaskStatus{}
		d.m[id] = status
		d.stats.Update(&status)
	}

	return d, d.stop
}
