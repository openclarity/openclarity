FROM golang:1.16-alpine AS builder
WORKDIR /go/src/kubei/
COPY ./ ./
RUN CGO_ENABLED=0 go build -o kubei ./cmd/kubei/

FROM alpine:3.11.3
RUN apk add --no-cache bash ca-certificates
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN mkdir /app
COPY pkg/webapp/view.html /app/
RUN chmod +rw /app/view.html
COPY --from=builder /go/src/kubei/kubei /app/
RUN chmod +x /app/kubei
ENTRYPOINT ["/app/kubei"]

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.name="kubei" \
    org.label-schema.description="Vulnerabilities scanning tool that allows users to get an accurate and immediate risk assessment of their kubernetes clusters" \
    org.label-schema.url="https://github.com/cisco-open/kubei" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-url="https://github.com/cisco-open/kubei"

USER 1000