FROM alpine:3.11.3

RUN apk add --no-cache bash

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

RUN mkdir -p /usr/local/portshift/

COPY ./kubei /usr/local/portshift/

RUN chmod +x /usr/local/portshift/kubei

WORKDIR /usr/local/portshift/


FROM golang:1.8-alpine as builder

FROM alpine:3.8

RUN apk add --no-cache ca-certificates

RUN mkdir -p /usr/local/portshift/

RUN touch /usr/local/portshift/db

RUN chmod +rw /usr/local/portshift/db

COPY ./view.html /usr/local/portshift/

RUN chmod +rw /usr/local/portshift/view.html

COPY ./kubei /usr/local/portshift/

RUN chmod +x /usr/local/portshift/kubei

EXPOSE 8080

ENTRYPOINT ["/usr/local/portshift/kubei"]
