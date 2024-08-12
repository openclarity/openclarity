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

package fanout

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

// FanOut replicates the byte stream to a list of consumers read from r io.Reader.
func FanOut(ctx context.Context, r io.Reader, consumers []func(io.Reader) error) error {
	if r == nil || consumers == nil {
		return nil
	}

	var wg sync.WaitGroup
	numOfConsumers := len(consumers)
	errs := make(chan error, numOfConsumers+1)

	writer := newFanOutWriter(numOfConsumers)
	readers := writer.Out()

	// Start consumers
	for idx, consumer := range consumers {
		reader := readers[idx]

		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case <-ctx.Done():
				errs <- fmt.Errorf("failed to read: %w", ctx.Err())
				break
			case errs <- consumer(reader):
			}
		}()
	}

	// Start source reader
	producer := func() error {
		defer func(w *fanOutWriter) {
			_ = w.Close()
		}(writer)

		_, err := io.Copy(writer.In(), r)

		return err //nolint:wrapcheck
	}

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()

		select {
		case <-ctx.Done():
			errs <- fmt.Errorf("failed to read: %w", ctx.Err())
			break
		case errs <- producer():
		}
	}(ctx)
	wg.Wait()
	close(errs)

	// Drain error channel
	listOfErrors := make([]error, 0)
	for e := range errs {
		if e != nil {
			listOfErrors = append(listOfErrors, e)
		}
	}

	return errors.Join(listOfErrors...)
}
