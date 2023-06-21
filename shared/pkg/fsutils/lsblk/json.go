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
	"bytes"
	"encoding/json"
	"fmt"
)

type BlockDevicesJSON struct {
	BlockDevices []BlockDeviceJSON `json:"blockdevices"`
}

type BlockDeviceJSON struct {
	BlockDevice

	Children []BlockDeviceJSON `json:"children,omitempty"`
}

func (d *BlockDeviceJSON) flatten() []BlockDevice {
	var result []BlockDevice

	result = append(result, d.BlockDevice)

	if d.Children != nil {
		for _, c := range d.Children {
			result = append(result, c.flatten()...)
		}
	}

	return result
}

func parseJSONFormat(b *bytes.Buffer) ([]BlockDevice, error) {
	if b == nil {
		return nil, nil
	}

	j := BlockDevicesJSON{}
	decoder := json.NewDecoder(b)
	if err := decoder.Decode(&j); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	var blockDevices []BlockDevice
	for _, dev := range j.BlockDevices {
		blockDevices = append(blockDevices, dev.flatten()...)
	}

	return blockDevices, nil
}
