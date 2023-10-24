#!/bin/bash

set -euo pipefail

mkdir -p /etc/vmclarity
mkdir -p /opt/vmclarity

cat << 'EOF' > /etc/vmclarity/deploy.sh
#!/bin/bash
set -euo pipefail

# Install the latest version of docker from the offical
# docker repository instead of the older version built into
# ubuntu, so that we can use docker compose v2.
#
# To install this we need to add the docker apt repo gpg key
# to the apt keyring, and then add the apt sources based on
# our version of ubuntu. Then we can finally apt install all
# the required docker components.
apt-get update
apt-get install -y ca-certificates curl gnupg
mkdir -p /etc/apt/keyrings
chmod 755 /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --yes --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

if [ "__DatabaseToUse__" == "Postgresql" ]; then
  # Enable and start/restart postgres
  echo "COMPOSE_PROFILES=postgres" >> /etc/vmclarity/service.env

  # Configure the VMClarity backend to use the local postgres
  # service
  echo "DATABASE_DRIVER=POSTGRES" > /etc/vmclarity/apiserver.env
  echo "DB_NAME=vmclarity" >> /etc/vmclarity/apiserver.env
  echo "DB_USER=vmclarity" >> /etc/vmclarity/apiserver.env
  echo "DB_PASS=__PostgresDBPassword__" >> /etc/vmclarity/apiserver.env
  echo "DB_HOST=postgresql" >> /etc/vmclarity/apiserver.env
  echo "DB_PORT_NUMBER=5432" >> /etc/vmclarity/apiserver.env
elif [ "__DatabaseToUse__" == "External Postgresql" ]; then
  # Configure the VMClarity backend to use the postgres
  # database configured by the user.
  echo "DATABASE_DRIVER=POSTGRES" > /etc/vmclarity/apiserver.env
  echo "DB_NAME=__ExternalDBName__" >> /etc/vmclarity/apiserver.env
  echo "DB_USER=__ExternalDBUsername__" >> /etc/vmclarity/apiserver.env
  echo "DB_PASS=__ExternalDBPassword__" >> /etc/vmclarity/apiserver.env
  echo "DB_HOST=__ExternalDBHost__" >> /etc/vmclarity/apiserver.env
  echo "DB_PORT_NUMBER=__ExternalDBPort__" >> /etc/vmclarity/apiserver.env
elif [ "__DatabaseToUse__" == "SQLite" ]; then
  # Configure the VMClarity backend to use the SQLite DB
  # driver and configure the storage location so that it
  # persists.
  echo "DATABASE_DRIVER=LOCAL" > /etc/vmclarity/apiserver.env
  echo "LOCAL_DB_PATH=/data/vmclarity.db" >> /etc/vmclarity/apiserver.env
fi

# Replace anywhere in the config.env __CONTROLPLANE_HOST__
# with the local ipv4 IP address of the VMClarity server.
local_ip_address="$(curl -s -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2021-02-01&format=text")"
sed -i "s/__CONTROLPLANE_HOST__/${local_ip_address}/" /etc/vmclarity/orchestrator.env

# Reload the systemd daemon to ensure that the VMClarity unit
# has been detected.
systemctl daemon-reload

# Create directory required for grype-server
/usr/bin/mkdir -p /opt/grype-server
/usr/bin/chown -R 1000:1000 /opt/grype-server

# Create directory required for vmclarity apiserver
/usr/bin/mkdir -p /opt/vmclarity

# Create directory for exploit db server
/usr/bin/mkdir -p /opt/exploits

# Create directory for trivy server
/usr/bin/mkdir -p /opt/trivy-server

# Create directory for yara rule server
/usr/bin/mkdir -p /opt/yara-rule-server

# Enable and start/restart VMClarity backend
systemctl enable vmclarity.service
systemctl restart vmclarity.service
EOF
chmod 744 /etc/vmclarity/deploy.sh

cat << 'EOF' > /etc/vmclarity/yara-rule-server.yaml
enable_json_log: true
rule_update_schedule: "0 0 * * *"
rule_sources:
  - name: "base"
    url: "https://github.com/Yara-Rules/rules/archive/refs/heads/master.zip"
    exclude_regex: ".*index.*.yar|.*/utils/.*|.*/deprecated/.*|.*index_.*|.*MALW_AZORULT.yar"
  - name: "magic"
    url: "https://github.com/securitymagic/yara/archive/refs/heads/main.zip"
    exclude_regex: ".*index.*.yar"
EOF
chmod 644 /etc/vmclarity/yara-rule-server.yaml

cat << 'EOF' > /etc/vmclarity/orchestrator.env
VMCLARITY_ORCHESTRATOR_PROVIDER=Azure
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

VMCLARITY_ORCHESTRATOR_APISERVER_ADDRESS=http://apiserver:8888
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_CONTAINER_IMAGE=__ScannerContainerImage__
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_APISERVER_ADDRESS=http://__CONTROLPLANE_HOST__:8888
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_EXPLOITSDB_ADDRESS=http://__CONTROLPLANE_HOST__:1326
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_TRIVY_SERVER_ADDRESS=http://__CONTROLPLANE_HOST__:9992
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_GRYPE_SERVER_ADDRESS=__CONTROLPLANE_HOST__:9991
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_YARA_RULE_SERVER_ADDRESS=http://__CONTROLPLANE_HOST__:9993
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_DELETE_POLICY=__AssetScanDeletePolicy__
VMCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_FRESHCLAM_MIRROR=http://__CONTROLPLANE_HOST__:1000/clamav
EOF
chmod 644 /etc/vmclarity/orchestrator.env

cat << 'EOF' > /etc/vmclarity/vmclarity.yaml
version: '3'

services:
  apiserver:
    image: __APIServerContainerImage__
    command:
      - run
      - --log-level
      - info
    ports:
      - "8888:8888"
    env_file: ./apiserver.env
    volumes:
      - type: bind
        source: /opt/vmclarity
        target: /data
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  orchestrator:
    image: __OrchestratorContainerImage__
    command:
      - run
      - --log-level
      - info
    env_file: ./orchestrator.env
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  ui:
    image: __UIContainerImage__
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  uibackend:
    image: __UIBackendContainerImage__
    command:
      - run
      - --log-level
      - info
    env_file: ./uibackend.env
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  gateway:
    image: nginx
    ports:
      - "80:80"
    configs:
      - source: gateway_config
        target: /etc/nginx/nginx.conf
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  exploit-db-server:
    image: __ExploitDBServerContainerImage__
    ports:
      - "1326:1326"
    volumes:
      - type: bind
        source: /opt/exploits
        target: /vuls
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  trivy-server:
    image: __TrivyServerContainerImage__
    command:
      - server
    ports:
      - "9992:9992"
    env_file: ./trivy-server.env
    volumes:
      - type: bind
        source: /opt/trivy-server
        target: /home/scanner/.cache
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  grype-server:
    image: __GrypeServerContainerImage__
    command:
      - run
      - --log-level
      - warning
    ports:
      - "9991:9991"
    volumes:
      - type: bind
        source: /opt/grype-server
        target: /data
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  freshclam-mirror:
    image: __FreshclamMirrorContainerImage__
    ports:
      - "1000:80"
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  yara-rule-server:
    image: __YaraRuleServerContainerImage__
    command:
      - run
    ports:
      - "9993:8080"
    configs:
      - source: yara_rule_server_config
        target: /etc/yara-rule-server/config.yaml
    volumes:
      - type: bind
        source: /opt/yara-rule-server
        target: /var/lib/yara-rule-server
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  postgresql:
    image: __PostgresqlContainerImage__
    env_file: ./postgres.env
    ports:
      - "5432:5432"
    profiles:
      - postgres
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure

  swagger-ui:
    image: swaggerapi/swagger-ui:v5.3.1
    environment:
      CONFIG_URL: /apidocs/swagger-config.json
    configs:
      - source: swagger_config
        target: /usr/share/nginx/html/swagger-config.json

configs:
  gateway_config:
    file: ./gateway.conf
  swagger_config:
    file: ./swagger-config.json
  yara_rule_server_config:
    file: ./yara-rule-server.yaml
EOF

cat << 'EOF' > /etc/vmclarity/swagger-config.json
{
    "urls": [
        {
            "name": "VMClarity API",
            "url": "/api/openapi.json"
        }
    ]
}
EOF
chmod 644 /etc/vmclarity/swagger-config.json

cat << 'EOF' > /etc/vmclarity/uibackend.env
##
## UIBackend configuration
##
# Host for the VMClarity backend server
APISERVER_HOST=apiserver
# Port number for the VMClarity backend server
APISERVER_PORT=8888
EOF
chmod 644 /etc/vmclarity/uibackend.env

cat << 'EOF' > /etc/vmclarity/service.env
# COMPOSE_PROFILES=
EOF
chmod 644 /etc/vmclarity/service.env

cat << 'EOF' > /etc/vmclarity/trivy-server.env
TRIVY_LISTEN=0.0.0.0:9992
TRIVY_CACHE_DIR=/home/scanner/.cache/trivy
EOF
chmod 644 /etc/vmclarity/trivy-server.env

cat << 'EOF' > /etc/vmclarity/postgres.env
POSTGRESQL_USERNAME=vmclarity
POSTGRESQL_PASSWORD=__PostgresDBPassword__
POSTGRESQL_DATABASE=vmclarity
EOF
chmod 644 /etc/vmclarity/postgres.env

cat << 'EOF' > /etc/vmclarity/gateway.conf
events {
    worker_connections 1024;
}

http {
    upstream ui {
        server ui:80;
    }

    upstream uibackend {
        server uibackend:8890;
    }

    upstream apiserver {
        server apiserver:8888;
    }

    server {
        listen 80;
        absolute_redirect off;

        location / {
            proxy_pass http://ui/;
        }

        location /ui/api/ {
            proxy_pass http://uibackend/;
        }

        location /api/ {
            proxy_set_header X-Forwarded-Host $http_host;
            proxy_set_header X-Forwarded-Prefix /api;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_pass http://apiserver/;
        }

        location /apidocs/ {
            proxy_pass http://swagger-ui:8080/;
        }
    }
}
EOF
chmod 644 /etc/vmclarity/gateway.conf

cat << 'EOF' > /lib/systemd/system/vmclarity.service
[Unit]
Description=VmClarity
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Type=oneshot
RemainAfterExit=true
EnvironmentFile=/etc/vmclarity/service.env
ExecStart=/usr/bin/docker compose -p vmclarity -f /etc/vmclarity/vmclarity.yaml up -d --wait --remove-orphans
ExecStop=/usr/bin/docker compose -p vmclarity -f /etc/vmclarity/vmclarity.yaml down

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/vmclarity.service

/etc/vmclarity/deploy.sh
