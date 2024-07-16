# syntax=docker/dockerfile:1@sha256:e87caa74dcb7d46cd820352bfea12591f3dba3ddc4285e19c7dcd13359f7cefd

FROM --platform=$BUILDPLATFORM node:20-slim AS site-build

WORKDIR /src

COPY --link ./ui /src

RUN --mount=type=cache,target=/src/node_modules \
    npm install

RUN --mount=type=cache,target=/src/node_modules \
    npm run build

FROM nginx:1.27.0@sha256:67682bda769fae1ccf5183192b8daf37b64cae99c6c3302650f6f8bf5f0f95df

COPY --from=site-build ["/src/build", "/usr/share/nginx/html"]

COPY --link ["ui/nginx.conf", "/etc/nginx/conf.d/default.conf"]
