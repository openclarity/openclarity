# End-to-end testing guide

## Table of Contents

- [Installing a specific VMClarity build on AWS](#installing-a-specific-vmclarity-build-on-aws)
  - [1. Build the containers and publish them to your Docker registry](#1.-build-the-containers-and-publish-them-to-your-docker-registry)
  - [2. Install VMClarity CloudFormation](#2.-install-vmclarity-cloudformation)
  - [3. Ensure that VMClarity backend is working correctly](#3.-ensure-that-vmclarity-backend-is-working-correctly)
- [Performing an end-to-end test](#performing-an-end-to-end-test)

## Installing a specific VMClarity build on AWS

### 1. Build the containers and publish them to your Docker registry

```shell
DOCKER_REGISTRY=<your docker registry> make push-docker
```

### 2. Install VMClarity CloudFormation

1. Ensure you have an SSH key pair uploaded to AWS EC2
2. Go to CloudFormation -> Create Stack -> Upload template.
3. Upload the VMClarity.cfn
4. Follow the wizard through to the end
    - Set the `VMClarity Backend Container Image` and `VMClarity Scanner Container Image` parameters in the wizard to
    use custom images (from step 1.) for deployment.
    - Change the Asset Scan Delete Policy to `OnSuccess` or `Never` if debugging scanner VMs is required.
5. Wait for install to complete

### 3. Ensure that VMClarity backend is working correctly

1. Get the IP address from the CloudFormation stack's Output Tab
2. `ssh ubuntu@<ip address>`
3. Check the VMClarity Logs

   ```shell
   sudo journalctl -u vmclarity
   ```

## Performing an end-to-end test

1. Copy the example [scan-config.json](assets/scan-config.json) into the ubuntu user's home directory

   ```shell
   scp docs/assets/scan-config.json ubuntu@<ip address>:~/scan-config.json
   ```

2. Edit the scan-config.json

   a. Give the scan config a unique name

   b. Enable the different scan families you want:

    ```shell
    "scanFamiliesConfig": {
      "sbom": {
        "enabled": true
      },
      "vulnerabilities": {
        "enabled": true
      },
      "exploits": {
        "enabled": true
      }
    },
    ```

   c. Configure the scope of the test

      - By Region, VPC or Security group:

        ```shell
        "scope": "contains(assetInfo.location, '<name of region>/<name of vpc>') and contains(assetInfo.securityGroups, '{\"id\":\"<name of sec group>\"}')"
        ```

      - By tag:

        ```shell
        "scope": "contains(assetInfo.tags, '{\"key\":\"<key>\",\"value\":\"<value>\"}')"
        ```

      - All:

        ```shell
        "scope": ""
        ```

   d. Set operationTime to the time you want the scan to run. As long as the time
      is in the future it can be within seconds.

3. While ssh'd into the VMClarity server run

   ```shell
   curl -X POST http://localhost:8080/api/scanConfigs -H 'Content-Type: application/json' -d @scan-config.json
   ```

4. Check VMClarity logs to ensure that everything is performing as expected

   ```shell
   sudo journalctl -u vmclarity
   ```

5. Monitor the asset scans

   - Get scans:

     ```shell
     curl -X GET http://localhost:8080/api/scans
     ```

     After the operationTime in the scan config created above there should be a new
     scan object created in Pending.

     Once discovery has been performed, the scan's assetIDs list should be
     populated will all the assets to be scanned by this scan.

     The scan will then create all the "assetScans" for tracking the scan
     process for each asset. When that is completed the scan will move to
     "InProgress".

   - Get asset scans:

     ```shell
     curl -X GET http://localhost:8080/api/assetScans
     ```
