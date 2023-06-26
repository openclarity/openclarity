#!/bin/bash 

set -euo pipefail

apt-get update
apt-get install -y docker.io 

mkdir -p /etc/vmclarity
mkdir -p /opt/vmclarity

cat << 'EOF' > /etc/vmclarity/deploy.sh
#!/bin/bash
set -euo pipefail

# Create the docker network for the VMClarity services if it
# doesn't exist.
if docker network ls | grep vmclarity; then
  echo "network already exists"
else
  docker network create vmclarity
fi

# Reload the systemd daemon to ensure that all the VMClarity
# units have been detected.
systemctl daemon-reload

# Enable and start/restart exploit-db-server
systemctl enable exploit-db-server.service
systemctl restart exploit-db-server.service

# Enable and start/restart trivy server
systemctl enable trivy_server.service
systemctl restart trivy_server.service

# Enable and start/restart grype_server
systemctl enable grype_server.service
systemctl restart grype_server.service

# Enable and start/restart freshclam mirror
systemctl enable vmclarity_freshclam_mirror.service
systemctl restart vmclarity_freshclam_mirror.service

if [ "__DatabaseToUse__" == "Postgresql" ]; then
  # Enable and start/restart postgres
  systemctl enable postgres.service
  systemctl restart postgres.service

  # Configure the VMClarity backend to use the local postgres
  # service
  echo "DATABASE_DRIVER=POSTGRES" >> /etc/vmclarity/config.env
  echo "DB_NAME=vmclarity" >> /etc/vmclarity/config.env
  echo "DB_USER=vmclarity" >> /etc/vmclarity/config.env
  echo "DB_PASS=__PostgresDBPassword__" >> /etc/vmclarity/config.env
  echo "DB_HOST=postgres.service" >> /etc/vmclarity/config.env
  echo "DB_PORT_NUMBER=5432" >> /etc/vmclarity/config.env
elif [ "__DatabaseToUse__" == "External Postgresql" ]; then
  # Configure the VMClarity backend to use the postgres
  # database configured by the user.
  echo "DATABASE_DRIVER=POSTGRES" >> /etc/vmclarity/config.env
  echo "DB_NAME=__ExternalDBName__" >> /etc/vmclarity/config.env
  echo "DB_USER=__ExternalDBUsername__" >> /etc/vmclarity/config.env
  echo "DB_PASS=__ExternalDBPassword__" >> /etc/vmclarity/config.env
  echo "DB_HOST=__ExternalDBHost__" >> /etc/vmclarity/config.env
  echo "DB_PORT_NUMBER=__ExternalDBPort__" >> /etc/vmclarity/config.env
elif [ "__DatabaseToUse__" == "SQLite" ]; then
  # Configure the VMClarity backend to use the SQLite DB
  # driver and configure the storage location so that it
  # persists.
  echo "DATABASE_DRIVER=LOCAL" >> /etc/vmclarity/config.env
  echo "LOCAL_DB_PATH=/data/vmclarity.db" >> /etc/vmclarity/config.env
fi

# Replace anywhere in the config.env __BACKEND_REST_HOST__
# with the local ipv4 IP address of the VMClarity server.
local_ip_address="$(curl -s -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2021-02-01&format=text")"
sed -i "s/__BACKEND_REST_HOST__/${local_ip_address}/" /etc/vmclarity/config.env

# Enable and start/restart VMClarity backend
systemctl enable vmclarity.service
systemctl restart vmclarity.service
EOF
chmod 744 /etc/vmclarity/deploy.sh

cat << 'EOF' > /etc/vmclarity/config.env
PROVIDER=Azure
VMCLARITY_AZURE_SUBSCRIPTION_ID=__AZURE_SUBSCRIPTION_ID__
VMCLARITY_AZURE_SCANNER_LOCATION=__AZURE_SCANNER_LOCATION__
VMCLARITY_AZURE_SCANNER_RESOURCE_GROUP=__AZURE_SCANNER_RESOURCE_GROUP__
VMCLARITY_AZURE_SCANNER_SUBNET_ID=__AZURE_SCANNER_SUBNET_ID__
VMCLARITY_AZURE_SCANNER_PUBLIC_KEY=__AZURE_SCANNER_PUBLIC_KEY__
VMCLARITY_AZURE_SCANNER_VM_SIZE=__AZURE_SCANNER_VM_SIZE__
VMCLARITY_AZURE_SCANNER_IMAGE_PUBLISHER=__AZURE_SCANNER_IMAGE_PUBLISHER__
VMCLARITY_AZURE_SCANNER_IMAGE_OFFER=__AZURE_SCANNER_IMAGE_OFFER__
VMCLARITY_AZURE_SCANNER_IMAGE_SKU=__AZURE_SCANNER_IMAGE_SKU__
VMCLARITY_AZURE_SCANNER_IMAGE_VERSION=__AZURE_SCANNER_IMAGE_VERSION__
VMCLARITY_AZURE_SCANNER_SECURITY_GROUP=__AZURE_SCANNER_SECURITY_GROUP__
VMCLARITY_AZURE_SCANNER_STORAGE_ACCOUNT_NAME=__AZURE_SCANNER_STORAGE_ACCOUNT_NAME__
VMCLARITY_AZURE_SCANNER_STORAGE_CONTAINER_NAME=__AZURE_SCANNER_STORAGE_CONTAINER_NAME__

BACKEND_REST_HOST=__BACKEND_REST_HOST__
BACKEND_REST_PORT=8888
SCANNER_CONTAINER_IMAGE=__ScannerContainerImage__
TRIVY_SERVER_ADDRESS=http://__BACKEND_REST_HOST__:9992
GRYPE_SERVER_ADDRESS=__BACKEND_REST_HOST__:9991
DELETE_JOB_POLICY=__AssetScanDeletePolicy__
ALTERNATIVE_FRESHCLAM_MIRROR_URL=http://__BACKEND_REST_HOST__:1000/clamav
EOF
chmod 644 /etc/vmclarity/config.env 

cat << 'EOF' > /etc/vmclarity/service.env
BACKEND_CONTAINER_IMAGE=__BackendContainerImage__
BACKEND_LOG_LEVEL=info
EOF
chmod 644 /etc/vmclarity/service.env 

cat << 'EOF' > /lib/systemd/system/vmclarity.service
[Unit]
Description=VmClarity
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
EnvironmentFile=/etc/vmclarity/service.env
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=/usr/bin/mkdir -p /opt/vmclarity
ExecStartPre=/usr/bin/docker pull ${BACKEND_CONTAINER_IMAGE}
ExecStart=/usr/bin/docker run \
  --rm --name %n \
  --network vmclarity \
  -p 0.0.0.0:8888:8888/tcp \
  -v /opt/vmclarity:/data \
  --env-file /etc/vmclarity/config.env \
  ${BACKEND_CONTAINER_IMAGE} \
  run \
  --log-level ${BACKEND_LOG_LEVEL}

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/vmclarity.service

cat << 'EOF' > /lib/systemd/system/exploit-db-server.service
[Unit]
Description=Exploit DB Server
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=/usr/bin/mkdir -p /opt/exploits
ExecStartPre=/usr/bin/docker pull __ExploitDBServerContainerImage__
ExecStart=/usr/bin/docker run \
  --rm --name %n \
  --network vmclarity \
  -p 0.0.0.0:1326:1326/tcp \
  -v /opt/exploits:/vuls \
  __ExploitDBServerContainerImage__

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/exploit-db-server.service 

mkdir -p /etc/trivy-server

cat << 'EOF' > /etc/trivy-server/config.env
TRIVY_LISTEN=0.0.0.0:9992
TRIVY_CACHE_DIR=/home/scanner/.cache/trivy
EOF
chmod 644 /etc/trivy-server/config.env

cat << 'EOF' > /lib/systemd/system/trivy_server.service
[Unit]
Description=Trivy Server
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=/usr/bin/mkdir -p /opt/trivy-server
ExecStartPre=/usr/bin/docker pull __TrivyServerContainerImage__
ExecStart=/usr/bin/docker run \
  --rm --name %n \
  --network vmclarity \
  -p 0.0.0.0:9992:9992/tcp \
  -v /opt/trivy-server:/home/scanner/.cache \
  --env-file /etc/trivy-server/config.env \
  __TrivyServerContainerImage__ server

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/trivy_server.service 

mkdir -p /etc/grype-server

cat << 'EOF' > /etc/grype-server/config.env
DB_ROOT_DIR=/opt/grype-server/db
EOF
chmod 644 /etc/grype-server/config.env 

cat << 'EOF' > /lib/systemd/system/grype_server.service
[Unit]
Description=Grype Server
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=/usr/bin/mkdir -p /opt/grype-server
ExecStartPre=/usr/bin/chown -R 1000:1000 /opt/grype-server
ExecStartPre=/usr/bin/docker pull __GrypeServerContainerImage__
ExecStart=/usr/bin/docker run \
  --rm --name %n \
  --network vmclarity \
  -p 0.0.0.0:9991:9991/tcp \
  -v /opt/grype-server:/opt/grype-server \
  --env-file /etc/grype-server/config.env \
  __GrypeServerContainerImage__ run --log-level warning

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/grype_server.service

cat << 'EOF' > /lib/systemd/system/vmclarity_freshclam_mirror.service
[Unit]
Description=Deploys the freshclam mirror service
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=/usr/bin/docker pull __FreshclamMirrorContainerImage__
ExecStart=/usr/bin/docker run \
  --rm --name %n \
  --network vmclarity \
  -p 0.0.0.0:1000:80/tcp \
  __FreshclamMirrorContainerImage__

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/vmclarity_freshclam_mirror.service 

cat << 'EOF' > /lib/systemd/system/postgres.service
[Unit]
Description=Postgresql Database Server
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Restart=always
ExecStartPre=-/usr/bin/docker stop %n
ExecStartPre=-/usr/bin/docker rm %n
ExecStartPre=/usr/bin/docker pull __PostgresqlContainerImage__
ExecStart=/usr/bin/docker run \
  --rm --name %n \
  --network vmclarity \
  -e POSTGRESQL_USERNAME=vmclarity \
  -e POSTGRESQL_PASSWORD=__PostgresDBPassword__ \
  -e POSTGRESQL_DATABASE=vmclarity \
  -p 127.0.0.1:5432:5432/tcp \
  __PostgresqlContainerImage__

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/postgres.service

/etc/vmclarity/deploy.sh
