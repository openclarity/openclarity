// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package utils

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

type WrappedError struct {
	err error
}

func (e WrappedError) Error() string {
	return e.err.Error()
}

func (e WrappedError) Unwrap() error {
	return e.err
}

type WrappedMultiError struct {
	errs []error
}

func (e WrappedMultiError) Error() string {
	return errors.Join(e.errs...).Error()
}

func (e WrappedMultiError) Unwrap() []error {
	return e.errs
}

func TestUnwrapErrors(t *testing.T) {
	testErrors := []error{
		errors.New("error 1"),
		errors.New("error 2"),
		fmt.Errorf("error 3.1: %w", errors.New("error 3.2")),
	}

	tests := []struct {
		Name string
		Err  error

		ExpectedErrors []error
	}{
		{
			Name:           "Nil error",
			Err:            nil,
			ExpectedErrors: nil,
		},
		{
			Name: "Unwrapped error",
			Err:  testErrors[0],
			ExpectedErrors: []error{
				testErrors[0],
			},
		},
		{
			Name: "Wrapped error",
			Err: WrappedError{
				err: testErrors[0],
			},
			ExpectedErrors: []error{
				testErrors[0],
			},
		},
		{
			Name: "Wrapped errors",
			Err: WrappedMultiError{
				errs: testErrors,
			},
			ExpectedErrors: testErrors,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			errs := UnwrapErrors(test.Err)
			g.Expect(errs).Should(Equal(test.ExpectedErrors))
		})
	}
}

func TestUnwrapErrorStrings(t *testing.T) {
	testErrors := []error{
		errors.New("error 1"),
		errors.New("error 2"),
		fmt.Errorf("error 3.1: %w", errors.New("error 3.2")),
	}

	tests := []struct {
		Name string
		Err  error

		ExpectedErrors []string
	}{
		{
			Name:           "Nil error",
			Err:            nil,
			ExpectedErrors: nil,
		},
		{
			Name: "Unwrapped error",
			Err:  testErrors[0],
			ExpectedErrors: []string{
				testErrors[0].Error(),
			},
		},
		{
			Name: "Wrapped error",
			Err: WrappedError{
				err: testErrors[0],
			},
			ExpectedErrors: []string{
				testErrors[0].Error(),
			},
		},
		{
			Name: "Wrapped errors",
			Err: WrappedMultiError{
				errs: testErrors,
			},
			ExpectedErrors: []string{
				testErrors[0].Error(),
				testErrors[1].Error(),
				testErrors[2].Error(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			errs := UnwrapErrorStrings(test.Err)
			g.Expect(errs).Should(Equal(test.ExpectedErrors))
		})
	}
}
