# Troubleshooting and debugging VMClarity

## Table of Contents

- [How to debug the Scanner VMs](#how-to-debug-the-scanner-vms)
  - [Docker and Kubernetes](#docker-and-kubernetes-provider)
  - [Cloud providers](#cloud-providers)

## How to debug the Scanner VMs

### Docker and Kubernetes provider

For Docker provider, scanners are created as containers, while as pods in case of Kubernetes. In both cases, you can
access them directly and check the logs.

### Cloud providers

On cloud providers (AWS, Azure, GCP) VMClarity is configured to create the Scanner VMs with the same key-pair that the
VMClarity server has. The Scanner VMs run in a private network, however the VMClarity Server can be used as a
bastion/jump host to reach them via SSH.

```shell
ssh -i <key-pair private key> -J ubuntu@<vmclarity server public IP> ubuntu@<scanner VM private IP address>
```

Once SSH access has been established, the status of the VM's start up
configuration can be debugged by checking the cloud-init logs:

```shell
sudo journalctl -u cloud-final
```

And the vmclarity-scanner service logs:

```shell
sudo journalctl -u vmclarity-scanner
```
