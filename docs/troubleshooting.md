# Troubleshooting and debugging VMClarity

## Table of Contents

- [How to debug the Scanner VMs](#how-to-debug-the-scanner-vms)
  - [AWS](#debug-scanner-VM-on-AWS)

## How to debug the Scanner VMs

How to debug the Scanner VMs can differ per provider these are documented
below.

### Debug Scanner VM on AWS

On AWS VMClarity is configured to create the Scanner VMs with the same key-pair
that the VMClarity server has. The Scanner VMs run in a private network,
however the VMClarity Server can be used as a bastion/jump host to reach them
via SSH.

```
ssh -i <key-pair private key> -J ubuntu@<vmclarity server public IP> ubuntu@<scanner VM private IP address>
```

Once SSH access has been established, the status of the VM's start up
configuration can be debugged by checking the cloud-init logs:

```
sudo journalctl -u cloud-final
```

And the vmclarity-scanner service logs:

```
sudo journalctl -u vmclarity-scanner
```

