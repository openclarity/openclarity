# Configuration

## Orchestrator

| Environment Variable                      | Required  | Default | Description                                  |
|-------------------------------------------|-----------|---------|----------------------------------------------|
| `DELETE_JOB_POLICY`                       |           |         |                                              |
| `SCANNER_CONTAINER_IMAGE`                 |           |         |                                              |
| `GITLEAKS_BINARY_PATH`                    |           |         |                                              |
| `CLAM_BINARY_PATHCLAM_BINARY_PATH`        |           |         |                                              |
| `YARA_BINARY_PATH`                        |           |         |                                              |
| `FRESHCLAM_BINARY_PATH`                   |           |         |                                              |
| `ALTERNATIVE_FRESHCLAM_MIRROR_URL`        |           |         |                                              |
| `LYNIS_INSTALL_PATH`                      |           |         |                                              |
| `SCANNER_VMCLARITY_BACKEND_ADDRESS`       |           |         |                                              |
| `EXPLOIT_DB_ADDRESS`                      |           |         |                                              |
| `TRIVY_SERVER_ADDRESS`                    |           |         |                                              |
| `TRIVY_SERVER_TIMEOUT`                    |           |         |                                              |
| `YARA_RULE_SERVER_ADDRESS`                |           |         |                                              |
| `GRYPE_SERVER_ADDRESS`                    |           |         |                                              |
| `GRYPE_SERVER_TIMEOUT`                    |           |         |                                              |
| `CHKROOTKIT_BINARY_PATH`                  |           |         |                                              |
| `SCAN_CONFIG_POLLING_INTERVAL`            |           |         |                                              |
| `SCAN_CONFIG_RECONCILE_TIMEOUT`           |           |         |                                              |
| `SCAN_POLLING_INTERVAL`                   |           |         |                                              |
| `SCAN_RECONCILE_TIMEOUT`                  |           |         |                                              |
| `SCAN_TIMEOUT`                            |           |         |                                              |
| `ASSET_SCAN_POLLING_INTERVAL`             |           |         |                                              |
| `ASSET_SCAN_RECONCILE_TIMEOUT`            |           |         |                                              |
| `ASSET_SCAN_PROCESSOR_POLLING_INTERVAL`   |           |         |                                              |
| `ASSET_SCAN_PROCESSOR_RECONCILE_TIMEOUT`  |           |         |                                              |
| `DISCOVERY_INTERVAL`                      |           |         |                                              |
| `CONTROLLER_STARTUP_DELAY`                |           |         |                                              |
| `PROVIDER`                                | **yes**   | `aws`   | Provider used for Asset discovery and scans |

## Provider

### AWS

| Environment Variable                   | Required | Default      | Description                                                                   |
|----------------------------------------|----------|--------------|-------------------------------------------------------------------------------|
| `VMCLARITY_AWS_REGION`                 | **yes**  |              | Region where the Scanner instance needs to be created                         |
| `VMCLARITY_AWS_SUBNET_ID`              | **yes**  |              | SubnetID where the Scanner instance needs to be created                       |
| `VMCLARITY_AWS_SECURITY_GROUP_ID`      | **yes**  |              | SecurityGroupId which needs to be attached to the Scanner instance            |
| `VMCLARITY_AWS_KEYPAIR_NAME`           |          |              | Name of the SSH KeyPair to use for Scanner instance launch                    |
| `VMCLARITY_AWS_SCANNER_AMI_ID`         | **yes**  |              | The AMI image used for creating Scanner instance                              |
| `VMCLARITY_AWS_SCANNER_INSTANCE_TYPE`  |          | `t2.large`   | The instance type used for Scanner instance                                   |
| `VMCLARITY_AWS_BLOCK_DEVICE_NAME`      |          | `xvdh`       | Block device name used for attaching Scanner volume to the Scanner instance   |
