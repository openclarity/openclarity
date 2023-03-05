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

package cloudinit

const cloudInitTmpl string = `#cloud-config
package_upgrade: true
packages:
  - docker.io
write_files:
  - path: /opt/vmclarity/scanconfig.yaml
    permissions: "0644"
    content: |
{{ .ScannerCLIConfig | indent 6 }}
  - path: /etc/systemd/system/vmclarity-scanner.service
    permissions: "0644"
    content: |
      [Unit]
      Description=VMClarity scanner job
      Requires=docker.service
      After=network.target docker.service

      [Service]
      Type=oneshot
      WorkingDirectory=/opt/vmclarity
      ExecStartPre=docker pull {{ .ScannerImage }}
      ExecStart=docker run --rm --name %n --privileged \
          -v /opt/vmclarity:/opt/vmclarity \
          -v /run:/run \
          {{ .ScannerImage }} \
          --config /opt/vmclarity/scanconfig.yaml \
          --server {{ .ServerAddress }} \
          --mount-attached-volume \
          --scan-result-id {{ .ScanResultID }}

      [Install]
      WantedBy=multi-user.target
runcmd:
  - [ systemctl, daemon-reload ]
  - [ systemctl, start, docker.service ]
  - [ systemctl, start, vmclarity-scanner.service ]
`
