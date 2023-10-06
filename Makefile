SHELL=/bin/bash

# Project variables
BINARY_NAME ?= vmclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_REGISTRY ?= ghcr.io/openclarity
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= ${VERSION}
VMCLARITY_TOOLS_BASE ?=
GO_VERSION = 1.20

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := $(ROOT_DIR)/bin

# Dependency versions
LICENSEI_VERSION = 0.9.0

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_0-9-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: build
build: ui build-all-go ## Build All

.PHONY: build-all-go
build-all-go: bin/vmclarity-apiserver bin/vmclarity-cli bin/vmclarity-orchestrator bin/vmclarity-ui-backend ## Build All GO executables

.PHONY: ui
ui: ## Build UI
	@(echo "Building UI ..." )
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build

bin/vmclarity-orchestrator: $(shell find api) $(shell find cmd/vmclarity-orchestrator) $(shell find pkg) go.mod go.sum | $(BIN_DIR) ## Build vmclarity-orchestrator
	go build -race -o bin/vmclarity-orchestrator cmd/vmclarity-orchestrator/main.go

bin/vmclarity-apiserver: $(shell find api) $(shell find cmd/vmclarity-apiserver) $(shell find pkg) go.mod go.sum | $(BIN_DIR) ## Build vmclarity-apiserver
	go build -race -o bin/vmclarity-apiserver cmd/vmclarity-apiserver/main.go

bin/vmclarity-cli: $(shell find api) $(shell find cmd/vmclarity-cli) $(shell find pkg) go.mod go.sum | $(BIN_DIR) ## Build CLI
	go build -race -o bin/vmclarity-cli cmd/vmclarity-cli/main.go

bin/vmclarity-ui-backend: $(shell find api) $(shell find cmd/vmclarity-ui-backend) $(shell find pkg) go.mod go.sum | $(BIN_DIR) ## Build vmclarity-ui-backend
	go build -race -o bin/vmclarity-ui-backend cmd/vmclarity-ui-backend/main.go

.PHONY: docker
docker: docker-apiserver docker-cli docker-orchestrator docker-ui docker-ui-backend ## Build All Docker images

.PHONY: push-docker
push-docker: push-docker-apiserver push-docker-cli push-docker-orchestrator push-docker-ui push-docker-ui-backend ## Build and Push All Docker images

ifneq ($(strip $(VMCLARITY_TOOLS_BASE)),)
VMCLARITY_TOOLS_CLI_DOCKER_ARG=--build-arg VMCLARITY_TOOLS_BASE=${VMCLARITY_TOOLS_BASE}
endif

.PHONY: docker-cli
docker-cli: ## Build CLI Docker image
	@(echo "Building cli docker image ..." )
	docker build --file ./Dockerfile.cli --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		${VMCLARITY_TOOLS_CLI_DOCKER_ARG} \
		-t ${DOCKER_IMAGE}-cli:${DOCKER_TAG} \
		.

.PHONY: push-docker-cli
push-docker-cli: docker-cli ## Build and Push CLI Docker image
	@echo "Publishing cli docker image ..."
	docker push $(DOCKER_IMAGE)-cli:$(DOCKER_TAG)

.PHONY: docker-orchestrator
docker-orchestrator: ## Build Backend Orchestrator image
	@(echo "Building orchestrator docker image ..." )
	docker build --file ./Dockerfile.orchestrator --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-orchestrator:${DOCKER_TAG} .

.PHONY: push-docker-orchestrator
push-docker-orchestrator: docker-orchestrator ## Build and Push Orchestrator Docker image
	@echo "Publishing orchestrator docker image ..."
	docker push ${DOCKER_IMAGE}-orchestrator:${DOCKER_TAG}

.PHONY: docker-apiserver
docker-apiserver: ## Build Backend API Server image
	@(echo "Building apiserver docker image ..." )
	docker build --file ./Dockerfile.apiserver --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-apiserver:${DOCKER_TAG} .

.PHONY: push-docker-apiserver
push-docker-apiserver: docker-apiserver ## Build and Push API Server Docker image
	@echo "Publishing apiserver docker image ..."
	docker push ${DOCKER_IMAGE}-apiserver:${DOCKER_TAG}

.PHONY: docker-ui
docker-ui: ## Build UI image
	@(echo "Building ui docker image ..." )
	docker build --file ./Dockerfile.ui \
		-t ${DOCKER_IMAGE}-ui:${DOCKER_TAG} .

.PHONY: push-docker-ui
push-docker-ui: docker-ui ## Build and Push UI Docker image
	@echo "Publishing ui docker image ..."
	docker push ${DOCKER_IMAGE}-ui:${DOCKER_TAG}

.PHONY: docker-ui-backend
docker-ui-backend: ## Build UI Backend Docker image
	@(echo "Building ui-backend docker image ..." )
	docker build --file ./Dockerfile.uibackend --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}-ui-backend:${DOCKER_TAG} .

.PHONY: push-docker-ui-backend
push-docker-ui-backend: docker-ui-backend ## Build and Push UI Backend Docker image
	@echo "Publishing ui-backend docker image ..."
	docker push ${DOCKER_IMAGE}-ui-backend:${DOCKER_TAG}

.PHONY: test
test: ## Run Unit Tests
	@go test ./...

.PHONY: e2e
e2e: docker-apiserver docker-cli docker-orchestrator docker-ui docker-ui-backend ## Run go e2e test against code
	@cd e2e && \
	export APIServerContainerImage=${DOCKER_REGISTRY}/vmclarity-apiserver:${DOCKER_TAG} && \
	export OrchestratorContainerImage=${DOCKER_REGISTRY}/vmclarity-orchestrator:${DOCKER_TAG} && \
	export ScannerContainerImage=${DOCKER_REGISTRY}/vmclarity-cli:${DOCKER_TAG} && \
	export UIContainerImage=${DOCKER_REGISTRY}/vmclarity-ui:${DOCKER_TAG} && \
	export UIBackendContainerImage=${DOCKER_REGISTRY}/vmclarity-ui-backend:${DOCKER_TAG} && \
	go test -v -failfast -test.v -test.paniconexit0 -timeout 2h -ginkgo.v .

.PHONY: clean-ui
clean-ui:
	@(rm -rf ui/build ; echo "UI cleanup done" )

.PHONY: clean-golangci-lint
clean-golangci-lint:
	@(rm -rf bin/golangci-lint* ; echo "Golangci lint cleanup done" )

.PHONY: clean-licensei
clean-licensei:
	@(rm -rf bin/licensei* ; echo "Licensei cleanup done" )

.PHONY: clean-go
clean-go:
	@(rm -rf bin/vmclarity* ; echo "GO executables cleanup done" )

.PHONY: clean
clean: clean-ui clean-golangci-lint clean-licensei clean-go ## Clean all build artifacts

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

GOLANGCI_BIN := $(BIN_DIR)/golangci-lint
GOLANGCI_CONFIG := $(ROOT_DIR)/.golangci.yml
GOLANGCI_VERSION := 1.54.2

bin/golangci-lint: bin/golangci-lint-$(GOLANGCI_VERSION)
	@ln -sf golangci-lint-$(GOLANGCI_VERSION) bin/golangci-lint

bin/golangci-lint-$(GOLANGCI_VERSION): | $(BIN_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b "$(BIN_DIR)" "v$(GOLANGCI_VERSION)"
	@mv bin/golangci-lint $@

GOMODULES := $(shell find $(ROOT_DIR) -name 'go.mod')
LINTGOMODULES = $(addprefix lint-, $(GOMODULES))
FIXGOMODULES = $(addprefix fix-, $(GOMODULES))

.PHONY: $(LINTGOMODULES)
$(LINTGOMODULES):
	cd $(dir $(@:lint-%=%)) && "$(GOLANGCI_BIN)" run -c "$(GOLANGCI_CONFIG)"

.PHONY: lint-go
lint-go: bin/golangci-lint $(LINTGOMODULES)

.PHONY: lint-cfn
lint-cfn:
	# Requires cfn-lint to be installed
	# https://github.com/aws-cloudformation/cfn-lint#install
	cfn-lint installation/aws/VmClarity.cfn

.PHONY: lint
lint: lint-go lint-cfn ## Run linters

.PHONY: $(FIXGOMODULES)
$(FIXGOMODULES):
	cd $(dir $(@:fix-%=%)) && "$(GOLANGCI_BIN)" run -c "$(GOLANGCI_CONFIG)" --fix

.PHONY: fix
fix: bin/golangci-lint $(FIXGOMODULES) ## Fix lint violations

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}: | $(BIN_DIR)
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-check
license-check: bin/licensei ## Run license check
	./bin/licensei check

.PHONY: license-header
license-header: bin/licensei ## Run license header check
	./bin/licensei header

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	./bin/licensei cache

.PHONY: check
check: lint test helm-lint ## Run tests and linters

TIDYGOMODULES = $(addprefix tidy-, $(GOMODULES))

.PHONY: $(TIDYGOMODULES)
$(TIDYGOMODULES):
	cd $(dir $(@:tidy-%=%)) && go mod tidy -go=$(GO_VERSION)

.PHONY: gomod-tidy
gomod-tidy: $(TIDYGOMODULES)

.PHONY: gen-api
gen-api: gen-apiserver-api gen-uibackend-api ## Generating API code

.PHONY: gen-apiserver-api
gen-apiserver-api: ## Generating API for backend code
	@(echo "Generating API for backend code ..." )
	@(cd api; go generate)

.PHONY: gen-uibackend-api
gen-uibackend-api: ## Generating API for UI backend code
	@(echo "Generating API for UI backend code ..." )
	@(cd pkg/uibackend/api; go generate)

.PHONY: helm-docs
helm-docs:
	docker run --rm --volume "$(shell pwd):/helm-docs" -u $(shell id -u) jnorwood/helm-docs:v1.11.0

.PHONY: helm-lint
helm-lint:
	docker run --rm --workdir /workdir --volume "$(shell pwd):/workdir" quay.io/helmpack/chart-testing:v3.8.0 ct lint --all

ACTIONLINT_BIN := $(BIN_DIR)/actionlint
ACTIONLINT_VERSION := 1.6.26

bin/actionlint: bin/actionlint-$(ACTIONLINT_VERSION)
	@ln -sf actionlint-$(ACTIONLINT_VERSION) bin/actionlint

bin/actionlint-$(ACTIONLINT_VERSION): | $(BIN_DIR)
	curl -sSfL https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash \
	| bash -s -- "$(ACTIONLINT_VERSION)" "$(BIN_DIR)"
	@mv bin/actionlint $@

.PHONY: lint-actions
lint-actions: bin/actionlint
	@$(ACTIONLINT_BIN) -color
