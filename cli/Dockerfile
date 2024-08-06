# syntax=docker/dockerfile:1@sha256:fe40cf4e92cd0c467be2cfc30657a680ae2398318afd50b0c80585784c604f28

ARG VMCLARITY_TOOLS_BASE=ghcr.io/openclarity/vmclarity-tools-base:v0.6.0@sha256:dc9779be25bdbd3a8725e768764a223b8ba5fd906fc5afc67b417886086806d8
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

WORKDIR /build/cli

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
    -o /bin/vmclarity-cli cmd/main.go

FROM ${VMCLARITY_TOOLS_BASE}

RUN apk upgrade --quiet --no-cache

RUN apk add --update --no-cache \
    util-linux \
    ca-certificates \
    libc6-compat

COPY --from=builder /bin/vmclarity-cli /bin/vmclarity-cli

ENTRYPOINT ["/bin/vmclarity-cli"]

CMD ["--help"]
