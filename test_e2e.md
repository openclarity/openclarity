# Update and install the cloud formation

## Build the containers and publish them to your docker hub

```
DOCKER_REGISTRY=<your docker hub> make push-docker
```

## Update installation/aws/VMClarity.cfn

Update the cloud formation with the pushed docker images, for example:

```
@@ -123,7 +123,7 @@ Resources:
                     DATABASE_DRIVER=LOCAL
                     BACKEND_REST_ADDRESS=__BACKEND_REST_ADDRESS__
                     BACKEND_REST_PORT=8888
-                    SCANNER_CONTAINER_IMAGE=tehsmash/vmclarity-cli:dc2d75a10e5583e97f516be26fcdbb484f98d5c3
+                    SCANNER_CONTAINER_IMAGE=tehsmash/vmclarity-cli:9bba94334c1de1aeed63ed12de3784d561fc4f1b
                   - JobImageID: !FindInMap
                       - AWSRegionArch2AMI
                       - !Ref "AWS::Region"
@@ -145,13 +145,13 @@ Resources:
                 ExecStartPre=-/usr/bin/docker stop %n
                 ExecStartPre=-/usr/bin/docker rm %n
                 ExecStartPre=/usr/bin/mkdir -p /opt/vmclarity
-                ExecStartPre=/usr/bin/docker pull tehsmash/vmclarity-backend:dc2d75a10e5583e97f516be26fcdbb484f98d5c3
+                ExecStartPre=/usr/bin/docker pull tehsmash/vmclarity-backend:9bba94334c1de1aeed63ed12de3784d561fc4f1b
                 ExecStart=/usr/bin/docker run \
                   --rm --name %n \
                   -p 0.0.0.0:8888:8888/tcp \
                   -v /opt/vmclarity:/data \
                   --env-file /etc/vmclarity/config.env \
-                  tehsmash/vmclarity-backend:dc2d75a10e5583e97f516be26fcdbb484f98d5c3 run --log-level info
+                  tehsmash/vmclarity-backend:9bba94334c1de1aeed63ed12de3784d561fc4f1b run --log-level info

                 [Install]
                 WantedBy=multi-user.target
```

# Go to AWS -> Cloudformation and create a stack.

* Ensure you have an SSH key pair uploaded to AWS Ec2
* Go to CloudFormation -> Create Stack -> From Template.
* Upload the modified VMClarity.cfn
* Follow the wizard through to the end
* Wait for install to complete

# Ssh to the VMClarity server

* Get the IP address from the CloudFormation stack's Output Tab
* `ssh ubuntu@<ip address>`
* Check the VMClarity Logs
  ```
  sudo journalctl -u vmclarity
  ```

# Create Scan Config

1. Copy the scanConfig.json into the ubuntu user's home directory

   ```
   scp scanConfig.json ubuntu@<ip address>:~/scanConfig.json
   ```

2. Edit the scanConfig.json

   a. Give the scan config a unique name

   b. Enable the different scan families you want:

    ```
    "scanFamiliesConfig": {
      "sbom": {
        "enabled": true
      },
      "vulnerabilties": {
        "enabled": true
      },
      "exploits": {
        "enabled": true
      }
    },
    ```

   c. Configure the scope of the test

      * By Region, VPC or Security group:

        ```
        "scope" {
          "objectType": "AwsScanScope",
          "regions": [
            {
             "name": "eu-west-1",
             "vpcs": [
               {
                 "name": "<name of vpc>",
                 "securityGroups": [
                   {
                     "name": "<name of sec group>"
                   }
                 ]
               }
             ]
            }
          ]
        }
        ```

      * By tag:

        ```
        "scope": {
          "instanceTagSelector": [
            {
              "key": "<key>",
              "value": "<value>"
            }
          ]
        }
        ```

      * All:

        ```
        "scope": {
          "all": true
        }
        ```

   d. Set operationTime to the time you want the scan to run. As long as the time
      is in the future it can be within seconds.

3. While ssh'd into the VMClarity server run

   ```
   curl -X POST http://localhost:8888/api/scanConfigs -H 'Content-Type: application/json' -d @scanConfig.json
   ```

4. Watch the VMClarity logs again

   ```
   sudo journalctl -u vmclarity -f
   ```

5. Monitor the scan results

   * Get scans:

     ```
     curl -X GET http://localhost:8888/api/scans
     ```

     After the operationTime in the scan config created above there should be a new
     scan object created in Pending.

     Once discovery has been performed, the scan's "targets" list should be
     populated will all the targets to be scanned by this scan.

     The scan will then create all the "scanResults" for tracking the scan
     process for each target. When that is completed the scan will move to
     "InProgress".

   * Get Scan Results:

     ```
     curl -X GET http://localhost:8888/api/scanResults
     ```
