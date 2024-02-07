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
	"errors"
	"fmt"
	"time"

	"github.com/openclarity/vmclarity/core/log"
)

type RequeueAfterError struct {
	d   time.Duration
	msg string
}

func (rae RequeueAfterError) Error() string {
	if rae.msg != "" {
		return fmt.Sprintf("%v so requeuing after %v", rae.msg, rae.d)
	}
	return fmt.Sprintf("requeuing after %v", rae.d)
}

func NewRequeueAfterError(d time.Duration, msg string) error {
	return RequeueAfterError{d, msg}
}

type Reconciler[T ReconcileEvent] struct {
	// Reconcile function which will be called whenever there is an event on EventChan
	ReconcileFunction func(context.Context, T) error

	// Maximum amount of time to spend trying to reconcile one item before
	// moving onto the next item.
	ReconcileTimeout time.Duration

	// The queue which the reconciler will receive events to reconcile on.
	Queue Dequeuer[T]
}

func (r *Reconciler[T]) Start(ctx context.Context) {
	go func() {
		logger := log.GetLoggerFromContextOrDiscard(ctx)

		for {
			// queue.Get will block until an item is available to
			// return.
			item, err := r.Queue.Dequeue(ctx)
			if err != nil {
				logger.Errorf("Failed to get item from queue: %v", err)
			} else {
				// NOTE: shadowing logger variable is intentional
				logger := logger.WithFields(item.ToFields())
				logger.Infof("Reconciling item")
				timeoutCtx, cancel := context.WithTimeout(ctx, r.ReconcileTimeout)
				err := r.ReconcileFunction(timeoutCtx, item)

				// Make sure timeout context is canceled to
				// prevent orphaned resources
				cancel()

				// If reconcile has requested that we requeue the item
				// by returning a RequeueAfterError then requeue the
				// item with the duration specified, otherwise mark the
				// item as Done.
				var requeueAfterError RequeueAfterError
				switch {
				case errors.As(err, &requeueAfterError):
					logger.Infof("Requeue item: %v", err)
					r.Queue.RequeueAfter(item, requeueAfterError.d)
				case err != nil:
					logger.Errorf("Failed to reconcile item: %v", err)
					fallthrough
				default:
					r.Queue.Done(item)
				}
			}

			// Check if the parent context done if so we also need
			// to exit.
			select {
			case <-ctx.Done():
				logger.Infof("Shutting down: %v", ctx.Err())
				return
			default:
			}
		}
	}()
}
