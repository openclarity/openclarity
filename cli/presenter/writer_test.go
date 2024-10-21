// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
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

package presenter

import (
	"os"
	"reflect"
	"testing"
)

func TestConsoleWriter(t *testing.T) {
	writer := &ConsoleWriter{Output: os.Stdout}

	tests := []struct {
		name   string
		b      []byte
		prefix string
		want   error
	}{
		{
			name:   "nil bytes no error",
			b:      nil,
			prefix: "sbom",
			want:   nil,
		},
		{
			name:   "empty bytes no error",
			b:      []byte{},
			prefix: "sbom",
			want:   nil,
		},
		{
			name:   "test content no error",
			b:      []byte("test"),
			prefix: "sbom",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writer.Write(tt.b, tt.prefix); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConsoleWriter.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileWriter(t *testing.T) {
	writer := &FileWriter{Path: "."}

	tests := []struct {
		name     string
		b        []byte
		filename string
		want     error
	}{
		{
			name:     "nil bytes no error",
			b:        nil,
			filename: "sbom.cdx",
			want:     nil,
		},
		{
			name:     "empty bytes no error",
			b:        []byte{},
			filename: "sbom.cdx",
			want:     nil,
		},
		{
			name:     "test content no error",
			b:        []byte("test"),
			filename: "sbom.cdx",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writer.Write(tt.b, tt.filename); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileWriter.Write() = %v, want %v", got, tt.want)
			}
		})
		// cleanup
		os.Remove(tt.filename)
	}
}
