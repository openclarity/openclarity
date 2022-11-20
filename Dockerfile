FROM golang:1.19.3-alpine AS builder

WORKDIR /build
COPY go.* ./
RUN go mod download

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH

# Copy runtime_scan go.mod & go.sum
WORKDIR /build/runtime_scan
COPY runtime_scan/go.* ./
RUN go mod download

# Copy runtime_scan code
WORKDIR /build
COPY runtime_scan ./runtime_scan

# Copy and build backend code
COPY . .
RUN go build -ldflags="-s -w \
     -X 'github.com/openclarity/vmclarity/pkg/version.Version=${VERSION}' \
     -X 'github.com/openclarity/vmclarity/pkg/version.CommitHash=${COMMIT_HASH}' \
     -X 'github.com/openclarity/vmclarity/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}'" -o vmclarity ./main.go

FROM alpine:3.16

WORKDIR /app

COPY --from=builder ["/build/vmclarity", "./vmclarity"]

ENTRYPOINT ["/app/vmclarity"]
