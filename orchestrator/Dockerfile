# syntax=docker/dockerfile:1@sha256:fe40cf4e92cd0c467be2cfc30657a680ae2398318afd50b0c80585784c604f28

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.4.0@sha256:0cd3f05c72d6c9b038eb135f91376ee1169ef3a330d34e418e65e2a5c2e9c0d4 AS xx

FROM --platform=$BUILDPLATFORM golang:1.22.5-bullseye@sha256:583d5af8289d30de50aa0dcf4985d8b8746e52622becd6e1a62cfe191d5275a5 AS builder

COPY --link --from=xx / /

ARG TARGETPLATFORM

RUN --mount=type=cache,id=${TARGETPLATFORM}-apt,target=/var/cache/apt,sharing=locked \
    apt-get update \
    && xx-apt-get install -y --no-install-recommends \
      gcc \
      libc6-dev

WORKDIR /build/orchestrator

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=/build,ro \
    xx-go mod download -x

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH
ARG BUILD_OPTS

ENV CGO_ENABLED=1

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=/build,ro \
    xx-go build ${BUILD_OPTS} -ldflags="-s -w -extldflags -static \
     -X 'github.com/openclarity/vmclarity/core/version.Version=${VERSION}' \
     -X 'github.com/openclarity/vmclarity/core/version.CommitHash=${COMMIT_HASH}' \
     -X 'github.com/openclarity/vmclarity/core/version.BuildTimestamp=${BUILD_TIMESTAMP}'" \
    -o /bin/vmclarity-orchestrator ./cmd/main.go

RUN xx-verify /bin/vmclarity-orchestrator

FROM alpine:3.20@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5

COPY --from=builder ["/bin/vmclarity-orchestrator", "/bin/vmclarity-orchestrator"]

RUN apk add --update --no-cache \
    ca-certificates \
    libc6-compat

ENTRYPOINT ["/bin/vmclarity-orchestrator"]
