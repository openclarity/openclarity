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

	log "github.com/sirupsen/logrus"
)

type Poller[T any] struct {
	Logger *log.Entry

	// How often to re-poll the API for new items and try to publish them
	// on the event channel. If the current items aren't handled they will
	// be dropped and new items fetched when the PollPeriod is up.
	PollPeriod time.Duration

	// The function which will be called to get the list of items to be
	// published on the event channel.
	GetItems func(context.Context) ([]T, error)
}

func (p *Poller[T]) Start(ctx context.Context) chan T {
	eventChan := make(chan T)
	go func() {
		defer close(eventChan)
		for {
			// Create a timeout context so that we can re-poll the
			// items at fixed intervals regardless of how far
			// through the items we got, this prevents us holding
			// onto to stale items.
			timeoutCtx, cancel := context.WithTimeout(ctx, p.PollPeriod)
			defer cancel()

			items, err := p.GetItems(timeoutCtx)
			if err != nil {
				p.Logger.Errorf("Failed to get items to reconcile: %v", err)
			} else {
				p.Logger.Infof("Found %d items", len(items))
			itemLoop:
				for _, item := range items {
					select {
					// Try to send item to the reconcile event listener
					case eventChan <- item:

					// We've been stopped from reconciling
					// any more items, either we've been
					// cancelled or we've timed out.
					case <-timeoutCtx.Done():
						p.Logger.Errorf("Failed to reconcile all items within PollPeriod")
						break itemLoop
					}
				}
			}

			// If we've completed all the reconciles within the
			// reconcile period, keep waiting for the reconcile
			// time to be up before re-requesting items. This
			// ensures that each reconcile is a fixed length of
			// time.
			<-timeoutCtx.Done()

			select {
			// Check if the parent context was the reason
			// timeoutCtx was canceled, if it was the parent
			// context then we've been cancelled so we must stop
			// and return, otherwise continue to the next
			// poll.
			case <-ctx.Done():
				p.Logger.Info("Shutting down")
				return
			default:
			}
		}
	}()
	return eventChan
}
