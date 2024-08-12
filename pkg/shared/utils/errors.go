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

type (
	singleWrappedError = interface{ Unwrap() error }
	multiWrappedError  = interface{ Unwrap() []error }
)

// UnwrapErrors returns a slice of errors by unwrapping err error.
// Returned slice is nil if err is nil.
// Slice with the original err as a single element is returned if err does not implement `Unwrap() error`
// or `Unwrap []error` interfaces.
// UnwrapErrors does not perform recursive lookup, so only the top level err is unwrapped.
func UnwrapErrors(err error) []error {
	if err == nil {
		return nil
	}

	errs := make([]error, 0)

	// nolint:errorlint
	switch e := err.(type) {
	case singleWrappedError:
		if u := e.Unwrap(); u != nil {
			errs = append(errs, u)
		}
	case multiWrappedError:
		if u := e.Unwrap(); u != nil {
			errs = append(errs, u...)
		}
	default:
		errs = append(errs, err)
	}

	return errs
}

// UnwrapErrorStrings returns a slice of error strings by unwrapping err error.
// Returned slice is nil if provider err is nil.
// Slice with the original err as a single element is returned if err does not implement `Unwrap() error`
// or `Unwrap []error` interfaces.
// UnwrapErrorStrings does not perform recursive lookup, so only the top level err is unwrapped.
func UnwrapErrorStrings(err error) []string {
	if err == nil {
		return nil
	}

	errs := make([]string, 0)
	unwrappedErrs := UnwrapErrors(err)
	for _, u := range unwrappedErrs {
		errs = append(errs, u.Error())
	}

	return errs
}
