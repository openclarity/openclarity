SHELL=/bin/bash

# Project variables
BINARY_NAME ?= vmclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_REGISTRY ?= ghcr.io/openclarity
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= ${VERSION}

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
build: ## Build VMClarity
	@(echo "Building VMClarity" )
	@(go build -ldflags="-s -w \
           -X 'github.com/openclarity/vmclarity/pkg/version.Version=${VERSION}'" -o bin/${BINARY_NAME} ./main.go \
           && ls -lh bin/${BINARY_NAME})

.PHONY: docker
docker:  ## Build VMClarity docker image
	@(echo "Building VMClarity docker image ..." )
	docker build --file ./Dockerfile --build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIMESTAMP=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg COMMIT_HASH=$(shell git rev-parse HEAD) \
		-t ${DOCKER_IMAGE}:${DOCKER_TAG} .

.PHONY: push-docker
push-docker: docker ## Build and Push VMClarity docker image
	@echo "Publishing VMClarity Docker image ..."
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}

.PHONY: test
test: ## Run Unit Tests
	@(CGO_ENABLED=0 go test ./...)
	@(cd runtime_scan && go test ./...)

.PHONY: clean
clean: ## Clean all build artifacts

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b ./bin/ v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	./bin/golangci-lint run
	cd runtime_scan && ../bin/golangci-lint run

.PHONY: fix
fix: bin/golangci-lint ## Fix lint violations
	./bin/golangci-lint run --fix
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

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	./bin/licensei cache

.PHONY: check
check: lint test ## Run tests and linters

.PHONY: gomod-tidy
gomod-tidy:
	cd runtime_scan && go mod tidy