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
	"time"
)

type Enqueuer[T ReconcileEvent] interface {
	Enqueue(item T)
	EnqueueAfter(item T, d time.Duration)
}

type Dequeuer[T ReconcileEvent] interface {
	Dequeue(ctx context.Context) (T, error)
	Done(item T)
	RequeueAfter(item T, d time.Duration)
}

type Queue[T ReconcileEvent] struct {
	// Channel used internally to block Dequeue when the queue is empty,
	// and notify Dequeue when a new item is added through Enqueue.
	itemAdded chan struct{}

	// A slice which represents the queue, Enqueue will add to the end of
	// the slice, Dequeue will remove from the head of the slice.
	queue []T

	// A map used as a set of unique items which are in the queue. This is
	// used by Enqueue and Has to provide a quick reference to whats in the
	// queue without needing to loop through the queue slice.
	inqueue map[string]struct{}

	// A map used as a set of unique items which are processing. This keeps
	// track of items which have been Dequeued but are still being
	// processed outside of the queue and therefore shouldn't be requeued
	// if an Enqueue request is received for them.
	//
	// We use a separate map for this instead of reusing inqueue to prevent
	// calls to Done() removing items from inqueue when they are actually
	// in the queue slice.
	processing map[string]struct{}

	// A map to track items which have been scheduled to be queued at a
	// later date. We keep track of these items to prevent them being
	// enqueued earlier than their scheduled time through Enqueue.
	waitingForEnqueue map[string]struct{}

	// A mutex lock which protects the queue from simultaneous reads and
	// writes ensuring the queue can be used by multiple go routines safely.
	l sync.Mutex
}

func NewQueue[T ReconcileEvent]() *Queue[T] {
	return &Queue[T]{
		itemAdded:         make(chan struct{}),
		queue:             make([]T, 0),
		inqueue:           make(map[string]struct{}),
		processing:        make(map[string]struct{}),
		waitingForEnqueue: make(map[string]struct{}),
	}
}

// Dequeue until it can dequeue an item from the queue or the passed context is
// cancelled. The queue will keep track of the item and prevent Enqueuing until
// "Done" is called with the dequeued item to acknowledge its processing is
// completed.
func (q *Queue[T]) Dequeue(ctx context.Context) (T, error) {
	// Grab the lock so that we can check the length of q.queue safely.
	q.l.Lock()
	if len(q.queue) == 0 {
		// Unlock while we wait for an item to be added to the queue
		q.l.Unlock()

		// If the queue is empty, block waiting for the itemAdded
		// notification or context timeout.
		select {
		case <-q.itemAdded:
			// We know we have an item added to the queue so grab
			// the lock so that we can dequeue it safely
			q.l.Lock()
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
	defer q.l.Unlock()

	item := q.queue[0]
	q.queue = q.queue[1:]
	itemKey := item.Hash()
	delete(q.inqueue, itemKey)
	q.processing[itemKey] = struct{}{}

	return item, nil
}

// Enqueue will add item to the queue if its not in the queue already.
func (q *Queue[T]) Enqueue(item T) {
	q.l.Lock()
	defer q.l.Unlock()

	q.enqueue(item)
}

// EnqueueAfter will tell the queue to keep track of the item preventing it
// being enqueued before the specified duration through Enqueue. It will then
// Enqueue it after a defined duration. If item already in the queue, is
// processing or waiting to be enqueued, nothing will be done.
func (q *Queue[T]) EnqueueAfter(item T, d time.Duration) {
	q.l.Lock()
	defer q.l.Unlock()

	q.enqueueAfter(item, d)
}

// Internal enqueueAfter function that it can be reused by the public
// EnqueueAfter and public RequeueAfter functions, these should not be called
// without obtaining a lock on Queue first.
func (q *Queue[T]) enqueueAfter(item T, d time.Duration) {
	itemKey := item.Hash()
	_, inQueue := q.inqueue[itemKey]
	_, isProcessing := q.processing[itemKey]
	_, isWaitingForEnqueue := q.waitingForEnqueue[itemKey]
	if inQueue || isProcessing || isWaitingForEnqueue {
		// item is already known by the queue so there is nothing to do
		return
	}

	q.waitingForEnqueue[itemKey] = struct{}{}
	go func() {
		<-time.After(d)
		q.l.Lock()
		defer q.l.Unlock()
		delete(q.waitingForEnqueue, itemKey)
		q.enqueue(item)
	}()
}

// Internal enqueue function that it can be reused by public functions Enqueue
// and EnqueueAfter.
func (q *Queue[T]) enqueue(item T) {
	itemKey := item.Hash()
	_, inQueue := q.inqueue[itemKey]
	_, isProcessing := q.processing[itemKey]
	_, isWaitingForEnqueue := q.waitingForEnqueue[itemKey]
	if !inQueue && !isProcessing && !isWaitingForEnqueue {
		q.queue = append(q.queue, item)
		q.inqueue[itemKey] = struct{}{}

		select {
		case q.itemAdded <- struct{}{}:
		default:
		}
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

	itemKey := item.Hash()
	_, inQueue := q.inqueue[itemKey]
	_, isProcessing := q.processing[itemKey]
	_, isWaitingForEnqueue := q.waitingForEnqueue[itemKey]
	return inQueue || isProcessing || isWaitingForEnqueue
}

// Done will tell the queue that the processing for the item is completed and
// that it should stop tracking that item.
func (q *Queue[T]) Done(item T) {
	q.l.Lock()
	defer q.l.Unlock()
	delete(q.processing, item.Hash())
}

// RequeueAfter will mark a processing item as Done and then schedule it to be
// requeued after the specified duration. This should be used instead of
// calling Done and then EnqueueAfter to prevent someone Enqueuing the same
// item between Done and EnqueueAfter.
func (q *Queue[T]) RequeueAfter(item T, d time.Duration) {
	q.l.Lock()
	defer q.l.Unlock()

	delete(q.processing, item.Hash())
	q.enqueueAfter(item, d)
}

// Returns the number of items currently being actively processed.
func (q *Queue[T]) ProcessingCount() int {
	q.l.Lock()
	defer q.l.Unlock()

	return len(q.processing)
}
