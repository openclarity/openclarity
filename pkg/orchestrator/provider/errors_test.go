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

package provider

import (
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		Name string
		Err  error

		ExpectedError         operationError
		ExpectedToBeRetryable bool
		ExpectedRetryAfter    time.Duration
	}{
		{
			Name:                  "FatalError",
			Err:                   FatalError{Err: errors.New("fatal error")},
			ExpectedError:         FatalError{},
			ExpectedToBeRetryable: false,
			ExpectedRetryAfter:    -1,
		},
		{
			Name:                  "RetryableError",
			Err:                   RetryableError{Err: errors.New("retryable error"), After: time.Minute},
			ExpectedError:         RetryableError{},
			ExpectedToBeRetryable: true,
			ExpectedRetryAfter:    time.Minute,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := test.ExpectedError
			match := errors.As(test.Err, &err)

			g.Expect(match).Should(BeTrue())
			g.Expect(err.Retryable()).Should(Equal(test.ExpectedToBeRetryable))
			g.Expect(err.RetryAfter()).Should(Equal(test.ExpectedRetryAfter))
		})
	}
}
