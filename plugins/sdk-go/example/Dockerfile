FROM --platform=$BUILDPLATFORM golang:1.23.4-alpine3.19@sha256:5f3336882ad15d10ac1b59fbaba7cb84c35d4623774198b36ae60edeba45fd84 AS builder

ARG TARGETOS TARGETARCH

WORKDIR /build/plugins/sdk-go/example

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=.,target=/build,ro \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=/build,ro \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags="-s -w -extldflags -static" \
    -o /bin/scanner ./main.go

FROM alpine:3.21@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099

COPY --from=builder /bin/scanner /bin/scanner

USER 65534

ENTRYPOINT ["/bin/scanner"]
