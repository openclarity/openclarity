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

package internal

import (
	"unicode"
	"unicode/utf8"
)

func isDoubleQuote(r rune) bool {
	return r == '\u0022'
}

// ScanPairs is a modified version of bufio.ScanWords to support space delimited key-value pairs where the values are
// double-quoted and may include whitespaces. Like: 'FOO1="FOO" FOO2="BAR BAZ"'.
func ScanPairs(data []byte, atEOF bool) (int, []byte, error) {
	// Skip leading spaces.
	start := 0
	var width int
	for ; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !unicode.IsSpace(r) {
			break
		}
	}
	// Scan until space or closing double quote
	var quoted bool
	for i := start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])

		switch {
		case isDoubleQuote(r):
			quoted = !quoted
		case unicode.IsSpace(r) && !quoted:
			return i + width, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}
