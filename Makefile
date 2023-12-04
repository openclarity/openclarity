####
## Make settings
####

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec
.DEFAULT_GOAL := help

####
## Project variables
####

VERSION ?= $(shell git rev-parse --short HEAD)
DOCKER_REGISTRY ?= ghcr.io/openclarity
DOCKER_TAG ?= $(VERSION)
VMCLARITY_TOOLS_BASE ?=
GO_VERSION ?= 1.21.4

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
HELM_OCI_REPOSITORY := ghcr.io/openclarity/charts
DIST_DIR ?= $(ROOT_DIR)/dist

VMCLARITY_APISERVER_IMAGE = $(DOCKER_REGISTRY)/vmclarity-apiserver:$(DOCKER_TAG)
VMCLARITY_ORCHESTRATOR_IMAGE = $(DOCKER_REGISTRY)/vmclarity-orchestrator:$(DOCKER_TAG)
VMCLARITY_UI_IMAGE = $(DOCKER_REGISTRY)/vmclarity-ui:$(DOCKER_TAG)
VMCLARITY_UIBACKEND_IMAGE = $(DOCKER_REGISTRY)/vmclarity-ui-backend:$(DOCKER_TAG)
VMCLARITY_SCANNER_IMAGE = $(DOCKER_REGISTRY)/vmclarity-cli:$(DOCKER_TAG)
VMCLARITY_CR_DISCOVERY_SERVER_IMAGE = $(DOCKER_REGISTRY)/vmclarity-cr-discovery-server:$(DOCKER_TAG)

####
## Load additional makefiles
####

include makefile.d/*.mk

$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: ui build-all-go ## Build all components

.PHONY: build-all-go
build-all-go: bin/vmclarity-apiserver bin/vmclarity-cli bin/vmclarity-orchestrator bin/vmclarity-ui-backend bin/vmclarity-cr-discovery-server ## Build all go components

bin/vmclarity-orchestrator: $(shell find api) $(shell find cmd/vmclarity-orchestrator) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -ldflags="-s -w \
		-X 'github.com/openclarity/vmclarity/pkg/version.Version=$(VERSION)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.CommitHash=$(COMMIT_HASH)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.BuildTimestamp=$(BUILD_TIMESTAMP)'" \
		-o $@ cmd/vmclarity-orchestrator/main.go

bin/vmclarity-apiserver: $(shell find api) $(shell find cmd/vmclarity-apiserver) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -ldflags="-s -w \
		-X 'github.com/openclarity/vmclarity/pkg/version.Version=$(VERSION)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.CommitHash=$(COMMIT_HASH)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.BuildTimestamp=$(BUILD_TIMESTAMP)'" \
		-o $@ cmd/vmclarity-apiserver/main.go

bin/vmclarity-cli: $(shell find api) $(shell find cmd/vmclarity-cli) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -ldflags="-s -w  \
		-X 'github.com/openclarity/vmclarity/pkg/version.Version=$(VERSION)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.CommitHash=$(COMMIT_HASH)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.BuildTimestamp=$(BUILD_TIMESTAMP)'" \
		-o $@ cmd/vmclarity-cli/main.go

bin/vmclarity-ui-backend: $(shell find api) $(shell find cmd/vmclarity-ui-backend) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -ldflags="-s -w \
		-X 'github.com/openclarity/vmclarity/pkg/version.Version=$(VERSION)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.CommitHash=$(COMMIT_HASH)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.BuildTimestamp=$(BUILD_TIMESTAMP)'" \
		-o $@ cmd/vmclarity-ui-backend/main.go

bin/vmclarity-cr-discovery-server: $(shell find api) $(shell find cmd/vmclarity-cr-discovery-server) $(shell find pkg) go.mod go.sum | $(BIN_DIR)
	go build -race -o bin/vmclarity-cr-discovery-server cmd/vmclarity-cr-discovery-server/main.go

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
	cd $(dir $(@:tidy-%=%)) && go mod tidy -go=$(GO_VERSION)

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
	export VMCLARITY_E2E_APISERVER_IMAGE=$(VMCLARITY_APISERVER_IMAGE) \
           VMCLARITY_E2E_ORCHESTRATOR_IMAGE=$(VMCLARITY_ORCHESTRATOR_IMAGE) \
           VMCLARITY_E2E_UI_IMAGE=$(VMCLARITY_UI_IMAGE) \
           VMCLARITY_E2E_UIBACKEND_IMAGE=$(VMCLARITY_UIBACKEND_IMAGE) \
           VMCLARITY_E2E_SCANNER_IMAGE=$(VMCLARITY_SCANNER_IMAGE) && \
	cd e2e && \
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
docker: docker-apiserver docker-cli docker-orchestrator docker-ui docker-ui-backend docker-cr-discovery-server ## Build All Docker images

.PHONY: docker-apiserver
docker-apiserver: ## Build API Server container image
	$(info Building apiserver docker image ...)
	docker build --file ./Dockerfile.apiserver --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP) \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(VMCLARITY_APISERVER_IMAGE) .

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
		-t $(VMCLARITY_SCANNER_IMAGE) .

.PHONY: docker-orchestrator
docker-orchestrator: ## Build Orchestrator container image
	$(info Building orchestrator docker image ...)
	docker build --file ./Dockerfile.orchestrator --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)  \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(VMCLARITY_ORCHESTRATOR_IMAGE) .

.PHONY: docker-ui
docker-ui: ## Build UI container image
	$(info Building ui docker image ...)
	docker build --file ./Dockerfile.ui \
		-t $(VMCLARITY_UI_IMAGE) .

.PHONY: docker-ui-backend
docker-ui-backend: ## Build UI Backend container image
	$(info Building ui-backend docker image ...)
	docker build --file ./Dockerfile.uibackend --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)  \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(VMCLARITY_UIBACKEND_IMAGE) .

.PHONY: docker-cr-discovery-server
docker-cr-discovery-server: ## Build K8S Image Resolver Docker image
	$(info Building cr-discovery-server docker image ...)
	docker build --file ./Dockerfile.cr-discovery-server --build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)  \
		--build-arg COMMIT_HASH=$(COMMIT_HASH) \
		-t $(VMCLARITY_CR_DISCOVERY_SERVER_IMAGE) .

.PHONY: push-docker
push-docker: push-docker-apiserver push-docker-cli push-docker-orchestrator push-docker-ui push-docker-ui-backend push-docker-cr-discovery-server ## Build and Push All Docker images

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

.PHONY: push-docker-cr-discovery-server
push-docker-cr-discovery-server: docker-cr-discovery-server ## Build and Push K8S Image Resolver Docker image
	@echo "Publishing cr-discovery-server docker image ..."
	docker push $(DOCKER_IMAGE)-cr-discovery-server:$(DOCKER_TAG)

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
gen-helm-docs: bin/helm-docs ## Generating documentation for Helm chart
	$(info Generating Helm chart(s) documentation ...)
	$(HELMDOCS_BIN) --chart-search-root $(HELM_CHART_DIR)

##@ Release

.DELETE_ON_ERROR:

.PHONY: clean-dist
clean-dist:
	rm -rf $(DIST_DIR)/*

.PHONY: dist-all
dist-all: dist-bicep dist-cloudformation dist-docker-compose dist-gcp-deployment dist-helm-chart dist-vmclarity-cli

CLI_OSARCH := $(shell echo {linux-,darwin-}{amd64,arm64})
CLI_BINARIES := $(CLI_OSARCH:%=$(DIST_DIR)/%/vmclarity-cli)
CLI_TARS := $(CLI_OSARCH:%=$(DIST_DIR)/vmclarity-cli-$(VERSION)-%.tar.gz)
CLI_TAR_SHA256SUMS := $(CLI_TARS:%=%.sha256sum)

.PHONY: dist-vmclarity-cli
dist-vmclarity-cli: $(CLI_BINARIES) $(CLI_TARS) $(CLI_TAR_SHA256SUMS) | $(DIST_DIR) ## Create vmclarity-cli release artifacts

$(DIST_DIR)/vmclarity-cli-$(VERSION)-%.tar.gz: $(DIST_DIR)/%/vmclarity-cli $(DIST_DIR)/%/LICENSE $(DIST_DIR)/%/README.md
	$(info --- Bundling $(dir $<) into $(notdir $@))
	tar cv -f $@ -C $(dir $<) --use-compress-program='gzip -9' $(notdir $^)

$(DIST_DIR)/%/vmclarity-cli: $(shell find api) $(shell find cmd/vmclarity-cli) $(shell find pkg) go.mod go.sum
	$(info --- Building $(notdir $@) for $*)
	GOOS=$(firstword $(subst -, ,$*)) \
	GOARCH=$(lastword $(subst -, ,$*)) \
	CGO_ENABLED=0 \
	go build -ldflags="-s -w \
		-X 'github.com/openclarity/vmclarity/pkg/version.Version=$(VERSION)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.CommitHash=$(COMMIT_HASH)' \
		-X 'github.com/openclarity/vmclarity/pkg/version.BuildTimestamp=$(BUILD_TIMESTAMP)'" \
		-o $@ cmd/$(notdir $@)/main.go

$(DIST_DIR)/%/LICENSE: $(ROOT_DIR)/LICENSE
	cp -v $< $@

$(DIST_DIR)/%/README.md: $(ROOT_DIR)/README.md
	cp -v $< $@

CFN_DIR := $(INSTALLATION_DIR)/aws
CFN_FILES := $(shell find $(CFN_DIR))
CFN_DIST_DIR := $(DIST_DIR)/cloudformation

.PHONY: dist-cloudformation
dist-cloudformation: $(DIST_DIR)/aws-cloudformation-$(VERSION).tar.gz $(DIST_DIR)/aws-cloudformation-$(VERSION).tar.gz.sha256sum ## Create AWS CloudFormation release artifacts

$(DIST_DIR)/aws-cloudformation-$(VERSION).tar.gz: $(DIST_DIR)/aws-cloudformation-$(VERSION).bundle $(CFN_DIST_DIR)/LICENSE | $(CFN_DIST_DIR)
	$(info --- Bundle $(CFN_DIST_DIR) into $(notdir $@))
	tar cv -f $@ -C $(CFN_DIST_DIR) --use-compress-program='gzip -9' $(shell ls $(CFN_DIST_DIR))

$(DIST_DIR)/aws-cloudformation-$(VERSION).bundle: $(CFN_FILES) | $(CFN_DIST_DIR)
	$(info --- Generate Cloudformation bundle)
	cp -vR $(CFN_DIR)/* $(CFN_DIST_DIR)/
	sed -i -E 's@(ghcr\.io\/openclarity\/vmclarity\-(apiserver|cli|orchestrator|ui-backend|ui)):latest@\1:$(VERSION)@' $(CFN_DIST_DIR)/VmClarity.cfn
	@touch $@

$(CFN_DIST_DIR)/LICENSE: $(ROOT_DIR)/LICENSE | $(CFN_DIST_DIR)
	$(info --- Copy $(notdir $@) to $@)
	cp -v $< $@

$(CFN_DIST_DIR):
	@mkdir -p $@

BICEP_DIR := $(INSTALLATION_DIR)/azure
BICEP_FILES := $(shell find $(BICEP_DIR))
BICEP_DIST_DIR := $(DIST_DIR)/bicep

.PHONY: dist-bicep
dist-bicep: $(DIST_DIR)/azure-bicep-$(VERSION).tar.gz $(DIST_DIR)/azure-bicep-$(VERSION).tar.gz.sha256sum ## Create Azure Bicep release artifacts

$(DIST_DIR)/azure-bicep-$(VERSION).tar.gz: $(DIST_DIR)/azure-bicep-$(VERSION).bundle $(BICEP_DIST_DIR)/LICENSE | $(BICEP_DIST_DIR)
	$(info --- Bundle $(BICEP_DIST_DIR) into $(notdir $@))
	tar cv -f $@ -C $(BICEP_DIST_DIR) --use-compress-program='gzip -9' $(shell ls $(BICEP_DIST_DIR))

$(DIST_DIR)/azure-bicep-$(VERSION).bundle: $(BICEP_FILES) bin/bicep | $(BICEP_DIST_DIR)
	$(info --- Generate Bicep bundle)
	cp -vR $(BICEP_DIR)/* $(BICEP_DIST_DIR)/
	sed -i -E 's@(ghcr\.io\/openclarity\/vmclarity\-(apiserver|cli|orchestrator|ui-backend|ui)):latest@\1:$(VERSION)@' \
		$(BICEP_DIST_DIR)/*.bicep $(BICEP_DIST_DIR)/vmclarity-UI.json
	$(BICEP_BIN) build $(BICEP_DIST_DIR)/vmclarity.bicep
	@touch $@

$(BICEP_DIST_DIR)/LICENSE: $(ROOT_DIR)/LICENSE | $(BICEP_DIST_DIR)
	cp -v $< $@

$(BICEP_DIST_DIR):
	@mkdir -p $@

DOCKER_COMPOSE_DIR := $(INSTALLATION_DIR)/docker
DOCKER_COMPOSE_FILES := $(shell find $(DOCKER_COMPOSE_DIR))
DOCKER_COMPOSE_DIST_DIR := $(DIST_DIR)/docker-compose

.PHONY: dist-docker-compose
dist-docker-compose: $(DIST_DIR)/docker-compose-$(VERSION).tar.gz $(DIST_DIR)/docker-compose-$(VERSION).tar.gz.sha256sum ## Create Docker Compose release artifacts

$(DIST_DIR)/docker-compose-$(VERSION).tar.gz: $(DIST_DIR)/docker-compose-$(VERSION).bundle $(DOCKER_COMPOSE_DIST_DIR)/LICENSE | $(DOCKER_COMPOSE_DIST_DIR)
	$(info --- Bundle $(DOCKER_COMPOSE_DIST_DIR) into $(notdir $@))
	tar cv -f $@ -C $(DOCKER_COMPOSE_DIST_DIR) --use-compress-program='gzip -9' $(shell ls $(DOCKER_COMPOSE_DIST_DIR))

$(DIST_DIR)/docker-compose-$(VERSION).bundle: $(DOCKER_COMPOSE_FILES) | $(DOCKER_COMPOSE_DIST_DIR)
	$(info --- Generate Docker Compose bundle)
	cp -vR $(DOCKER_COMPOSE_DIR)/* $(DOCKER_COMPOSE_DIST_DIR)/
	sed -i -E 's@(ghcr\.io\/openclarity\/vmclarity\-(apiserver|cli|orchestrator|ui-backend|ui)):latest@\1:$(VERSION)@' \
		$(DOCKER_COMPOSE_DIST_DIR)/*.yml $(DOCKER_COMPOSE_DIST_DIR)/*.yaml $(DOCKER_COMPOSE_DIST_DIR)/*.env
	@touch $@

$(DOCKER_COMPOSE_DIST_DIR)/LICENSE: $(ROOT_DIR)/LICENSE | $(DOCKER_COMPOSE_DIST_DIR)
	$(info --- Copy $(notdir $@) to $@)
	cp -v $< $@

$(DOCKER_COMPOSE_DIST_DIR):
	@mkdir -p $@

GCP_DM_DIR := $(INSTALLATION_DIR)/gcp/dm
GCP_DM_FILES := $(shell find $(GCP_DM_DIR))
GCP_DM_DIST_DIR := $(DIST_DIR)/gcp-deployment

.PHONY: dist-gcp-deployment
dist-gcp-deployment: $(DIST_DIR)/gcp-deployment-$(VERSION).tar.gz $(DIST_DIR)/gcp-deployment-$(VERSION).tar.gz.sha256sum ## Create Google Cloud Deployment bundle

$(DIST_DIR)/gcp-deployment-$(VERSION).tar.gz: $(DIST_DIR)/gcp-deployment-$(VERSION).bundle $(GCP_DM_DIST_DIR)/LICENSE | $(GCP_DM_DIST_DIR)
	$(info --- Bundle $(GCP_DM_DIST_DIR) into $(notdir $@))
	tar cv -f $@ -C $(GCP_DM_DIST_DIR) --use-compress-program='gzip -9' $(shell ls $(GCP_DM_DIST_DIR))

$(DIST_DIR)/gcp-deployment-$(VERSION).bundle: $(GCP_DM_FILES) | $(GCP_DM_DIST_DIR)
	$(info --- Generate Google Cloud Deployment bundle)
	cp -vR $(GCP_DM_DIR)/* $(GCP_DM_DIST_DIR)/
	sed -i -E 's@(ghcr\.io\/openclarity\/vmclarity\-(apiserver|cli|orchestrator|ui-backend|ui)):latest@\1:$(VERSION)@' \
		$(GCP_DM_DIST_DIR)/vmclarity.py.schema $(GCP_DM_DIST_DIR)/components/vmclarity-server.py.schema
	@touch $@

$(GCP_DM_DIST_DIR)/LICENSE: $(ROOT_DIR)/LICENSE | $(GCP_DM_DIST_DIR)
	cp -v $< $@

$(GCP_DM_DIST_DIR):
	@mkdir -p $@

HELM_CHART_DIR := $(INSTALLATION_DIR)/kubernetes/helm/vmclarity
HELM_CHART_FILES := $(shell find $(HELM_CHART_DIR))
HELM_CHART_DIST_DIR := $(DIST_DIR)/helm-vmclarity-chart

.PHONY: dist-helm-chart
dist-helm-chart: $(DIST_DIR)/vmclarity-$(VERSION:v%=%).tgz $(DIST_DIR)/vmclarity-$(VERSION:v%=%).tgz.sha256sum ## Create Helm Chart bundle

$(DIST_DIR)/vmclarity-$(VERSION:v%=%).tgz: $(DIST_DIR)/helm-vmclarity-chart-$(VERSION:v%=%).bundle bin/helm | $(HELM_CHART_DIST_DIR)
	$(info --- Bundle $(HELM_CHART_DIST_DIR) into $(notdir $@))
	$(HELM_BIN) package $(HELM_CHART_DIST_DIR) --version "$(VERSION:v%=%)" --app-version "$(VERSION)" --destination $(DIST_DIR)

$(DIST_DIR)/helm-vmclarity-chart-$(VERSION:v%=%).bundle: $(HELM_CHART_FILES) bin/yq bin/helm-docs | $(HELM_CHART_DIST_DIR)
	$(info --- Generate Helm Chart bundle)
	cp -vR $(HELM_CHART_DIR)/* $(HELM_CHART_DIST_DIR)/
	$(YQ_BIN) -i '.apiserver.image.tag = "$(VERSION)" | .orchestrator.image.tag = "$(VERSION)" | .orchestrator.scannerImage.tag = "$(VERSION)" | .ui.image.tag = "$(VERSION)" | .uibackend.image.tag = "$(VERSION)"' \
	$(HELM_CHART_DIST_DIR)/values.yaml
	$(YQ_BIN) -i '.version = "$(VERSION:v%=%)" | .appVersion = "$(VERSION)"' $(HELM_CHART_DIST_DIR)/Chart.yaml
	$(HELMDOCS_BIN) --chart-search-root $(HELM_CHART_DIST_DIR)
	@touch $@

$(HELM_CHART_DIST_DIR):
	@mkdir -p $@

.PHONY: publish-helm-chart
publish-helm-chart: $(DIST_DIR)/vmclarity-$(VERSION:v%=%).tgz bin/helm ## Publish Helm Chart bundle to OCI registry
	$(HELM_BIN) push $< oci://$(HELM_OCI_REPOSITORY)

$(DIST_DIR)/%.sha256sum: | $(DIST_DIR)
	$(info --- Generate SHA256 for $(notdir $@))
	shasum -a 256 $(basename $@) | sed "s@$(dir $@)@@" > $@

.PHONY: generate-release-notes
generate-release-notes: $(DIST_DIR)/CHANGELOG.md ## Generate Release Notes

GITCLIFF_OPTS := --strip all
ifeq ($(CI),true)
	GITCLIFF_OPTS += -vv --latest --tag $(VERSION)
else
	GITCLIFF_OPTS += --unreleased --bump
endif

$(DIST_DIR)/CHANGELOG.md: $(ROOT_DIR)/cliff.toml bin/git-cliff | $(DIST_DIR)
	$(GITCLIFF_BIN) --config $(ROOT_DIR)/cliff.toml --output $@ $(GITCLIFF_OPTS)
