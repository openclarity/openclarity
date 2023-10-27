####
## Make settings
####

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
.DEFAULT_GOAL := help

####
## Project variables
####

BINARY_NAME ?= vmclarity
VERSION ?= $(COMMIT_HASH)
DOCKER_REGISTRY ?= ghcr.io/openclarity
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= $(VERSION)
VMCLARITY_TOOLS_BASE ?=

####
## Runtime variables
####

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := $(ROOT_DIR)/bin
GOMODULES := $(shell find $(ROOT_DIR) -name 'go.mod')
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH := $(shell git rev-parse HEAD)
INSTALLATION_DIR := $(ROOT_DIR)/installation
HELM_CHART_DIR := $(INSTALLATION_DIR)/kubernetes/helm

include makefile.d/*.mk

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: ui build-all-go ## Build all components

.PHONY: build-all-go
build-all-go: bin/vmclarity-apiserver bin/vmclarity-cli bin/vmclarity-orchestrator bin/vmclarity-ui-backend ## Build all go components

bin/vmclarity-orchestrator: $(shell find api) $(shell find cmd/vmclarity-orchestrator) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -o bin/vmclarity-orchestrator cmd/vmclarity-orchestrator/main.go

bin/vmclarity-apiserver: $(shell find api) $(shell find cmd/vmclarity-apiserver) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -o bin/vmclarity-apiserver cmd/vmclarity-apiserver/main.go

bin/vmclarity-cli: $(shell find api) $(shell find cmd/vmclarity-cli) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -o bin/vmclarity-cli cmd/vmclarity-cli/main.go

bin/vmclarity-ui-backend: $(shell find api) $(shell find cmd/vmclarity-ui-backend) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -o bin/vmclarity-ui-backend cmd/vmclarity-ui-backend/main.go

.PHONY: clean
clean: clean-ui clean-go ## Clean all build artifacts

.PHONY: clean-go
clean-go: ## Clean all Go build artifacts
	@rm -rf bin/vmclarity*
	$(info GO executables cleanup done)

.PHONY: clean-ui
clean-ui: ## Clean UI build
	@rm -rf ui/build
	$(info UI cleanup done)

.PHONY: $(LINTGOMODULES)
TIDYGOMODULES = $(addprefix tidy-, $(GOMODULES))

$(TIDYGOMODULES):
	cd $(dir $(@:tidy-%=%)) && go mod tidy -go=$$(cat .go-version)

.PHONY: gomod-tidy
gomod-tidy: $(TIDYGOMODULES) ## Run go mod tidy for all go modules

.PHONY: ui
ui: ## Build UI component
	$(info Building UI ...)
	@(cd ui && npm i && npm run build)
	@ls -l ui/build

##@ Testing

.PHONY: check
check: lint test ## Run tests and linters

LINTGOMODULES = $(addprefix lint-, $(GOMODULES))
FIXGOMODULES = $(addprefix fix-, $(GOMODULES))

.PHONY: $(LINTGOMODULES)
$(LINTGOMODULES):
	cd $(dir $(@:lint-%=%)) && "$(GOLANGCI_BIN)" run -c "$(GOLANGCI_CONFIG)"

.PHONY: $(FIXGOMODULES)
$(FIXGOMODULES):
	cd $(dir $(@:fix-%=%)) && "$(GOLANGCI_BIN)" run -c "$(GOLANGCI_CONFIG)" --fix

.PHONY: fix
fix: bin/golangci-lint $(FIXGOMODULES) ## Fix linter errors in Go source code

.PHONY: e2e
e2e: docker-apiserver docker-cli docker-orchestrator docker-ui docker-ui-backend ## Run end-to-end test suite
	@cd e2e && \
	export APIServerContainerImage=$(DOCKER_REGISTRY)/vmclarity-apiserver:$(DOCKER_TAG) && \
	export OrchestratorContainerImage=$(DOCKER_REGISTRY)/vmclarity-orchestrator:$(DOCKER_TAG) && \
	export ScannerContainerImage=$(DOCKER_REGISTRY)/vmclarity-cli:$(DOCKER_TAG) && \
	export UIContainerImage=$(DOCKER_REGISTRY)/vmclarity-ui:$(DOCKER_TAG) && \
	export UIBackendContainerImage=$(DOCKER_REGISTRY)/vmclarity-ui-backend:$(DOCKER_TAG) && \
	go test -v -failfast -test.v -test.paniconexit0 -timeout 2h -ginkgo.v .

.PHONY: license-check
license-check: bin/licensei license-cache ## Check licenses for software components
	$(LICENSEI_BIN) check

.PHONY: license-header
license-header: bin/licensei ## Check license headers in source code files
	$(LICENSEI_BIN) header

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	$(LICENSEI_BIN) cache

.PHONY: lint
lint: license-check license-header lint-actions lint-bicep lint-cfn lint-go lint-helm ## Run all the linters

.PHONY: lint-actions
lint-actions: bin/actionlint ## Lint Github Actions
	@$(ACTIONLINT_BIN) -color

.PHONY: lint-bicep
lint-bicep: bin/bicep ## Lint Azure Bicep template(s)
	@$(BICEP_BIN) lint installation/azure/vmclarity.bicep

.PHONY: lint-cfn
lint-cfn: bin/cfn-lint ## Lint AWS CloudFormation template
	$(CFNLINT_BIN) installation/aws/VmClarity.cfn

.PHONY: lint-go
lint-go: bin/golangci-lint $(LINTGOMODULES) ## Lint Go source code

.PHONY: lint-helm
lint-helm: ## Lint Helm charts
	docker run --rm --workdir /workdir --volume "$(ROOT_DIR):/workdir" quay.io/helmpack/chart-testing:v3.8.0 ct lint --all

.PHONY: test
test: ## Run Go unit tests
	@go test ./...

##@ Docker

.PHONY: docker
docker: docker-apiserver docker-cli docker-orchestrator docker-ui docker-ui-backend ## Build All Docker images

.PHONY: docker-apiserver
docker-apiserver: ## Build API Server container image
	$(info Building apiserver docker image ...)
	docker build --file ./Dockerfile.apiserver --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(DOCKER_IMAGE)-apiserver:$(DOCKER_TAG) .

ifneq ($(strip $(VMCLARITY_TOOLS_BASE)),)
VMCLARITY_TOOLS_CLI_DOCKER_ARG=--build-arg VMCLARITY_TOOLS_BASE=${VMCLARITY_TOOLS_BASE}
endif

.PHONY: docker-cli
docker-cli: ## Build CLI container image
	$(info Building cli docker image ...)
	docker build --file ./Dockerfile.cli --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)  \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		${VMCLARITY_TOOLS_CLI_DOCKER_ARG} \
		-t $(DOCKER_IMAGE)-cli:$(DOCKER_TAG) .

.PHONY: docker-orchestrator
docker-orchestrator: ## Build Orchestrator container image
	$(info Building orchestrator docker image ...)
	docker build --file ./Dockerfile.orchestrator --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)  \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(DOCKER_IMAGE)-orchestrator:$(DOCKER_TAG) .

.PHONY: docker-ui
docker-ui: ## Build UI container image
	$(info Building ui docker image ...)
	docker build --file ./Dockerfile.ui \
		-t $(DOCKER_IMAGE)-ui:$(DOCKER_TAG) .

.PHONY: docker-ui-backend
docker-ui-backend: ## Build UI Backend container image
	$(info Building ui-backend docker image ...)
	docker build --file ./Dockerfile.uibackend --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)  \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(DOCKER_IMAGE)-ui-backend:$(DOCKER_TAG) .

.PHONY: push-docker
push-docker: push-docker-apiserver push-docker-cli push-docker-orchestrator push-docker-ui push-docker-ui-backend ## Build and push all container images

.PHONY: push-docker-apiserver
push-docker-apiserver: docker-apiserver ## Build and push API Server container image
	$(info Publishing apiserver docker image ...)
	docker push $(DOCKER_IMAGE)-apiserver:$(DOCKER_TAG)

.PHONY: push-docker-cli
push-docker-cli: docker-cli ## Build and push CLI Docker image
	$(info Publishing cli docker image ...)
	docker push $(DOCKER_IMAGE)-cli:$(DOCKER_TAG)

.PHONY: push-docker-orchestrator
push-docker-orchestrator: docker-orchestrator ## Build and push Orchestrator container image
	$(info Publishing orchestrator docker image ...)
	docker push $(DOCKER_IMAGE)-orchestrator:$(DOCKER_TAG)

.PHONY: push-docker-ui
push-docker-ui: docker-ui ## Build and Push UI container image
	$(info Publishing ui docker image ...)
	docker push $(DOCKER_IMAGE)-ui:$(DOCKER_TAG)

.PHONY: push-docker-ui-backend
push-docker-ui-backend: docker-ui-backend ## Build and push UI Backend container image
	$(info Publishing ui-backend docker image ...)
	docker push $(DOCKER_IMAGE)-ui-backend:$(DOCKER_TAG)

##@ Code generation

.PHONY: gen
gen: gen-api gen-bicep gen-helm-docs ## Generating all code, manifests, docs

.PHONY: gen-api
gen-api: gen-apiserver-api gen-uibackend-api ## Generating API code

.PHONY: gen-apiserver-api
gen-apiserver-api: ## Generating Go library for API specification
	$(info Generating API for backend code ...)
	@(cd api && go generate)

.PHONY: gen-uibackend-api
gen-uibackend-api: ## Generating Go library for UI Backend API specification
	$(info Generating API for UI backend code ...)
	@(cd pkg/uibackend/api && go generate)

.PHONY: gen-bicep
gen-bicep: bin/bicep ## Generating Azure Bicep template(s)
	$(info Generating Azure Bicep template(s) ...)
	@$(BICEP_BIN) build installation/azure/vmclarity.bicep

.PHONY: gen-helm-docs
gen-helm-docs: ## Generating documentation for Helm chart
	$(info Generating Helm chart(s) documentation ...)
	docker run --rm --volume "$(HELM_CHART_DIR):/helm-docs" -u $(shell id -u) jnorwood/helm-docs:v1.11.0

