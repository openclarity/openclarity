# syntax=docker/dockerfile:1@sha256:fe40cf4e92cd0c467be2cfc30657a680ae2398318afd50b0c80585784c604f28

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

FROM nginxinc/nginx-unprivileged:1.27.0

COPY --from=site-build ["/src/ui/build", "/usr/share/nginx/html"]

COPY --link ["ui/nginx.conf", "/etc/nginx/conf.d/default.conf"]
