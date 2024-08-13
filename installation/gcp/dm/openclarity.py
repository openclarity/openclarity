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

"""Creates VMClarity"""

from hashlib import sha1


def GenerateConfig(context):
    """Creates VMClarity"""

    deployment_name_hash = sha1(context.env["deployment"].encode("utf8")).hexdigest()[:10]
    prefix = f"vmclarity-{context.env['deployment']}"

    network_name = f"{prefix}-network"
    staticip_name = f"{prefix}-static-ip"
    server_name = f"{prefix}-server"
    service_account_name = f"vmclarity-{deployment_name_hash}-sa"
    cloud_router_name = f"{prefix}-cloud-router"

    resources = [
        {
            "name": network_name,
            "type": "components/network.py",
        },
        {
            "name": staticip_name,
            "type": "components/static-ip.py",
            "properties": {
                "region": context.properties["region"],
            },
        },
        {
            "name": "firewallRules",
            "type": "components/firewall-rules.py",
            "properties": {
                "network": network_name,
            },
        },
        {
            "name": service_account_name,
            "type": "components/service-account.py",
        },
        {
            "name": "roles",
            "type": "components/roles.py",
            "properties": {
                "serviceAccount": service_account_name,
            },
        },
        {
            "name": cloud_router_name,
            "type": "components/cloud-router.py",
            "properties": {
                "network": network_name,
                "region": context.properties["region"],
            },
        },
        {
            "name": server_name,
            "type": "components/vmclarity-server.py",
            "properties": {
                "machineType": context.properties["machineType"],
                "zone": context.properties["zone"],
                "network": network_name,
                "staticIp": "$(ref." + staticip_name + ".address)",
                "sshPublicKey": context.properties["sshPublicKey"],
                "region": context.properties["region"],
                "serviceAccount": service_account_name,
                "scannerMachineType": context.properties["scannerMachineType"],
                "scannerSourceImage": context.properties["scannerSourceImage"],
                "databaseToUse": context.properties["databaseToUse"],
                "apiserverContainerImage": context.properties["apiserverContainerImage"],
                "orchestratorContainerImage": context.properties["orchestratorContainerImage"],
                "uiContainerImage": context.properties["uiContainerImage"],
                "uibackendContainerImage": context.properties["uibackendContainerImage"],
                "scannerContainerImage": context.properties["scannerContainerImage"],
                "exploitDBServerContainerImage": context.properties[
                    "exploitDBServerContainerImage"
                ],
                "trivyServerContainerImage": context.properties[
                    "trivyServerContainerImage"
                ],
                "grypeServerContainerImage": context.properties[
                    "grypeServerContainerImage"
                ],
                "freshclamMirrorContainerImage": context.properties[
                    "freshclamMirrorContainerImage"
                ],
                "postgresqlContainerImage": context.properties[
                    "postgresqlContainerImage"
                ],
                "assetScanDeletePolicy": context.properties["assetScanDeletePolicy"],
                # Optional properties
                "postgresDBPassword": context.properties.get("postgresDBPassword", ""),
                "externalDBName": context.properties.get("externalDBName", ""),
                "externalDBUsername": context.properties.get("externalDBUsername", ""),
                "externalDBPassword": context.properties.get("externalDBPassword", ""),
                "externalDBHost": context.properties.get("externalDBHost", ""),
                "externalDBPort": context.properties.get("externalDBPort", ""),
            },
        },
    ]
    outputs = [
        {
            "name": "ip",
            "value": "$(ref." + staticip_name + ".address)",
        }
    ]
    return {"resources": resources, "outputs": outputs}
