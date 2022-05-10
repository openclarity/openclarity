FROM golang:1.17.9-alpine AS builder

WORKDIR /build/
COPY main.go ./

COPY ./go.* ./
RUN go mod download

RUN go build -o test .

FROM alpine:3.15.4

WORKDIR /app

COPY --from=builder ["/build/test", "./test"]

ENTRYPOINT ["/app/test"]
