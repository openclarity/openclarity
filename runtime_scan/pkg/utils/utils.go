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

package utils

func StringPtr(val string) *string {
	ret := val
	return &ret
}

func BoolPtr(val bool) *bool {
	ret := val
	return &ret
}

func Int32Ptr(val int32) *int32 {
	ret := val
	return &ret
}

func IntPtr(val int) *int {
	ret := val
	return &ret
}

func PointerTo[T any](value T) *T {
	return &value
}
