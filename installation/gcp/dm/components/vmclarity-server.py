# Copyright © 2023 Cisco Systems, Inc. and its affiliates.
# All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Creates the virtual machine with environment variables and startup script."""

import vmclarity_install_script

COMPUTE_URL_BASE = "https://www.googleapis.com/compute/v1/"

IMAGE_SELFLINK = "projects/ubuntu-os-cloud/global/images/ubuntu-2204-jammy-v20230630"


def GenerateConfig(context):
    """Creates the virtual machine."""

    startup_script = vmclarity_install_script.GenerateInstallScript(context)

    resources = [
        {
            "name": context.env["name"],
            "type": "compute.v1.instance",
            "properties": {
                "zone": context.properties["zone"],
                "tags": {
                    "items": [
                        "vmclarity-control-plane",
                    ],
                },
                "machineType": (
                    f"{COMPUTE_URL_BASE}"
                    f"projects/{context.env['project']}/"
                    f"zones/{context.properties['zone']}/"
                    f"machineTypes/{context.properties['machineType']}"
                ),
                "disks": [
                    {
                        "deviceName": "boot",
                        "type": "PERSISTENT",
                        "boot": True,
                        "autoDelete": True,
                        "initializeParams": {
                            "sourceImage": f"{COMPUTE_URL_BASE}{IMAGE_SELFLINK}",
                            "diskSizeGb": 30,
                            "diskType": (
                                f"projects/{context.env['project']}/"
                                f"zones/{context.properties['zone']}/"
                                f"diskTypes/pd-balanced"
                            ),
                        },
                    }
                ],
                "networkInterfaces": [
                    {
                        "network": f"$(ref.{context.properties['network']}.selfLink)",
                        "accessConfigs": [
                            {
                                "name": "External NAT",
                                "type": "ONE_TO_ONE_NAT",
                                "natIP": context.properties["staticIp"],
                            }
                        ],
                    }
                ],
                "serviceAccounts": [
                    {
                        "email": f"$(ref.{context.properties['serviceAccount']}.email)",
                        "scopes": [
                            "https://www.googleapis.com/auth/cloud-platform",
                        ],
                    }
                ],
                "metadata": {
                    "items": [
                        {
                            "key": "startup-script",
                            "value": startup_script,
                        },
                        {
                            "key": "ssh-keys",
                            "value": f"vmclarity:{context.properties['sshPublicKey']}",
                        },
                    ],
                },
            },
        }
    ]

    return {"resources": resources}
