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

package slice

func ToMap(input []string) map[string]bool {
	output := map[string]bool{}
	for _, s := range input {
		output[s] = true
	}

	return output
}

// FindUnique returns a list of strings that exist in 'a' and not in 'b'.
func FindUnique(a, b []string) []string {
	aMap := ToMap(a)
	bMap := ToMap(b)

	return notInMap(aMap, bMap)
}

// returns a list of strings that exist in 'a' and not in 'b'.
func notInMap(a, b map[string]bool) []string {
	var ret []string

	for s := range a {
		if !b[s] {
			ret = append(ret, s)
		}
	}

	return ret
}

func RemoveStringDuplicates(slice []string) []string {
	retMap := make(map[string]struct{})
	var ret []string // nolint:prealloc

	for _, elem := range slice {
		retMap[elem] = struct{}{}
	}
	for elem := range retMap {
		ret = append(ret, elem)
	}

	return ret
}

func RemoveEmptyStrings(slice []string) []string {
	var ret []string
	for _, str := range slice {
		if str != "" {
			ret = append(ret, str)
		}
	}

	return ret
}

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
