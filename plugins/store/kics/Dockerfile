FROM --platform=$BUILDPLATFORM golang:1.22.4-alpine3.19@sha256:d9b1f00c269b4f82b8a28c87b4feeb34164305b29f6ce37f5c8cb076aac691c3 AS builder

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

FROM alpine:3.20@sha256:77726ef6b57ddf65bb551896826ec38bc3e53f75cdde31354fbffb4f25238ebd

COPY --from=builder /bin/vmclarity-scanner /bin/vmclarity-scanner
COPY --from=builder /artifacts/assets/queries /queries

USER 65534

ENTRYPOINT ["/bin/vmclarity-scanner"]
