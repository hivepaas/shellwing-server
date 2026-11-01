UID := $(shell id -u)

# ----- Development tools -----
init: build-devtools

DEVTOOLS_IMAGE := shellwing-devtools
DEVTOOLS_CMD := docker run --user "$(UID)" --rm --volume "$(PWD)":/app --network="host" $(DEVTOOLS_IMAGE)
build-devtools:
	@docker build --file ./tools/docker/Dockerfile --tag ${DEVTOOLS_IMAGE} .

GO_MOD_ENV=GOPRIVATE=github.com/hivepaas/*
mod:
	@$(GO_MOD_ENV) go mod tidy && go mod vendor && go mod verify

lint:
	$(DEVTOOLS_CMD) go run ./tools/goroutinelint .
	$(DEVTOOLS_CMD) go run ./tools/errcodelint
	$(DEVTOOLS_CMD) go run ./tools/openapi lint
	$(DEVTOOLS_CMD) golangci-lint --timeout=5m run -v ./...

lint-local:
	@go run ./tools/goroutinelint .
	@go run ./tools/errcodelint
	@go run ./tools/openapi lint
	# Run this cmd locally once to install golangci-lint binary
	# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.14.0
	# FASTER golangci-lint --timeout=5m run -v --new-from-rev=HEAD~1
	golangci-lint --timeout=5m run -v ./...

# The everyday one. Cached: a package nothing touched is not run again.
test:
	@go test ./...

# Before pushing anything that touches concurrency.
test-race:
	@go test -race ./...

# Race plus the coverage report. This is what CI runs.
test-cover:
	@./scripts/test.sh

nilaway:
	# run once: go install go.uber.org/nilaway/cmd/nilaway@latest
	@nilaway ./...

vuln:
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...

trivy:
	@if command -v trivy >/dev/null 2>&1; then \
		trivy fs --skip-dirs "vendor,.temp,tmp,temp,.appdata,web" --severity CRITICAL,HIGH .; \
	else \
		docker run --rm -v "$(PWD)":/app -w /app aquasec/trivy:latest fs --skip-dirs "vendor,.temp,tmp,temp,web" --severity CRITICAL,HIGH .; \
	fi

# ----- Build flags -----
PROD_LDFLAGS := -s -w
PROD_FLAGS := -trimpath -ldflags="$(PROD_LDFLAGS)"

run:
	@go run ./cmd/app/...

# Dev builds
build:
	@go build -o shellwing_server ./cmd/app/...

# Production builds (Stripped & Trimpath)
build-prod:
	@go build $(PROD_FLAGS) -o shellwing_server ./cmd/app/...

# ----- Code generation -----
gen: gen-go gen-swag

gen-go:
	$(DEVTOOLS_CMD) env GOCACHE=/tmp/go-cache go generate ./...

gen-swag:
	@./tools/swag/swag.sh

fmt:
	golangci-lint fmt ./...

# ----- DB migration -----
DB_MIGRATE_DIR := db
DB_CONN_STR := host=localhost port=47432 dbname=shellwing user=shellwing password=abc123
DB_MIGRATE_BASE := $(DEVTOOLS_CMD) sql-migrate
DB_MIGRATE_ENV := development
DB_EXEC_BASE := $(DEVTOOLS_CMD) psql -d "$(DB_CONN_STR)"

# This is considered the remote env
ifdef SHW_PLATFORM
ifneq ($(SHW_PLATFORM), local)
	DB_CONN_STR := host=${SHW_DB_HOST} port=${SHW_DB_PORT} dbname=${SHW_DB_DB_NAME} user=${SHW_DB_USER} password=${SHW_DB_PASSWORD}
	DB_MIGRATE_BASE := sql-migrate
	DB_MIGRATE_ENV := main
	DB_EXEC_BASE := psql -d "${DB_CONN_STR}"
endif
endif

migrate-setup: build-devtools

migrate-new:
ifndef NAME
	$(error "Please provide migration name, i.e.: make $@ NAME=example_migration")
else
	$(DB_MIGRATE_BASE) new -config=${DB_MIGRATE_DIR}/dbconfig.yml $(NAME)
endif

migrate-status:
	$(DB_MIGRATE_BASE) status -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

migrate-up:
	$(DB_MIGRATE_BASE) up -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

migrate-down:
	$(DB_MIGRATE_BASE) down -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

migrate-redo:
	$(DB_MIGRATE_BASE) redo -config=${DB_MIGRATE_DIR}/dbconfig.yml -env=$(DB_MIGRATE_ENV)

seed-data:
	make migrate-up
	$(DB_EXEC_BASE) -f ${DB_MIGRATE_DIR}/seed/seed.sql

seed-data-with-clear:
	$(DB_EXEC_BASE) -f ${DB_MIGRATE_DIR}/seed/clear.sql
	make migrate-up
	$(DB_EXEC_BASE) -f ${DB_MIGRATE_DIR}/seed/seed.sql

# The web app (web/), built into dist-dashboard, which the server serves.
local-build-fe:
	cd web && yarn install --frozen-lockfile && yarn build
	rm -rf dist-dashboard
	mv web/dist dist-dashboard

# The release's image, built locally from the working tree as it is. The release
# workflow builds it from a tag (.github/workflows/release.yml).
docker-build:
	docker buildx build --load -t shellwing-server:dev .
