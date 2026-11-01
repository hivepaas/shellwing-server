# The image of a release: the server, the web app (web/) of the same
# commit, and the migrations it brings its database to at start.
#
#   docker buildx build -t shellwing-server:dev .
#
# It follows HivePaaS's deployment/release/Dockerfile, with these differences:
# - the web app is in this repository, under web/, not a repository of its own;
# - the server runs as a user of its own (uid 10001), and owns nothing in the
#   image but its data directory, /var/lib/shellwing, which is a volume;
# - it migrates its database when it starts (deployment/release/entrypoint.sh);
# - no key is in it: .dockerignore keeps the development keys and every *.key
#   out of the build, and of the configuration only config.production.toml is
#   copied. The production signing keys are docker secrets of the service.
#
# docs/deploying.md says how it runs; .github/workflows/release.yml builds it.

############################################################
# Build the web app
FROM node:26-alpine AS web_builder
# A checkout's git hooks have no place in an image's build.
ENV HUSKY=0
RUN apk update && apk upgrade --no-cache && apk add --update --no-cache yarn

WORKDIR /web

# The dependencies first, from the lockfile, so that a change of the sources
# alone does not install them again.
COPY web/package.json web/yarn.lock ./
RUN --mount=type=cache,target=/tmp/yarn-cache \
    yarn config set network-timeout 300000 && \
    yarn config set cache-folder /tmp/yarn-cache && \
    yarn install --frozen-lockfile

# A developer's checkout may hold its own build, modules and *.local env files
# (git ignores them): only what a clean checkout holds is built.
COPY --exclude=node_modules --exclude=dist --exclude=**/*.local web/ .
RUN yarn build

############################################################
# Build the server
FROM golang:1.27-alpine AS builder

RUN apk update && \
    apk add --no-cache -U make git ca-certificates && \
    rm -rf /var/cache/apk/*

ENV CGO_ENABLED=0

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# The web app is the stage above's: without it here, a change to it alone does
# not build the server again.
COPY --exclude=web . .
RUN make build-prod

# The database migration tool the entrypoint runs, pinned as HivePaaS pins it:
# built in a module of its own that raises golang.org/x/crypto past the one its
# release requires, which the image scan before a release refuses.
ARG SQL_MIGRATE_VERSION=v1.8.1
ARG SQL_MIGRATE_CRYPTO_VERSION=v0.57.0
RUN mkdir /tmp/sql-migrate && cd /tmp/sql-migrate && \
    go mod init sql-migrate-build && \
    go get "github.com/rubenv/sql-migrate/sql-migrate@${SQL_MIGRATE_VERSION}" \
        "golang.org/x/crypto@${SQL_MIGRATE_CRYPTO_VERSION}" && \
    go build -trimpath -o /build/sql-migrate github.com/rubenv/sql-migrate/sql-migrate && \
    rm -rf /tmp/sql-migrate

############################################################
# The image
FROM alpine:3.24
ARG VERSION
ARG REVISION

LABEL org.opencontainers.image.title="shellwing server" \
      org.opencontainers.image.source="https://github.com/hivepaas/shellwing-server" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

# ca-certificates: the mail server and S3 are reached over TLS. The
# health check's wget is busybox's.
RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*

# The server's user, and its data directory: shellwing.toml, which the server
# rewrites when it rotates a secret, and the directories it makes at start.
# Give the service a volume there (docs/deploying.md).
RUN addgroup -S -g 10001 shellwing && \
    adduser -S -D -H -u 10001 -G shellwing -h /var/lib/shellwing -s /sbin/nologin shellwing && \
    install -d -o shellwing -g shellwing -m 0700 /var/lib/shellwing

WORKDIR /shellwing

COPY --from=builder /build/shellwing_server ./
COPY --from=builder /build/db/dbconfig.yml ./db/
COPY --from=builder /build/db/migrations ./db/migrations
COPY --from=builder /build/config/config.production.toml ./config/
COPY --from=builder /build/sql-migrate /usr/local/bin/sql-migrate
COPY --from=builder --chmod=0755 /build/deployment/release/entrypoint.sh /usr/local/bin/shellwing-entrypoint
COPY --from=web_builder /web/dist ./dist-dashboard

# A config.toml in the data directory would win over it.
ENV SHW_CONFIG_FILE=config/config.production.toml

USER 10001:10001
EXPOSE 10100

HEALTHCHECK --start-period=60s --start-interval=5s --interval=30s --timeout=5s --retries=3 \
    CMD wget -q --spider "http://127.0.0.1:${SHW_HTTP_SERVER_PORT:-10100}/api/ping" || exit 1

ENTRYPOINT ["/usr/local/bin/shellwing-entrypoint"]
