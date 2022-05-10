SHELL=/bin/bash

# Project variables
BINARY_NAME ?= kubeclarity
DOCKER_REGISTRY ?= ghcr.io/openclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= ${VERSION}

# Dependency versions
GOLANGCI_VERSION = 1.42.0
LICENSEI_VERSION = 0.5.0

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: build
build: ui build-all-go ## Build All

.PHONY: build-all-go
build-all-go: cli backend sbom-db runtime-k8s-scanner cis-docker-benchmark-scanner ## Build All GO executables

.PHONY: ui
ui: ## Build UI
	@(echo "Building UI ..." )
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build

.PHONY: cli
cli: ## Build CLI
	@(echo "Building CLI ..." )
	@(cd cli && CGO_ENABLED=0 go build -ldflags "-X github.com/openclarity/kubeclarity/cli/pkg.GitRevision=${VERSION}" -o bin/cli ./main.go && ls -l bin/)

.PHONY: backend
backend: ## Build Backend
	@(echo "Building Backend ..." )
	@(cd backend && go build -o bin/backend cmd/backend/main.go && ls -l bin/)

.PHONY: sbom-db
sbom-db: ## Build SBOM DB
	@(echo "Building SBOM DB ..." )
	@(cd sbom_db/backend && go build -o bin/sbom_db cmd/main.go && ls -l bin/)

.PHONY: runtime-k8s-scanner
runtime-k8s-scanner: ## Build Runtime K8s Scanner
	@(echo "Building Runtime K8s Scanner ..." )
	@(cd runtime_k8s_scanner && go build -o bin/runtime_k8s_scanner cmd/main.go && ls -l bin/)

.PHONY: cis-docker-benchmark-scanner
cis-docker-benchmark-scanner: ## Build CIS Docker Benchmark Scanner
	@(echo "Building CIS Docker Benchmark Scanner ..." )
	@(cd cis_docker_benchmark_scanner && CGO_ENABLED=0 go build -o bin/cis_docker_benchmark_scanner cmd/main.go && ls -l bin/)

.PHONY: api
api: api-backend api-sbom-db api-runtime-scan

.PHONY: api-backend
api-backend: ## Generating Backend API code
	@(echo "Generating API code ..." )
	@(cd api; ./generate.sh)

.PHONY: api-runtime-scan
api-runtime-scan: ## Generating runtime scan API code
	@(echo "Generating runtime scan API code ..." )
	@(cd runtime_scan/api; ./generate.sh)

.PHONY: api-sbom-db
api-sbom-db: ## Generating SBOM DB API code
	@(echo "Generating SBOM DB API code ..." )
	@(cd sbom_db/api; ./generate.sh)

.PHONY: docker
docker: docker-backend docker-cli docker-sbom-db docker-runtime-k8s-scanner docker-cis-docker-benchmark-scanner ## Build All Docker images

.PHONY: push-docker
push-docker: push-docker-backend push-docker-cli push-docker-sbom-db push-docker-runtime-k8s-scanner push-docker-cis-docker-benchmark-scanner ## Build and Push All Docker images

.PHONY: docker-sbom-db
docker-sbom-db: ## Build SBOM DB Backend Docker image
	@(echo "Building SBOM DB backend docker image ..." )
	docker build --file ./Dockerfile.sbom_db --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-sbom-db:${DOCKER_TAG} .

.PHONY: push-docker-sbom-db
push-docker-sbom-db: docker-sbom-db ## Build and Push SBOM DB Backend Docker image
	@echo "Publishing SBOM DB backend Docker image ..."
	docker push ${DOCKER_IMAGE}-sbom-db:${DOCKER_TAG}

.PHONY: docker-backend
docker-backend: ## Build Backend Docker image
	@(echo "Building backend docker image ..." )
	docker build --file ./Dockerfile.backend --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}:${DOCKER_TAG} .

.PHONY: push-docker-backend
push-docker-backend: docker-backend ## Build and Push Backend Docker image
	@echo "Publishing backend docker image ..."
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}

.PHONY: docker-cli
docker-cli: ## Build CLI Docker image
	@(echo "Building cli docker image ..." )
	docker build --file ./Dockerfile.cli --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-cli:${DOCKER_TAG} .

.PHONY: push-docker-cli
push-docker-cli: docker-cli ## Build and Push CLI Docker image
	@echo "Publishing backend docker image ..."
	docker push ${DOCKER_IMAGE}-cli:${DOCKER_TAG}

.PHONY: docker-runtime-k8s-scanner
docker-runtime-k8s-scanner: ## Build runtime k8s scanner Docker image
	@(echo "Building runtime k8s scanner docker image ..." )
	docker build --file ./Dockerfile.runtime_k8s_scanner --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-runtime-k8s-scanner:${DOCKER_TAG} .

.PHONY: push-docker-runtime-k8s-scanner
push-docker-runtime-k8s-scanner: docker-runtime-k8s-scanner ## Build and Push runtime k8s scanner image
	@echo "Publishing runtime k8s scanner docker image ..."
	docker push ${DOCKER_IMAGE}-runtime-k8s-scanner:${DOCKER_TAG}

.PHONY: docker-cis-docker-benchmark-scanner
docker-cis-docker-benchmark-scanner: ## Build CIS docker benchmark scanner Docker image
	@(echo "Building CIS docker benchmark scanner docker image ..." )
	docker build --file ./Dockerfile.cis_docker_benchmark_scanner --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-cis-docker-benchmark-scanner:${DOCKER_TAG} .

.PHONY: push-docker-cis-docker-benchmark-scanner
push-docker-cis-docker-benchmark-scanner: docker-cis-docker-benchmark-scanner ## Build and Push CIS docker benchmark scanner image
	@echo "Publishing CIS docker benchmark scanner docker image ..."
	docker push ${DOCKER_IMAGE}-cis-docker-benchmark-scanner:${DOCKER_TAG}

.PHONY: test
test: ## Run Unit Tests
	@(cd backend && go test ./...)
	@(cd cli && CGO_ENABLED=0 go test ./...)
	@(cd shared && go test ./...)
	@(cd runtime_scan && go test ./...)
	@(cd sbom_db/backend && go test ./...)
	@(cd runtime_k8s_scanner && go test ./...)
	@(cd cis_docker_benchmark_scanner && GO111MODULE=on CGO_ENABLED=0 go test ./...)

.PHONY: clean
clean: clean-ui clean-backend clean-cli clean-runtime-k8s-scanner clean-cis-docker-benchmark-scanner ## Clean all build artifacts

.PHONY: clean-ui
clean-ui:
	@(rm -rf ui/build ; echo "UI cleanup done" )

.PHONY: clean-backend
clean-backend:
	@(rm -rf backend/bin ; echo "Backend cleanup done" )

.PHONY: clean-cli
clean-cli:
	@(rm -rf cli/bin ; echo "Cli cleanup done" )

.PHONY: clean-sbom-db
clean-sbom-db:
	@(rm -rf sbom_db/backend/bin ; echo "SBOM DB cleanup done" )

.PHONY: clean-runtime-k8s-scanner
clean-runtime-k8s-scanner:
	@(rm -rf runtime_k8s_scanner/bin ; echo "Runtime K8s Scanner cleanup done" )

.PHONY: clean-cis-docker-benchmark-scanner
clean-cis-docker-benchmark-scanner:
	@(rm -rf cis_docker_benchmark_scanner/bin ; echo "CIS docker benchmark Scanner cleanup done" )

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b ./bin/ v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	cd backend && ../bin/golangci-lint run
	cd cli && ../bin/golangci-lint run
	cd runtime_scan && ../bin/golangci-lint run
	cd sbom_db/backend && ../../bin/golangci-lint run
	cd runtime_k8s_scanner && ../bin/golangci-lint run
	cd cis_docker_benchmark_scanner && ../bin/golangci-lint run
	cd shared && ../bin/golangci-lint run

.PHONY: fix
fix: bin/golangci-lint ## Fix lint violations
	cd backend && ../bin/golangci-lint run --fix
	cd cli && ../bin/golangci-lint run --fix
	cd runtime_scan && ../bin/golangci-lint run --fix
	cd sbom_db/backend && ../../bin/golangci-lint run --fix
	cd runtime_k8s_scanner && ../bin/golangci-lint run --fix
	cd cis_docker_benchmark_scanner && ../bin/golangci-lint run --fix
	cd shared && ../bin/golangci-lint run --fix

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-check
license-check: bin/licensei ## Run license check
	./bin/licensei header
#	cd backend && ../bin/licensei check --config=../.licensei.toml
#	cd cli && ../bin/licensei check --config=../.licensei.toml
#	cd runtime_scan && ../bin/licensei check --config=../.licensei.toml
#	cd sbom_db/backend && ../../bin/licensei check --config=../../.licensei.toml
#	cd runtime_k8s_scanner && ../bin/licensei check --config=../.licensei.toml
#	cd cis_docker_benchmark_scanner && ../bin/licensei check --config=../.licensei.toml
#	cd shared && ../bin/licensei check --config=../.licensei.toml

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	./bin/licensei cache

.PHONY: check
check: lint test ## Run tests and linters

.PHONY: gomod-tidy
gomod-tidy:
	cd backend && go mod tidy
	cd shared && go mod tidy
	cd cli && go mod tidy
	cd runtime_scan && go mod tidy
	cd runtime_k8s_scanner && go mod tidy
	cd cis_docker_benchmark_scanner && go mod tidy

.PHONY: e2e
e2e:
	@echo "Running e2e tests ..."
	cd e2e && export DOCKER_TAG=${DOCKER_TAG} && go test .
