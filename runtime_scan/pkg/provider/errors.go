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
	"fmt"
	"time"
)

type operationError interface {
	error

	Retryable() bool
	RetryAfter() time.Duration
}

func FatalErrorf(tmpl string, parts ...interface{}) FatalError {
	return FatalError{Err: fmt.Errorf(tmpl, parts...)}
}

type FatalError struct {
	Err error
}

func (e FatalError) Error() string {
	return e.Err.Error()
}

func (e FatalError) Unwrap() error {
	return e.Err
}

func (e FatalError) Retryable() bool {
	return false
}

func (e FatalError) RetryAfter() time.Duration {
	return -1
}

func RetryableErrorf(d time.Duration, tmpl string, parts ...interface{}) RetryableError {
	return RetryableError{
		Err:   fmt.Errorf(tmpl, parts...),
		After: d,
	}
}

type RetryableError struct {
	Err   error
	After time.Duration
}

func (e RetryableError) Error() string {
	return e.Err.Error()
}

func (e RetryableError) Unwrap() error {
	return e.Err
}

func (e RetryableError) Retryable() bool {
	return true
}

func (e RetryableError) RetryAfter() time.Duration {
	return e.After
}
