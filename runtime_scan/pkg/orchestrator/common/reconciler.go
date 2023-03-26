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

type Reconciler[T any] struct {
	Logger *log.Entry

	// Channel on which reconcile events will be received
	EventChan chan T

	// Reconcile function which will be called whenever there is an event on EventChan
	ReconcileFunction func(context.Context, T) error

	// Maximum amount of time to spend trying to reconcile one item before
	// moving onto the next item.
	ReconcileTimeout time.Duration
}

func (r *Reconciler[T]) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case item := <-r.EventChan:
				timeoutCtx, cancel := context.WithTimeout(ctx, r.ReconcileTimeout)
				err := r.ReconcileFunction(timeoutCtx, item)
				if err != nil {
					r.Logger.Errorf("Failed to reconcile item: %v", err)
				}
				cancel()
			case <-ctx.Done():
				return
			}
		}
	}()
}
