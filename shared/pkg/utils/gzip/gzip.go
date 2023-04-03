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

package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

// CompressAndEncode gzip and base64 encode the source input.
func CompressAndEncode(source []byte) (string, error) {
	if source == nil {
		return "", nil
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(source); err != nil {
		return "", fmt.Errorf("failed to gzip write: %w", err)
	}
	if err := gz.Flush(); err != nil {
		return "", fmt.Errorf("failed to gzip flush: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("failed to gzip close: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

// DecodeAndUncompress base64 decode and unzip the source input.
func DecodeAndUncompress(source string) ([]byte, error) {
	if source == "" {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode: %w", err)
	}
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	s, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	return s, nil
}
