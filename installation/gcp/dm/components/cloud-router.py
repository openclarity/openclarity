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

"""Creates the cloud NAT"""


def GenerateConfig(context):
    """Creates the cloud NAT"""

    vmclarity_scanner_nat = {
        "name": f"{context.env['name']}-scanner-nat",
        "natIpAllocateOption": "AUTO_ONLY",
        "sourceSubnetworkIpRangesToNat": "ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES",
    }

    resources = [
        {
            "name": context.env["name"],
            "type": "compute.v1.router",
            "properties": {
                "network": f"$(ref.{context.properties['network']}.selfLink)",
                "region": context.properties["region"],
                "nats": [
                    vmclarity_scanner_nat,
                ],
            },
        }
    ]
    return {"resources": resources}
