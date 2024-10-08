#!/bin/bash

set -euo pipefail

mkdir -p /etc/openclarity
mkdir -p /opt/openclarity

cat << 'EOF' > /etc/openclarity/deploy.sh
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
  # Configure the OpenClarity backend to use the local postgres
  # service
  echo "OPENCLARITY_APISERVER_DATABASE_DRIVER=POSTGRES" > /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_NAME=openclarity" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_USER=openclarity" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_PASS=__PostgresDBPassword__" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_HOST=postgresql" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_PORT=5432" >> /etc/openclarity/apiserver.env
elif [ "__DatabaseToUse__" == "External Postgresql" ]; then
  # Configure the OpenClarity backend to use the postgres
  # database configured by the user.
  echo "OPENCLARITY_APISERVER_DATABASE_DRIVER=POSTGRES" > /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_NAME=__ExternalDBName__" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_USER=__ExternalDBUsername__" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_PASS=__ExternalDBPassword__" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_HOST=__ExternalDBHost__" >> /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_DB_PORT=__ExternalDBPort__" >> /etc/openclarity/apiserver.env
elif [ "__DatabaseToUse__" == "SQLite" ]; then
  # Configure the OpenClarity backend to use the SQLite DB
  # driver and configure the storage location so that it
  # persists.
  echo "OPENCLARITY_APISERVER_DATABASE_DRIVER=LOCAL" > /etc/openclarity/apiserver.env
  echo "OPENCLARITY_APISERVER_LOCAL_DB_PATH=/data/openclarity.db" >> /etc/openclarity/apiserver.env
fi

# Replace anywhere in the config.env __CONTROLPLANE_HOST__
# with the local ipv4 IP address of the OpenClarity server.
local_ip_address="$(curl -s -H Metadata:true --noproxy "*" "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2021-02-01&format=text")"
sed -i "s/__CONTROLPLANE_HOST__/${local_ip_address}/" /etc/openclarity/orchestrator.env

# Reload the systemd daemon to ensure that the OpenClarity unit
# has been detected.
systemctl daemon-reload

# Create directory required for grype-server
/usr/bin/mkdir -p /opt/grype-server
/usr/bin/chown -R 1000:1000 /opt/grype-server

# Create directory required for openclarity apiserver
/usr/bin/mkdir -p /opt/openclarity

# Create directory for exploit db server
/usr/bin/mkdir -p /opt/exploits

# Create directory for trivy server
/usr/bin/mkdir -p /opt/trivy-server

# Create directory for yara rule server
/usr/bin/mkdir -p /opt/yara-rule-server

# Enable and start/restart OpenClarity backend
systemctl enable openclarity.service
systemctl restart openclarity.service

# Add admin user to docker group and activate the changes
usermod -a -G docker __AdminUsername__
EOF
chmod 744 /etc/openclarity/deploy.sh

cat << 'EOF' > /etc/openclarity/yara-rule-server.yaml
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
chmod 644 /etc/openclarity/yara-rule-server.yaml

cat << 'EOF' > /etc/openclarity/orchestrator.env
OPENCLARITY_ORCHESTRATOR_PROVIDER=Azure
OPENCLARITY_AZURE_SUBSCRIPTION_ID=__AZURE_SUBSCRIPTION_ID__
OPENCLARITY_AZURE_SCANNER_LOCATION=__AZURE_SCANNER_LOCATION__
OPENCLARITY_AZURE_SCANNER_RESOURCE_GROUP=__AZURE_SCANNER_RESOURCE_GROUP__
OPENCLARITY_AZURE_SCANNER_SUBNET_ID=__AZURE_SCANNER_SUBNET_ID__
OPENCLARITY_AZURE_SCANNER_PUBLIC_KEY=__AZURE_SCANNER_PUBLIC_KEY__
OPENCLARITY_AZURE_SCANNER_VM_ARCHITECTURE_TO_SIZE_MAPPING=__AZURE_SCANNER_VM_ARCHITECTURE_TO_SIZE_MAPPING__
OPENCLARITY_AZURE_SCANNER_VM_ARCHITECTURE=__AZURE_SCANNER_VM_ARCHITECTURE__
OPENCLARITY_AZURE_SCANNER_IMAGE_PUBLISHER=__AZURE_SCANNER_IMAGE_PUBLISHER__
OPENCLARITY_AZURE_SCANNER_IMAGE_OFFER=__AZURE_SCANNER_IMAGE_OFFER__
OPENCLARITY_AZURE_SCANNER_VM_ARCHITECTURE_TO_IMAGE_SKU_MAPPING=__AZURE_SCANNER_VM_ARCHITECTURE_TO_IMAGE_SKU_MAPPING__
OPENCLARITY_AZURE_SCANNER_IMAGE_VERSION=__AZURE_SCANNER_IMAGE_VERSION__
OPENCLARITY_AZURE_SCANNER_SECURITY_GROUP=__AZURE_SCANNER_SECURITY_GROUP__
OPENCLARITY_AZURE_SCANNER_STORAGE_ACCOUNT_NAME=__AZURE_SCANNER_STORAGE_ACCOUNT_NAME__
OPENCLARITY_AZURE_SCANNER_STORAGE_CONTAINER_NAME=__AZURE_SCANNER_STORAGE_CONTAINER_NAME__

OPENCLARITY_ORCHESTRATOR_APISERVER_ADDRESS=http://apiserver:8888
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_CONTAINER_IMAGE=__ScannerContainerImage__
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_APISERVER_ADDRESS=http://__CONTROLPLANE_HOST__:8888
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_EXPLOITSDB_ADDRESS=http://__CONTROLPLANE_HOST__:1326
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_TRIVY_SERVER_ADDRESS=http://__CONTROLPLANE_HOST__:9992
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_GRYPE_SERVER_ADDRESS=__CONTROLPLANE_HOST__:9991
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_YARA_RULE_SERVER_ADDRESS=http://__CONTROLPLANE_HOST__:9993
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_DELETE_POLICY=__AssetScanDeletePolicy__
OPENCLARITY_ORCHESTRATOR_ASSETSCAN_WATCHER_SCANNER_FRESHCLAM_MIRROR=http://__CONTROLPLANE_HOST__:1000/clamav
EOF
chmod 644 /etc/openclarity/orchestrator.env

cat << 'EOF' > /etc/openclarity/openclarity.yaml
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
        source: /opt/openclarity
        target: /data
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure
    healthcheck:
      test: wget --no-verbose --tries=1 --spider http://127.0.0.1:8081/healthz/ready || exit 1
      interval: 10s
      retries: 60

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
    depends_on:
      apiserver:
        condition: service_healthy
    healthcheck:
      test: wget --no-verbose --tries=1 --spider http://127.0.0.1:8082/healthz/ready || exit 1
      interval: 10s
      retries: 60

  ui:
    image: __UIContainerImage__
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure
    depends_on:
      apiserver:
        condition: service_healthy

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
    depends_on:
      apiserver:
        condition: service_healthy
    healthcheck:
      test: wget --no-verbose --tries=1 --spider http://127.0.0.1:8083/healthz/ready || exit 1
      interval: 10s
      retries: 60

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
    healthcheck:
      test: ["CMD", "nc", "-z", "127.0.0.1", "1326"]
      interval: 10s
      retries: 60

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
    healthcheck:
      test: ["CMD", "nc", "-z", "127.0.0.1", "9992"]
      interval: 10s
      retries: 60

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
    healthcheck:
      test: wget --no-verbose --tries=10 --spider http://127.0.0.1:8080/healthz/ready || exit 1
      interval: 10s
      retries: 60

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
    healthcheck:
      test: wget --no-verbose --tries=1 --spider http://127.0.0.1:8082/healthz/ready || exit 1
      interval: 10s
      retries: 60

  swagger-ui:
    image: swaggerapi/swagger-ui:v5.17.14
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

touch /etc/openclarity/openclarity.override.yaml
# shellcheck disable=SC2050
if [ "__DatabaseToUse__" == "Postgresql" ]; then
  cat << 'EOF' > /etc/openclarity/openclarity.override.yaml
services:
  postgresql:
    image: __PostgresqlContainerImage__
    env_file: ./postgres.env
    ports:
      - "5432:5432"
    logging:
      driver: journald
    deploy:
      mode: replicated
      replicas: 1
      restart_policy:
        condition: on-failure
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -d openclarity -U openclarity"]
      interval: 10s
      retries: 60

  apiserver:
    depends_on:
      postgresql:
        condition: service_healthy
EOF
fi

cat << 'EOF' > /etc/openclarity/swagger-config.json
{
    "urls": [
        {
            "name": "OpenClarity API",
            "url": "/api/openapi.json"
        }
    ]
}
EOF
chmod 644 /etc/openclarity/swagger-config.json

cat << 'EOF' > /etc/openclarity/uibackend.env
##
## UIBackend configuration
##
# OpenClarity API server address
OPENCLARITY_UIBACKEND_APISERVER_ADDRESS=http://apiserver:8888
EOF
chmod 644 /etc/openclarity/uibackend.env

cat << 'EOF' > /etc/openclarity/trivy-server.env
TRIVY_LISTEN=0.0.0.0:9992
TRIVY_CACHE_DIR=/home/scanner/.cache/trivy
EOF
chmod 644 /etc/openclarity/trivy-server.env

cat << 'EOF' > /etc/openclarity/postgres.env
POSTGRESQL_USERNAME=openclarity
POSTGRESQL_PASSWORD=__PostgresDBPassword__
POSTGRESQL_DATABASE=openclarity
EOF
chmod 644 /etc/openclarity/postgres.env

cat << 'EOF' > /etc/openclarity/gateway.conf
events {
    worker_connections 1024;
}

http {
    upstream ui {
        server ui:8080;
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
chmod 644 /etc/openclarity/gateway.conf

cat << 'EOF' > /lib/systemd/system/openclarity.service
[Unit]
Description=OpenClarity
After=docker.service
Requires=docker.service

[Service]
TimeoutStartSec=0
Type=oneshot
RemainAfterExit=true
ExecStart=/usr/bin/docker compose -p openclarity -f /etc/openclarity/openclarity.yaml -f /etc/openclarity/openclarity.override.yaml up -d --wait --remove-orphans
ExecStop=/usr/bin/docker compose -p openclarity -f /etc/openclarity/openclarity.yaml -f /etc/openclarity/openclarity.override.yaml down

[Install]
WantedBy=multi-user.target
EOF
chmod 644 /lib/systemd/system/openclarity.service

/etc/openclarity/deploy.sh
