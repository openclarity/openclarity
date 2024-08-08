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

package diskutil

import (
	"bytes"
	_ "embed"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

//go:embed testdata/diskutil.txt
var diskutilOutput []byte

// nolint:maintidx
func TestParse(t *testing.T) {
	tests := []struct {
		Name     string
		TextData *bytes.Buffer
		Err      error

		ExpectedErrorMatcher    types.GomegaMatcher
		ExpectedBlockDeviceList []BlockDevice
	}{
		{
			Name:                    "Nil data",
			Err:                     nil,
			ExpectedErrorMatcher:    Not(HaveOccurred()),
			ExpectedBlockDeviceList: nil,
		},
		{
			Name:                 "Diskutil output",
			TextData:             bytes.NewBuffer(diskutilOutput),
			Err:                  nil,
			ExpectedErrorMatcher: Not(HaveOccurred()),
			ExpectedBlockDeviceList: []BlockDevice{
				{
					DeviceIdentifier: "disk0",
					Path:             "/dev/disk0",
					Whole:            true,
					PartOfWhole:      "disk0",
					DeviceMediaName:  "APPLE SSD 0000",

					VolumeName: "",
					Mounted:    false,
					FileSystem: "",

					Content:          "GUID_partition_scheme",
					OSCanBeInstalled: false,
					MediaType:        "Generic",
					Protocol:         "Apple Fabric",
					SMARTStatus:      "Verified",

					DiskSize:        1073741824,
					DeviceBlockSize: 4096,

					MediaOSUseOnly: false,
					MediaReadOnly:  false,
					VolumeReadOnly: false,

					DeviceLocation: "Internal",
					RemovableMedia: "Fixed",

					SolidState:         true,
					HardwareAESSupport: true,
				},
				{
					DeviceIdentifier: "disk3s6",
					Path:             "/dev/disk3s6",
					Whole:            false,
					PartOfWhole:      "disk3",

					VolumeName: "VM",
					Mounted:    true,
					MountPoint: "/System/Volumes/VM",

					PartitionType:   "0000-0000-0000-0000-0000",
					FSType:          "APFS",
					TypeBundle:      "apfs",
					NameUserVisible: "APFS",
					Owners:          "Enabled",

					OSCanBeInstalled:  false,
					BooterDisk:        "disk3s2",
					RecoveryDisk:      "disk3s3",
					MediaType:         "Generic",
					Protocol:          "Apple Fabric",
					SMARTStatus:       "Verified",
					VolumeUUID:        "0000-0000-0000-0000-0000",
					DiskPartitionUUID: "0000-0000-0000-0000-0000",

					DiskSize:        1073741824,
					DeviceBlockSize: 4096,

					VolumeUsedSpace:     2254857830.4,
					ContainerTotalSpace: 1073741824,
					ContainerFreeSpace:  1073741824,
					AllocationBlockSize: 4096,

					MediaOSUseOnly: false,
					MediaReadOnly:  false,
					VolumeReadOnly: false,

					DeviceLocation: "Internal",
					RemovableMedia: "Fixed",

					SolidState:         true,
					HardwareAESSupport: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			list, err := parse(test.TextData)

			g.Expect(err).Should(test.ExpectedErrorMatcher)
			g.Expect(list).Should(BeEquivalentTo(test.ExpectedBlockDeviceList))
		})
	}
}
