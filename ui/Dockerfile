# syntax=docker/dockerfile:1@sha256:9857836c9ee4268391bb5b09f9f157f3c91bb15821bb77969642813b0d00518d

FROM --platform=$BUILDPLATFORM node:20-slim AS site-build

WORKDIR /src

COPY ./ui /src/ui

COPY ./api/openapi.yaml /src/api/openapi.yaml

COPY ./uibackend/openapi.yaml /src/uibackend/openapi.yaml

RUN apt update

RUN apt install default-jre -y

RUN --mount=type=cache,target=/src/ui/node_modules \
    npm install --prefix /src/ui

RUN --mount=type=cache,target=/src/ui/node_modules \
    npm run build --prefix /src/ui

FROM nginxinc/nginx-unprivileged:1.27.3

COPY --from=site-build ["/src/ui/build", "/usr/share/nginx/html"]

COPY --link ["ui/nginx.conf", "/etc/nginx/conf.d/default.conf"]
