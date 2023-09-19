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

"""Helper for generating vmclarity server install script."""


def GenerateInstallScript(context):
    template = context.imports["vmclarity-install.sh"]

    values = {
        "DatabaseToUse": context.properties["databaseToUse"],
        "ScannerContainerImage": context.properties["scannerContainerImage"],
        "AssetScanDeletePolicy": context.properties["assetScanDeletePolicy"],
        "APIServerContainerImage": context.properties["apiserverContainerImage"],
        "OrchestratorContainerImage": context.properties["orchestratorContainerImage"],
        "UIContainerImage": context.properties["uiContainerImage"],
        "UIBackendContainerImage": context.properties["uibackendContainerImage"],
        "ExploitDBServerContainerImage": context.properties[
            "exploitDBServerContainerImage"
        ],
        "TrivyServerContainerImage": context.properties["trivyServerContainerImage"],
        "GrypeServerContainerImage": context.properties["grypeServerContainerImage"],
        "FreshclamMirrorContainerImage": context.properties[
            "freshclamMirrorContainerImage"
        ],
        "YaraRuleServerContainerImage": context.properties["yaraRuleServerContainerImage"],
        "PostgresqlContainerImage": context.properties["postgresqlContainerImage"],
        "ProjectID": context.env["project"],
        "ScannerZone": context.properties["zone"],
        "ScannerSubnet": (
            f"projects/{context.env['project']}/"
            f"regions/{context.properties['region']}/"
            f"subnetworks/{context.properties['network']}"
        ),
        "ScannerMachineType": context.properties["scannerMachineType"],
        "ScannerSourceImage": context.properties["scannerSourceImage"],
        # Optional parameters
        "PostgresDBPassword": context.properties.get("postgresDBPassword", ""),
        "ExternalDBName": context.properties.get("externalDBName", ""),
        "ExternalDBUsername": context.properties.get("externalDBUsername", ""),
        "ExternalDBPassword": context.properties.get("externalDBPassword", ""),
        "ExternalDBHost": context.properties.get("externalDBHost", ""),
        "ExternalDBPort": context.properties.get("externalDBPort", ""),
    }
    return template.format(**values)
