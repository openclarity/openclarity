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

"""Creates the firewall."""


def GenerateConfig(context):
    """Creates the firewall"""

    prefix = f"vmclarity-{context.env['deployment']}"

    resources = [
        {
            "name": f"{prefix}-allow-ssh",
            "type": "compute.v1.firewall",
            "properties": {
                "network": "$(ref." + context.properties["network"] + ".selfLink)",
                "direction": "INGRESS",
                "sourceRanges": ["0.0.0.0/0"],
                "allowed": [
                    {
                        "IPProtocol": "TCP",
                        "ports": [22],
                    }
                ],
            },
        },
        {
            "name": f"{prefix}-allow-scanner-to-control-plane",
            "type": "compute.v1.firewall",
            "properties": {
                "network": "$(ref." + context.properties["network"] + ".selfLink)",
                "direction": "INGRESS",
                "sourceRanges": ["10.128.0.0/9"],
                "targetTags": ["vmclarity-control-plane"],
                "allowed": [
                    {
                        "IPProtocol": "TCP",
                        "ports": ["0-65535"],
                    }
                ],
            },
        },
    ]
    return {"resources": resources}
