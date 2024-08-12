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

"""Creates the custom roles."""

from hashlib import sha1


def GenerateConfig(context):
    """Creates the custom roles"""

    deployment_name_hash = sha1(context.env["deployment"].encode("utf8")).hexdigest()[:10]
    prefix = f"vmclarity-{deployment_name_hash}"

    discoverer_role_name = f"{prefix}-discoverer-snapshotter"
    discoverer_role_id = discoverer_role_name.replace("-", "_")

    scanner_role_name = f"{prefix}-scanner"
    scanner_role_id = scanner_role_name.replace("-", "_")

    resources = [
        {
            "name": discoverer_role_name,
            "type": "gcp-types/iam-v1:projects.roles",
            "properties": {
                "parent": f"projects/{context.env['project']}",
                "roleId": discoverer_role_id,
                "role": {
                    "title": (
                        f"VMClarity {context.env['deployment']} "
                        f"Discoverer Snapshotter"
                    ),
                    "description": (
                        f"Role to allow vmclarity {context.env['deployment']} "
                        f"to discover and snapshot instances in the project"
                    ),
                    "stage": "GA",
                    "includedPermissions": [
                        "compute.regions.list",
                        "compute.disks.get",
                        "compute.disks.list",
                        "compute.instances.get",
                        "compute.instances.list",
                        "compute.disks.createSnapshot",
                        "compute.snapshots.get",
                        "compute.snapshots.list",
                        "compute.snapshots.create",
                        "compute.snapshots.delete",
                    ],
                },
            },
        },
        {
            "name": scanner_role_name,
            "type": "gcp-types/iam-v1:projects.roles",
            "properties": {
                "parent": f"projects/{context.env['project']}",
                "roleId": scanner_role_id,
                "role": {
                    "title": f"VMClarity {context.env['deployment']} Scanner",
                    "description": (
                        f"Role to allow vmclarity {context.env['deployment']} "
                        f"to create and manage scanner instances"
                    ),
                    "stage": "GA",
                    "includedPermissions": [
                        "compute.disks.use",
                        "compute.disks.get",
                        "compute.disks.list",
                        "compute.disks.create",
                        "compute.disks.delete",
                        "compute.disks.setLabels",
                        "compute.instances.attachDisk",
                        "compute.snapshots.useReadOnly",
                        "compute.instances.get",
                        "compute.instances.list",
                        "compute.instances.create",
                        "compute.instances.delete",
                        "compute.instances.setMetadata",
                        "compute.instances.setTags",
                        "compute.instances.setLabels",
                        "compute.images.useReadOnly",
                        "compute.subnetworks.use",
                    ],
                },
            },
        },
        {
            "name": f"{prefix}-discoverer-snapshotter-role-binding",
            "type": "gcp-types/cloudresourcemanager-v1:virtual.projects.iamMemberBinding",
            "properties": {
                "resource": context.env["project"],
                "role": f"projects/{context.env['project']}/roles/{discoverer_role_id}",
                "member": (
                    f"serviceAccount:"
                    f"$(ref.{context.properties['serviceAccount']}.email)"
                ),
            },
        },
        {
            "name": f"{prefix}-scanner-role-binding",
            "type": "gcp-types/cloudresourcemanager-v1:virtual.projects.iamMemberBinding",
            "properties": {
                "resource": context.env["project"],
                "role": f"projects/{context.env['project']}/roles/{scanner_role_id}",
                "member": (
                    f"serviceAccount:"
                    f"$(ref.{context.properties['serviceAccount']}.email)"
                ),
            },
        },
    ]
    return {"resources": resources}
