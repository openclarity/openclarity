# syntax=docker/dockerfile:1@sha256:4c68376a702446fc3c79af22de146a148bc3367e73c25a5803d453b6b3f722fb

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
