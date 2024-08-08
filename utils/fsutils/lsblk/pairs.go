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

package lsblk

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/go-viper/mapstructure/v2"

	"github.com/openclarity/vmclarity/utils/fsutils/lsblk/internal"
)

func parsePairsFormat(b *bytes.Buffer) ([]BlockDevice, error) {
	if b == nil {
		return nil, nil
	}

	data := make([]map[string]string, 0)

	lineScanner := bufio.NewScanner(b)
	lineScanner.Split(bufio.ScanLines)
	for lineScanner.Scan() {
		pairScanner := bufio.NewScanner(bytes.NewBuffer(lineScanner.Bytes()))
		pairScanner.Split(internal.ScanPairs)

		pairMap := make(map[string]string)

		for pairScanner.Scan() {
			key, value, ok := bytes.Cut(pairScanner.Bytes(), []byte("="))
			if !ok {
				continue
			}

			pairMap[string(key)] = string(bytes.Trim(value, "\""))
		}

		data = append(data, pairMap)
	}

	blockDeviceList := []BlockDevice{}
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &blockDeviceList,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(data); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	return blockDeviceList, nil
}
