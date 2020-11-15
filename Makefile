SOURCE := ./cmd/kubei/
BINARY := kubei
DOCKER_REGISTRY ?= gcr.io/development-infra-208909
VERSION ?= $(shell git rev-parse HEAD)
IMAGE_NAME ?= $(DOCKER_REGISTRY)/$(BINARY):$(VERSION)
TARGET_OS ?= linux

build:
	GOOS=$(TARGET_OS) CGO_ENABLED=0 go build -mod vendor -o $(BINARY) $(SOURCE)

# builds the current dev docker version
docker:
	docker build --build-arg BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ") \
		--build-arg VCS_REF=$(shell git rev-parse --short HEAD) \
			-t $(IMAGE_NAME) .

docker.push: docker
	docker push $(IMAGE_NAME)

test:
	GO111MODULE=on go test -v `go list ./pkg/...`
