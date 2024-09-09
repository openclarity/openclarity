FROM --platform=$BUILDPLATFORM golang:1.23.0-alpine3.19@sha256:fe8f9c7d418d3ac91787f11c31071c4814b6da5f9aae55bc581a7aacc264c395 AS builder

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

FROM alpine:3.20@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d

COPY --from=builder /bin/scanner /bin/scanner

USER 65534

ENTRYPOINT ["/bin/scanner"]
