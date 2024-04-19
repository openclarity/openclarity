## Image is built from root repo dir context.
## Example: docker build . -f plugins/sdk/go/example/Dockerfile -t go-scanner

FROM --platform=$BUILDPLATFORM golang:1.21.4-alpine3.18@sha256:70afe55365a265f0762257550bc38440e0d6d6b97020d3f8c85328f00200dd8e AS builder

ARG TARGETOS TARGETARCH

WORKDIR /build/plugins/sdk/go/example

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=.,target=/build,ro \
    go mod download -x

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,source=.,target=/build,ro \
    GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 \
    go build -ldflags="-s -w -extldflags -static" \
    -o /bin/go-scanner ./main.go

FROM alpine:3.18@sha256:11e21d7b981a59554b3f822c49f6e9f57b6068bb74f49c4cd5cc4c663c7e5160

COPY --from=builder /bin/go-scanner /bin/go-scanner

USER 65534

ENTRYPOINT ["/bin/go-scanner"]
