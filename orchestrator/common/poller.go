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
	"time"

	"github.com/openclarity/vmclarity/core/log"
)

type Poller[T ReconcileEvent] struct {
	// How often to re-poll the API for new items and try to publish them
	// on the event channel. If the current items aren't handled they will
	// be dropped and new items fetched when the PollPeriod is up.
	PollPeriod time.Duration

	// The function which will be called to get the list of items to be
	// published on the event channel.
	GetItems func(context.Context) ([]T, error)

	// The queue to which we add items to reconcile.
	Queue Enqueuer[T]
}

func (p *Poller[T]) pollThenWait(ctx context.Context) {
	logger := log.GetLoggerFromContextOrDiscard(ctx)

	// Create a timeout context so that we can re-poll the
	// items at fixed intervals regardless of how far
	// through the items we got, this prevents us holding
	// onto to stale items.
	timeoutCtx, cancel := context.WithTimeout(ctx, p.PollPeriod)

	// Defer cancel even though we're waiting on the timeout at the bottom
	// of the function, so that even if we panic for some reason the
	// context is cleaned up
	defer cancel()

	items, err := p.GetItems(timeoutCtx)
	if err != nil {
		logger.Errorf("Failed to get items to reconcile: %v", err)
	} else {
		logger.Debugf("Found %d items to reconcile, adding them to the queue", len(items))
		for _, item := range items {
			logger.WithFields(item.ToFields()).Debugf("Adding item to the queue")
			p.Queue.Enqueue(item)
		}
	}

	// Once we've added all the items to the queue wait for
	// the poll period time to be up before re-requesting
	// items. This ensures that each reconcile is a fixed
	// length of time regardless of how long the GetItems
	// request and Enqueuing logic took.
	<-timeoutCtx.Done()
}

func (p *Poller[T]) Start(ctx context.Context) {
	go func() {
		logger := log.GetLoggerFromContextOrDiscard(ctx)

		for {
			p.pollThenWait(ctx)

			select {
			// Check if the parent context was canceled, if it was
			// we must stop and return, otherwise continue to the
			// next poll.
			case <-ctx.Done():
				logger.Info("Shutting down")
				return
			default:
			}
		}
	}()
}
