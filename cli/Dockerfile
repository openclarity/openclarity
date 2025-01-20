# syntax=docker/dockerfile:1@sha256:93bfd3b68c109427185cd78b4779fc82b484b0b7618e36d0f104d4d801e66d25

ARG VMCLARITY_TOOLS_BASE=ghcr.io/openclarity/vmclarity-tools-base:v0.6.0@sha256:dc9779be25bdbd3a8725e768764a223b8ba5fd906fc5afc67b417886086806d8
# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.4.0@sha256:0cd3f05c72d6c9b038eb135f91376ee1169ef3a330d34e418e65e2a5c2e9c0d4 AS xx

FROM --platform=$BUILDPLATFORM golang:1.23.4-bullseye@sha256:f2b2023be63d8b3f7d50a6af87e56ba432d361a3d4c389dbe701305c66cbc871 AS builder

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
     -X 'github.com/openclarity/openclarity/core/version.Version=${VERSION}' \
     -X 'github.com/openclarity/openclarity/core/version.CommitHash=${COMMIT_HASH}' \
     -X 'github.com/openclarity/openclarity/core/version.BuildTimestamp=${BUILD_TIMESTAMP}'" \
    -o /bin/openclarity-cli cmd/main.go

FROM ${VMCLARITY_TOOLS_BASE}

RUN apk upgrade --quiet --no-cache

RUN apk add --update --no-cache \
    util-linux \
    ca-certificates \
    libc6-compat

COPY --from=builder /bin/openclarity-cli /bin/openclarity-cli

ENTRYPOINT ["/bin/openclarity-cli"]

CMD ["--help"]
