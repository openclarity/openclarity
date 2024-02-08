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
GO_VERSION ?= $(shell cat $(ROOT_DIR)/.go-version)

####
## Runtime variables
####

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := $(ROOT_DIR)/bin
GOMODULES := $(shell find $(ROOT_DIR) -name 'go.mod' -exec dirname {} \;)
BUILD_TIMESTAMP := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH := $(shell git rev-parse HEAD)
INSTALLATION_DIR := $(ROOT_DIR)/installation
HELM_CHART_DIR := $(INSTALLATION_DIR)/kubernetes/helm
HELM_OCI_REPOSITORY := ghcr.io/openclarity/charts
DIST_DIR ?= $(ROOT_DIR)/dist

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

LDFLAGS = -s -w
LDFLAGS += -X 'github.com/openclarity/vmclarity/core/version.Version=$(VERSION)'
LDFLAGS += -X 'github.com/openclarity/vmclarity/core/version.CommitHash=$(COMMIT_HASH)'
LDFLAGS += -X 'github.com/openclarity/vmclarity/core/version.BuildTimestamp=$(BUILD_TIMESTAMP)'

bin/vmclarity-orchestrator: $(shell find api provider orchestrator utils core) | $(BIN_DIR)
	cd orchestrator && go build -race -ldflags="$(LDFLAGS)" -o $(ROOT_DIR)/$@ cmd/main.go

bin/vmclarity-apiserver: $(shell find api api/server) | $(BIN_DIR)
	cd api/server && go build -race -ldflags="$(LDFLAGS)" -o $(ROOT_DIR)/$@ cmd/main.go

bin/vmclarity-cli: $(shell find api cli utils core) | $(BIN_DIR)
	cd cli && go build -race -ldflags="$(LDFLAGS)" -o $(ROOT_DIR)/$@ cmd/main.go

bin/vmclarity-ui-backend: $(shell find api uibackend/server)  | $(BIN_DIR)
	cd uibackend/server && go build -race -ldflags="$(LDFLAGS)" -o $(ROOT_DIR)/$@ cmd/main.go

bin/vmclarity-cr-discovery-server: $(shell find api containerruntimediscovery/server utils core) | $(BIN_DIR)
	cd containerruntimediscovery/server && go build -race -ldflags="$(LDFLAGS)" -o $(ROOT_DIR)/$@ cmd/main.go

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
	cd $(@:tidy-%=%) && go mod tidy -go=$(GO_VERSION)

.PHONY: gomod-tidy
gomod-tidy: $(TIDYGOMODULES) ## Run go mod tidy for all go modules

.PHONY: $(MODLISTGOMODULES)
MODLISTGOMODULES = $(addprefix modlist-, $(GOMODULES))

$(MODLISTGOMODULES):
	cd $(@:modlist-%=%) && go list -m -mod=readonly all 1> /dev/null

.PHONY: gomod-list
gomod-list: $(MODLISTGOMODULES)

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
	cd $(@:lint-%=%) && "$(GOLANGCI_BIN)" run -c "$(GOLANGCI_CONFIG)"

.PHONY: $(FIXGOMODULES)
$(FIXGOMODULES):
	cd $(@:fix-%=%) && "$(GOLANGCI_BIN)" run -c "$(GOLANGCI_CONFIG)" --fix

.PHONY: fix
fix: bin/golangci-lint $(FIXGOMODULES) ## Fix linter errors in Go source code

E2E_TARGETS =
E2E_ENV =
ifneq ($(CI),true)
	E2E_TARGETS += docker
	E2E_ENV += VMCLARITY_E2E_APISERVER_IMAGE=$(DOCKER_REGISTRY)/vmclarity-apiserver:$(DOCKER_TAG)
	E2E_ENV += VMCLARITY_E2E_ORCHESTRATOR_IMAGE=$(DOCKER_REGISTRY)/vmclarity-orchestrator:$(DOCKER_TAG)
	E2E_ENV += VMCLARITY_E2E_UI_IMAGE=$(DOCKER_REGISTRY)/vmclarity-ui:$(DOCKER_TAG)
	E2E_ENV += VMCLARITY_E2E_UIBACKEND_IMAGE=$(DOCKER_REGISTRY)/vmclarity-ui-backend:$(DOCKER_TAG)
	E2E_ENV += VMCLARITY_E2E_SCANNER_IMAGE=$(DOCKER_REGISTRY)/vmclarity-cli:$(DOCKER_TAG)
	E2E_ENV += VMCLARITY_E2E_CR_DISCOVERY_SERVER_IMAGE=$(DOCKER_REGISTRY)/vmclarity-cr-discovery-server:$(DOCKER_TAG)
endif

.PHONY: e2e
e2e: $(E2E_TARGETS) ## Run end-to-end test suite
	cd e2e && $(E2E_ENV) go test -v -failfast -test.v -test.paniconexit0 -timeout 2h -ginkgo.v .

VENDORMODULES = $(addprefix vendor-, $(GOMODULES))

$(VENDORMODULES):
	cd $(@:vendor-%=%) && go mod vendor

.PHONY: gomod-vendor
gomod-vendor: $(VENDORMODULES) # Make vendored copy of dependencies for all modules

LICENSECHECKMODULES = $(addprefix license-check-, $(GOMODULES))

$(LICENSECHECKMODULES):
	cd $(@:license-check-%=%) && "$(LICENSEI_BIN)" check --config "$(LICENSEI_CONFIG)"

.PHONY: license-check
license-check: bin/licensei $(LICENSECHECKMODULES) ## Check licenses for software components

LICENSECACHEMODULES = $(addprefix license-cache-, $(GOMODULES))

$(LICENSECACHEMODULES):
	cd $(@:license-cache-%=%) && "$(LICENSEI_BIN)" cache --config "$(LICENSEI_CONFIG)"

+.PHONY: license-cache
license-cache: bin/licensei $(LICENSECACHEMODULES) ## Generate license cache

.PHONY: lint
lint: license-check lint-actions lint-bicep lint-cfn lint-go lint-helm ## Run all the linters

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

GOTEST_OPTS := -failfast -timeout 30m -short
ifeq ($(CI),true)
	GOTEST_OPTS += -v
endif

TESTGOMODULES = $(addprefix test-, $(GOMODULES))

$(TESTGOMODULES):
	cd $(@:test-%=%) && go test $(GOTEST_OPTS) ./...

.PHONY: test
test: $(TESTGOMODULES) ## Run Go unit tests

##@ Docker

# Export params required in Docker Bake
BAKE_ENV = DOCKER_REGISTRY=$(DOCKER_REGISTRY)
BAKE_ENV += DOCKER_TAG=$(DOCKER_TAG)
BAKE_ENV += VERSION=$(VERSION)
BAKE_ENV += BUILD_TIMESTAMP=$(BUILD_TIMESTAMP)
BAKE_ENV += COMMIT_HASH=$(COMMIT_HASH)

BAKE_OPTS =
ifneq ($(strip $(VMCLARITY_TOOLS_BASE)),)
	BAKE_OPTS += --set vmclarity-cli.args.VMCLARITY_TOOLS_BASE=$(VMCLARITY_TOOLS_BASE)
endif

.PHONY: docker
docker: ## Build All Docker images
	$(info Building all docker images ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS)

.PHONY: docker-apiserver
docker-apiserver: ## Build API Server container image
	$(info Building apiserver docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-apiserver

.PHONY: docker-cli
docker-cli: ## Build CLI container image
	$(info Building cli docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-cli

.PHONY: docker-orchestrator
docker-orchestrator: ## Build Orchestrator container image
	$(info Building orchestrator docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-orchestrator

.PHONY: docker-ui
docker-ui: ## Build UI container image
	$(info Building ui docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-ui

.PHONY: docker-ui-backend
docker-ui-backend: ## Build UI Backend container image
	$(info Building ui-backend docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-ui-backend

.PHONY: docker-cr-discovery-server
docker-cr-discovery-server: ## Build K8S Image Resolver Docker image
	$(info Building cr-discovery-server docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-cr-discovery-server

.PHONY: push-docker
push-docker: BAKE_OPTS += --set *.output=type=registry
push-docker: ## Build and Push All Docker images
	$(info Publishing all docker images ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS)

.PHONY: push-docker-apiserver
push-docker-apiserver: BAKE_OPTS += --set *.output=type=registry
push-docker-apiserver: ## Build and push API Server container image
	$(info Publishing apiserver docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-apiserver

.PHONY: push-docker-cli
push-docker-cli: BAKE_OPTS += --set *.output=type=registry
push-docker-cli: ## Build and push CLI Docker image
	$(info Publishing cli docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-cli

.PHONY: push-docker-orchestrator
push-docker-orchestrator: BAKE_OPTS += --set *.output=type=registry
push-docker-orchestrator: ## Build and push Orchestrator container image
	$(info Publishing orchestrator docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-orchestrator

.PHONY: push-docker-ui
push-docker-ui: BAKE_OPTS += --set *.output=type=registry
push-docker-ui: ## Build and Push UI container image
	$(info Publishing ui docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-ui

.PHONY: push-docker-ui-backend
push-docker-ui-backend: BAKE_OPTS += --set *.output=type=registry
push-docker-ui-backend: ## Build and push UI Backend container image
	$(info Publishing ui-backend docker image ...)
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-ui-backend

.PHONY: push-docker-cr-discovery-server
push-docker-cr-discovery-server: BAKE_OPTS += --set *.output=type=registry
push-docker-cr-discovery-server: ## Build and Push K8S Image Resolver Docker image
	@echo "Publishing cr-discovery-server docker image ..."
	$(BAKE_ENV) docker buildx bake $(BAKE_OPTS) vmclarity-cr-discovery-server

##@ Code generation

.PHONY: gen
gen: gen-api gen-bicep gen-helm-docs ## Generating all code, manifests, docs

.PHONY: gen-api
gen-api: gen-apiserver-api gen-uibackend-api ## Generating API code

.PHONY: gen-apiserver-api
gen-apiserver-api: ## Generating Go library for API specification
	$(info Generating API for backend code ...)
	@(cd api/types && go generate)
	@(cd api/client && go generate)
	@(cd api/server && go generate)

.PHONY: gen-uibackend-api
gen-uibackend-api: ## Generating Go library for UI Backend API specification
	$(info Generating API for UI backend code ...)
	@(cd uibackend/types && go generate)
	@(cd uibackend/client && go generate)
	@(cd uibackend/server && go generate)

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

$(DIST_DIR)/%/vmclarity-cli: $(shell find api cli utils core)
	$(info --- Building $(notdir $@) for $*)
	GOOS=$(firstword $(subst -, ,$*)) \
	GOARCH=$(lastword $(subst -, ,$*)) \
	CGO_ENABLED=0 \
	go build -ldflags="$(LDFLAGS)" -o $@ cmd/$(notdir $@)/main.go

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
