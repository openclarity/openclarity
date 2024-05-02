FROM --platform=$BUILDPLATFORM golang:1.21.4-alpine3.18@sha256:70afe55365a265f0762257550bc38440e0d6d6b97020d3f8c85328f00200dd8e AS builder

WORKDIR /artifacts

RUN <<EOT
  set -e

  url=https://github.com/Checkmarx/kics/releases/download/v1.7.13/extracted-info.zip

  archive="$(basename ${url})"

  wget -q -O "${archive}" "${url}"

  unzip "${archive}"
EOT

ARG TARGETOS TARGETARCH

WORKDIR /build/plugins/store/kics

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=.,target=/build,ro \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=/build,ro \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags="-s -w -extldflags -static" \
    -o /bin/vmclarity-scanner ./main.go

FROM alpine:3.19@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b

COPY --from=builder /bin/vmclarity-scanner /bin/vmclarity-scanner
COPY --from=builder /artifacts/assets/queries /queries

USER 65534

ENTRYPOINT ["/bin/vmclarity-scanner"]
