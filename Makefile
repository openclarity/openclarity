SHELL=/bin/bash

# Project variables
BINARY_NAME ?= vmclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_REGISTRY ?= ghcr.io/openclarity
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= ${VERSION}
VMCLARITY_TOOLS_BASE ?=

# Dependency versions
GOLANGCI_VERSION = 1.49.0
LICENSEI_VERSION = 0.5.0

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: build
build: build-all-go ## Build All

.PHONY: build-all-go
build-all-go: backend cli ## Build All GO executables

.PHONY: backend
backend: ## Build Backend
	@(echo "Building Backend ..." )
	@(cd backend && go build -o bin/backend cmd/backend/main.go && ls -l bin/)

.PHONY: cli
cli: ## Build CLI
	@(echo "Building CLI ..." )
	@(cd cli && go build -ldflags="-X 'github.com/openclarity/vmclarity/cli/pkg.GitRevision=${VERSION}'" -o bin/vmclarity main.go && ls -l bin/)

.PHONY: docker
docker: docker-backend docker-cli ## Build All Docker images

.PHONY: push-docker
push-docker: push-docker-backend push-docker-cli ## Build and Push All Docker images

ifneq ($(strip $(VMCLARITY_TOOLS_BASE)),)
VMCLARITY_TOOLS_CLI_DOCKER_ARG=--build-arg VMCLARITY_TOOLS_BASE=${VMCLARITY_TOOLS_BASE}
endif

.PHONY: docker-cli
docker-cli: ## Build CLI Docker image
	@(echo "Building cli docker image ..." )
	docker build --file ./Dockerfile.cli --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) ${VMCLARITY_TOOLS_CLI_DOCKER_ARG} \
		-t ${DOCKER_IMAGE}:${DOCKER_TAG} .

.PHONY: push-docker-cli
push-docker-cli: docker-cli ## Build and Push CLI Docker image
	@echo "Publishing cli docker image ..."
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}

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

.PHONY: test
test: ## Run Unit Tests
	@(cd backend && go test ./...)
	@(cd runtime_scan && go test ./...)

.PHONY: clean-backend
clean-backend:
	@(rm -rf backend/bin ; echo "Backend cleanup done" )

.PHONY: clean
clean: clean-backend ## Clean all build artifacts

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b ./bin/ v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	cd backend && ../bin/golangci-lint run
	cd runtime_scan && ../bin/golangci-lint run

.PHONY: fix
fix: bin/golangci-lint ## Fix lint violations
	./bin/golangci-lint run --fix
	cd backend && ../bin/golangci-lint run --fix
	cd runtime_scan && ../bin/golangci-lint run --fix

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-check
license-check: bin/licensei ## Run license check
	./bin/licensei header
	cd backend && ../bin/licensei check --config=../.licensei.toml
	cd runtime_scan && ../bin/licensei check --config=../.licensei.toml

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	cd backend && ../bin/licensei cache --config=../.licensei.toml
	cd runtime_scan && ../bin/licensei cache --config=../.licensei.toml

.PHONY: check
check: lint test ## Run tests and linters

.PHONY: gomod-tidy
gomod-tidy:
	cd backend && go mod tidy
	cd runtime_scan && go mod tidy

.PHONY: api
api: ## Generating API code
	@(echo "Generating API code ..." )
	@(cd api; go generate)
