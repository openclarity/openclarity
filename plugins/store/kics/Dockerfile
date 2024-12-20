FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine3.19@sha256:5f3336882ad15d10ac1b59fbaba7cb84c35d4623774198b36ae60edeba45fd84 AS builder

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
    -o /bin/openclarity-scanner ./main.go

FROM alpine:3.21@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45

COPY --from=builder /bin/openclarity-scanner /bin/openclarity-scanner
COPY --from=builder /artifacts/assets/queries /queries

USER 65534

ENTRYPOINT ["/bin/openclarity-scanner"]
