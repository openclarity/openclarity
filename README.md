# VMClarity
VMClarity is a tool for agentless detection and management of Virtual Machine
Software Bill Of Materials (SBOM) and vulnerabilities

To install vmclarity in your AWS account [Click Here](https://eu-central-1.console.aws.amazon.com/cloudformation/home?region=eu-central-1#/stacks/create/review?templateUrl=https://raw.githubusercontent.com/openclarity/vmclarity/main/installation/aws/VmClarity.cfn&stackName=VmClarity)

# How to debug the scanner VM
By default, the scanner VM instance is created w/o a key pair and a public ip, in order to set it follow the instructions below:
1. SSH into the VMClarity server VM
2. Update `/etc/vmclarity/config.env` with `SCANNER_KEY_PAIR_NAME=<your key pair name>`
3. Restart the VMClarity service (`sudo systemctl restart vmclarity.service`)
4. Create a new scan.
5. After the scanner VM was created, add an inbound rule to allow inbound SSH.
6. SSH into the scanner VM with your key pair.
