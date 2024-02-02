## Testing with Vagrant

Vagrant is useful for testing locally the cloud-init functionality.
First we need to install Vagrant and Virtualbox.

```
brew install virtualbox virtualbox-extension-pack vagrant
```

Starting the instance with Vagrant locally:

```
cd scanner_boot_test/vagrant
VAGRANT_EXPERIMENTAL="cloud_init,disks" vagrant up
```

Starting an instance with cloud-init on AWS EC2:

```
aws ec2 run-instances --image-id ami-abcd1234 --count 1 --instance-type m3.medium \
--key-name my-key-pair --subnet-id subnet-abcd1234 --security-group-ids sg-abcd1234 \
--user-data file://cloud-config.cfg
```

On a running vagrant instance we can run the script manually for testing:
```
cd scanner_boot_test/vagrant
VAGRANT_EXPERIMENTAL="cloud_init,disks" vagrant up
vagrant ssh
cd /root
./scanscript.sh
```
