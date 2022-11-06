FROM golang:1.19.3-alpine AS builder

WORKDIR /build
COPY go.* ./
RUN go mod download

ARG VERSION
ARG BUILD_TIMESTAMP
ARG COMMIT_HASH

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
