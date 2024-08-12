# Installing VMClarity on GCP with Deployment Manager

1. Copy example configuration file to a new config

   ```
   cp vmclarity-config.example.yaml vmclarity-config.yaml
   ```

2. Edit configuration to add required fields.
   Check vmclarity.py.schema for other optional parameters.

3. Create a deployment using the gcloud CLI

   ```
   gcloud deployment-manager deployments create <vmclarity deployment name> --config vmclarity-config.yaml
   ```

4. Copy the VMClarity IP address from the output

5. SSH into the VMClarity server and open a tunnel to the UI

   ```
   ssh -L 8888:localhost:8888 vmclarity@<VMClarity server IP address>
   ```

6. To update the VMClarity configuration, modify the vmclarity-config.yaml, then update the deployment:

   ```
   gcloud deployment-manager deployments update <vmclarity deployment name> --config vmclarity-config.yaml
   ```
