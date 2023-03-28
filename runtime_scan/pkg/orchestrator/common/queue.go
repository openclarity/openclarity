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
	"context"
	"fmt"
	"sync"
)

type Enqueuer[T comparable] interface {
	Enqueue(item T)
}

type Dequeuer[T comparable] interface {
	Dequeue(ctx context.Context) (T, error)
}

type Queue[T comparable] struct {
	// Channel used internally to block Dequeue when the queue is empty,
	// and notify Dequeue when a new item is added through Enqueue.
	itemAdded chan struct{}

	// A slice which represents the queue, Enqueue will add to the end of
	// the slice, Dequeue will remove from the head of the slice.
	queue []T

	// A map used as a set of unique items which are in the queue. This is
	// used by Enqueue and Has to provide a quick reference to whats in the
	// queue without needing to loop through the queue slice.
	inqueue map[T]struct{}

	// A mutex lock which protects the queue from simultaneous reads and
	// writes ensuring the queue can be used by multiple go routines safely.
	l sync.Mutex
}

func NewQueue[T comparable]() *Queue[T] {
	return &Queue[T]{
		itemAdded: make(chan struct{}),
		queue:     make([]T, 0),
		inqueue:   map[T]struct{}{},
	}
}

// Dequeue until it can dequeue an item from the queue or the passed
// context is cancelled.
func (q *Queue[T]) Dequeue(ctx context.Context) (T, error) {
	if len(q.queue) == 0 {
		// If the queue is empty, block waiting for the itemAdded
		// notification or context timeout.
		select {
		case <-q.itemAdded:
			// continue
		case <-ctx.Done():
			var empty T
			return empty, fmt.Errorf("failed to get item: %w", ctx.Err())
		}
	} else {
		// If the queue isn't empty, consume any item added notification
		// so that its reset for the empty case
		select {
		case <-q.itemAdded:
		default:
		}
	}

	q.l.Lock()
	defer q.l.Unlock()

	item := q.queue[0]
	q.queue = q.queue[1:]
	delete(q.inqueue, item)

	return item, nil
}

// Enqueue will add item to the queue if its not in the queue already.
func (q *Queue[T]) Enqueue(item T) {
	q.l.Lock()
	defer q.l.Unlock()

	if _, ok := q.inqueue[item]; !ok {
		q.queue = append(q.queue, item)
		q.inqueue[item] = struct{}{}
	}

	select {
	case q.itemAdded <- struct{}{}:
	default:
	}
}

// Length returns the current length of the queue. It should not be used to
// gate calls to Dequeue as the lock will be released in-between so the queue
// could change.
func (q *Queue[T]) Length() int {
	q.l.Lock()
	defer q.l.Unlock()

	return len(q.queue)
}

// Has returns true if items is currently in the queue for informational
// purposes, it should not be used to gate Dequeue or Enqueue as the lock will
// be released in-between and so the queue could change.
func (q *Queue[T]) Has(item T) bool {
	q.l.Lock()
	defer q.l.Unlock()

	_, ok := q.inqueue[item]
	return ok
}
