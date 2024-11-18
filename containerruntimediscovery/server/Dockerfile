# syntax=docker/dockerfile:1@sha256:865e5dd094beca432e8c0a1d5e1c465db5f998dca4e439981029b3b81fb39ed5

# xx is a helper for cross-compilation
FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.4.0@sha256:0cd3f05c72d6c9b038eb135f91376ee1169ef3a330d34e418e65e2a5c2e9c0d4 AS xx

FROM --platform=$BUILDPLATFORM golang:1.23.2-bullseye@sha256:ecb3fe70e1fd6cef4c5c74246a7525c3b7d59c48ea0589bbb0e57b1b37321fb9 AS builder

COPY --from=xx / /

ARG TARGETPLATFORM

RUN --mount=type=cache,id=${TARGETPLATFORM}-apt,target=/var/cache/apt,sharing=locked \
    apt-get update \
    && xx-apt-get install -y --no-install-recommends \
      gcc \
      libc6-dev \
      libbtrfs-dev \
      libgpgme-dev \
      pkg-config \
      dpkg-dev

WORKDIR /build/containerruntimediscovery/server

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
        -o /bin/openclarity-cr-discovery-server cmd/main.go

RUN xx-verify /bin/openclarity-cr-discovery-server

FROM alpine:3.20@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a

COPY --from=builder ["/bin/openclarity-cr-discovery-server", "/bin/openclarity-cr-discovery-server"]

RUN apk add --update --no-cache \
    ca-certificates \
    libc6-compat

ENTRYPOINT ["/bin/openclarity-cr-discovery-server"]
